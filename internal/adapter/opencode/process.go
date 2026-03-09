package opencode

import (
	"context"
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sst/opencode-sdk-go"
	"github.com/sst/opencode-sdk-go/option"
)

// ---------------------------------------------------------------------------
// Process — manages a single OpenCode server process
// ---------------------------------------------------------------------------

// ProcessOpts configures how a managed OpenCode server is started.
type ProcessOpts struct {
	Binary       string
	Port         int
	Hostname     string
	Directory    string
	ConfigFile   string
	ConfigDir    string
	Env          []string
	Password     string
	Username     string
	StartTimeout time.Duration
	ClientOpts   []option.RequestOption
}

// Process manages a single OpenCode server process lifecycle.
type Process struct {
	mu      sync.Mutex
	cmd     *exec.Cmd
	client  *opencode.Client
	baseURL string
	opts    ProcessOpts
	port    int
	done    chan struct{}
}

func (p *Process) Client() *opencode.Client { return p.client }
func (p *Process) Port() int                { return p.port }
func (p *Process) BaseURL() string          { return p.baseURL }
func (p *Process) Done() <-chan struct{}    { return p.done }

func (p *Process) Running() bool {
	select {
	case <-p.done:
		return false
	default:
		return true
	}
}

// StartServer starts a new OpenCode server process and waits for it to
// become healthy.
func StartServer(ctx context.Context, opts ProcessOpts) (*Process, error) {
	binary := opts.Binary
	if binary == "" {
		binary = "opencode"
	}

	binPath, err := exec.LookPath(binary)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrNotInstalled, binary)
	}

	port := opts.Port
	if port == 0 {
		port, err = freePort()
		if err != nil {
			return nil, fmt.Errorf("find free port: %w", err)
		}
	}

	hostname := opts.Hostname
	if hostname == "" {
		hostname = "127.0.0.1"
	}

	timeout := opts.StartTimeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	args := []string{"serve", "--port", strconv.Itoa(port), "--hostname", hostname}
	cmd := exec.Command(binPath, args...)
	cmd.Dir = opts.Directory
	cmd.Env = buildEnv(opts)
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start opencode serve: %w", err)
	}

	done := make(chan struct{})
	go func() {
		_ = cmd.Wait()
		close(done)
	}()

	baseURL := fmt.Sprintf("http://%s:%d", hostname, port)
	clientOpts := buildClientOpts(baseURL, opts)
	client := opencode.NewClient(clientOpts...)

	proc := &Process{
		cmd:     cmd,
		client:  client,
		baseURL: baseURL,
		opts:    opts,
		port:    port,
		done:    done,
	}

	if err := proc.waitHealthy(ctx, timeout); err != nil {
		_ = proc.kill()
		return nil, err
	}

	return proc, nil
}

func (p *Process) Stop(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	select {
	case <-p.done:
		return nil
	default:
	}

	// Attempt graceful dispose via the SDK escape hatch.
	disposeCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	_ = p.client.Post(disposeCtx, "/instance/dispose", nil, nil)

	select {
	case <-p.done:
		return nil
	case <-time.After(2 * time.Second):
	}

	if err := interrupt(p.cmd.Process); err != nil {
		return p.kill()
	}

	select {
	case <-p.done:
		return nil
	case <-time.After(5 * time.Second):
		return p.kill()
	}
}

func (p *Process) kill() error {
	if p.cmd.Process == nil {
		return nil
	}
	err := p.cmd.Process.Kill()
	<-p.done
	return err
}

// healthResponse is the shape of GET /global/health (not in the SDK).
type healthResponse struct {
	Healthy bool   `json:"healthy"`
	Version string `json:"version"`
}

func (p *Process) waitHealthy(ctx context.Context, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	pollCtx, cancel := context.WithDeadline(ctx, deadline)
	defer cancel()

	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-p.done:
			return fmt.Errorf("%w: process exited before becoming healthy", ErrStartTimeout)
		case <-pollCtx.Done():
			return fmt.Errorf("%w: waited %v", ErrStartTimeout, timeout)
		case <-ticker.C:
			healthCtx, hcancel := context.WithTimeout(pollCtx, 2*time.Second)
			var h healthResponse
			err := p.client.Get(healthCtx, "/global/health", nil, &h)
			hcancel()
			if err == nil && h.Healthy {
				return nil
			}
		}
	}
}

// ---------------------------------------------------------------------------
// Standalone helpers
// ---------------------------------------------------------------------------

func IsInstalled(binary string) bool {
	if binary == "" {
		binary = "opencode"
	}
	_, err := exec.LookPath(binary)
	return err == nil
}

func BinaryPath(binary string) (string, error) {
	if binary == "" {
		binary = "opencode"
	}
	path, err := exec.LookPath(binary)
	if err != nil {
		return "", fmt.Errorf("%w: %s", ErrNotInstalled, binary)
	}
	return path, nil
}

func Version(ctx context.Context, binary string) (string, error) {
	if binary == "" {
		binary = "opencode"
	}
	binPath, err := exec.LookPath(binary)
	if err != nil {
		return "", fmt.Errorf("%w: %s", ErrNotInstalled, binary)
	}

	cmd := exec.CommandContext(ctx, binPath, "--version")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("opencode version: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// ---------------------------------------------------------------------------
// Discovery
// ---------------------------------------------------------------------------

// DiscoverOpts configures server discovery.
type DiscoverOpts struct {
	Binary       string
	ProbeTimeout time.Duration
	ClientOpts   []option.RequestOption
}

// discoveredServer represents an OpenCode server found via process-table scanning.
type discoveredServer struct {
	PID       int
	Hostname  string
	Port      int
	Version   string
	Directory string
	Healthy   bool
}

type serverCandidate struct {
	pid      int
	hostname string
	port     int
}

func DiscoverServers(ctx context.Context, opts DiscoverOpts) ([]discoveredServer, error) {
	binary := opts.Binary
	if binary == "" {
		binary = "opencode"
	}
	probeTimeout := opts.ProbeTimeout
	if probeTimeout == 0 {
		probeTimeout = 2 * time.Second
	}

	psCmd := exec.CommandContext(ctx, "ps", "-eo", "pid,args")
	out, err := psCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("run ps: %w", err)
	}

	candidates := parseProcessList(string(out), binary)
	if len(candidates) == 0 {
		return nil, nil
	}

	type result struct {
		server discoveredServer
		ok     bool
	}

	results := make([]result, len(candidates))
	var wg sync.WaitGroup
	for i, c := range candidates {
		wg.Add(1)
		go func(idx int, cand serverCandidate) {
			defer wg.Done()
			srv := probeCandidate(ctx, cand, probeTimeout, opts.ClientOpts)
			results[idx] = result{server: srv, ok: srv.Healthy}
		}(i, c)
	}
	wg.Wait()

	var servers []discoveredServer
	for _, r := range results {
		if r.ok {
			servers = append(servers, r.server)
		}
	}
	return servers, nil
}

func parseProcessList(output string, binary string) []serverCandidate {
	var candidates []serverCandidate

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		pid, err := strconv.Atoi(fields[0])
		if err != nil {
			continue
		}

		args := fields[1:]
		cmd := args[0]

		cmdBase := cmd
		if idx := strings.LastIndex(cmd, "/"); idx >= 0 {
			cmdBase = cmd[idx+1:]
		}
		if cmdBase != binary {
			continue
		}

		if len(args) > 1 && args[1] == "run" {
			continue
		}

		port, hostname := parseServerArgs(args[1:])
		if port == 0 {
			continue
		}

		candidates = append(candidates, serverCandidate{
			pid:      pid,
			hostname: hostname,
			port:     port,
		})
	}
	return candidates
}

func parseServerArgs(args []string) (port int, hostname string) {
	hostname = "127.0.0.1"

	for i := 0; i < len(args); i++ {
		arg := args[i]

		switch {
		case arg == "--port" && i+1 < len(args):
			if p, err := strconv.Atoi(args[i+1]); err == nil {
				port = p
			}
			i++
		case strings.HasPrefix(arg, "--port="):
			if p, err := strconv.Atoi(strings.TrimPrefix(arg, "--port=")); err == nil {
				port = p
			}
		case arg == "--hostname" && i+1 < len(args):
			hostname = args[i+1]
			i++
		case strings.HasPrefix(arg, "--hostname="):
			hostname = strings.TrimPrefix(arg, "--hostname=")
		}
	}
	return
}

func probeCandidate(ctx context.Context, cand serverCandidate, timeout time.Duration, clientOpts []option.RequestOption) discoveredServer {
	srv := discoveredServer{
		PID:      cand.pid,
		Hostname: cand.hostname,
		Port:     cand.port,
	}

	baseURL := fmt.Sprintf("http://%s:%d", cand.hostname, cand.port)
	probeOpts := make([]option.RequestOption, 0, len(clientOpts)+2)
	probeOpts = append(probeOpts, option.WithBaseURL(baseURL))
	probeOpts = append(probeOpts, option.WithRequestTimeout(timeout))
	probeOpts = append(probeOpts, clientOpts...)
	client := opencode.NewClient(probeOpts...)

	healthCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	var h healthResponse
	err := client.Get(healthCtx, "/global/health", nil, &h)
	if err != nil || !h.Healthy {
		return srv
	}
	srv.Healthy = true
	srv.Version = h.Version

	pathCtx, pathCancel := context.WithTimeout(ctx, timeout)
	defer pathCancel()
	pathInfo, err := client.Path.Get(pathCtx, opencode.PathGetParams{})
	if err == nil {
		srv.Directory = pathInfo.Directory
	}

	return srv
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

// buildClientOpts constructs SDK request options for a server process,
// including base URL and optional basic auth middleware.
func buildClientOpts(baseURL string, opts ProcessOpts) []option.RequestOption {
	reqOpts := make([]option.RequestOption, 0, len(opts.ClientOpts)+2)
	reqOpts = append(reqOpts, option.WithBaseURL(baseURL))

	if opts.Password != "" {
		username := opts.Username
		if username == "" {
			username = "opencode"
		}
		reqOpts = append(reqOpts, basicAuthMiddleware(username, opts.Password))
	}

	reqOpts = append(reqOpts, opts.ClientOpts...)
	return reqOpts
}

// basicAuthMiddleware returns a RequestOption that injects an HTTP Basic
// Authorization header into every request. The SDK has no built-in auth
// support, so we use the middleware escape hatch.
func basicAuthMiddleware(username, password string) option.RequestOption {
	creds := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
	return option.WithMiddleware(func(req *http.Request, next option.MiddlewareNext) (*http.Response, error) {
		req.Header.Set("Authorization", "Basic "+creds)
		return next(req)
	})
}

func buildEnv(opts ProcessOpts) []string {
	env := os.Environ()

	if opts.Password != "" {
		env = append(env, "OPENCODE_SERVER_PASSWORD="+opts.Password)
		username := opts.Username
		if username == "" {
			username = "opencode"
		}
		env = append(env, "OPENCODE_SERVER_USERNAME="+username)
	}

	if opts.ConfigFile != "" {
		env = append(env, "OPENCODE_CONFIG="+opts.ConfigFile)
	}
	if opts.ConfigDir != "" {
		env = append(env, "OPENCODE_CONFIG_DIR="+opts.ConfigDir)
	}

	env = append(env, opts.Env...)
	return env
}

func freePort() (int, error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	port := l.Addr().(*net.TCPAddr).Port
	l.Close()
	return port, nil
}

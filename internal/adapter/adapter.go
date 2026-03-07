package adapter

import (
	"fmt"
	"os"
	"syscall"
)

// Adapter defines an external tool that can be launched in a directory.
type Adapter struct {
	Name    string   `yaml:"name"`
	Command string   `yaml:"command"`
	Args    []string `yaml:"args,omitempty"`
	Default bool     `yaml:"default,omitempty"`
}

// Registry holds all configured adapters.
type Registry struct {
	adapters map[string]*Adapter
}

// NewRegistry creates a registry from a slice of adapters.
func NewRegistry(adapters []*Adapter) *Registry {
	r := &Registry{adapters: make(map[string]*Adapter)}
	for _, a := range adapters {
		r.adapters[a.Name] = a
	}
	return r
}

// Get returns an adapter by name.
func (r *Registry) Get(name string) (*Adapter, error) {
	a, ok := r.adapters[name]
	if !ok {
		return nil, fmt.Errorf("adapter %q not found", name)
	}
	return a, nil
}

// Default returns the adapter marked as default.
func (r *Registry) Default() (*Adapter, error) {
	for _, a := range r.adapters {
		if a.Default {
			return a, nil
		}
	}
	return nil, fmt.Errorf("no default adapter configured")
}

// All returns all adapters.
func (r *Registry) All() []*Adapter {
	result := make([]*Adapter, 0, len(r.adapters))
	for _, a := range r.adapters {
		result = append(result, a)
	}
	return result
}

// Exec replaces the current process with the adapter command, launched
// in the given directory. This does not return on success.
func (a *Adapter) Exec(dir string) error {
	binary, err := lookPath(a.Command)
	if err != nil {
		return fmt.Errorf("adapter %q: command %q not found in PATH", a.Name, a.Command)
	}

	// Build argv: command name + configured args + directory
	argv := []string{a.Command}
	argv = append(argv, a.Args...)

	env := os.Environ()

	// Change to the target directory before exec
	if err := os.Chdir(dir); err != nil {
		return fmt.Errorf("cannot change to directory %s: %w", dir, err)
	}

	return syscall.Exec(binary, argv, env)
}

// lookPath finds the absolute path of a command.
func lookPath(cmd string) (string, error) {
	// Use os/exec.LookPath equivalent via PATH search
	path := os.Getenv("PATH")
	for _, dir := range splitPath(path) {
		candidate := dir + "/" + cmd
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("command %q not found", cmd)
}

func splitPath(path string) []string {
	if path == "" {
		return nil
	}
	result := []string{}
	start := 0
	for i := 0; i < len(path); i++ {
		if path[i] == ':' {
			if i > start {
				result = append(result, path[start:i])
			}
			start = i + 1
		}
	}
	if start < len(path) {
		result = append(result, path[start:])
	}
	return result
}

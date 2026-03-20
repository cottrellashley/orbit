package server

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
)

// handleProxy routes requests by port (legacy route: /api/proxy/{port}/{path...}).
// Kept for backward compatibility — new clients should use handleNodeProxy.
func (s *Server) handleProxy(w http.ResponseWriter, r *http.Request) {
	portStr := r.PathValue("port")
	port, err := strconv.Atoi(portStr)
	if err != nil || port < 1 || port > 65535 {
		writeError(w, http.StatusBadRequest, "invalid port")
		return
	}

	s.reverseProxy(w, r, "127.0.0.1", port)
}

// handleNodeProxy routes requests by nodeID (new route: /api/nodes/{id}/proxy/{path...}).
// Looks up hostname:port from the node registry, supporting remote nodes.
func (s *Server) handleNodeProxy(w http.ResponseWriter, r *http.Request) {
	if s.nodeSvc == nil {
		writeError(w, http.StatusNotImplemented, "node service not configured")
		return
	}
	nodeID := r.PathValue("id")
	node, err := s.nodeSvc.GetNode(r.Context(), nodeID)
	if err != nil {
		writeError(w, http.StatusNotFound, fmt.Sprintf("node %q not found", nodeID))
		return
	}
	if !node.Healthy {
		writeError(w, http.StatusServiceUnavailable, fmt.Sprintf("node %q is unhealthy", nodeID))
		return
	}

	s.reverseProxy(w, r, node.Hostname, node.Port)
}

// reverseProxy forwards the request to the given hostname:port.
func (s *Server) reverseProxy(w http.ResponseWriter, r *http.Request, hostname string, port int) {
	targetPath := "/" + r.PathValue("path")
	if r.URL.RawQuery != "" {
		targetPath += "?" + r.URL.RawQuery
	}

	target, _ := url.Parse(fmt.Sprintf("http://%s:%d", hostname, port))
	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = target.Scheme
			req.URL.Host = target.Host
			req.URL.Path = targetPath
			req.URL.RawQuery = r.URL.RawQuery
			req.Host = target.Host
			// Forward basic auth if the server uses it
			if user, pass, ok := r.BasicAuth(); ok {
				req.SetBasicAuth(user, pass)
			}
		},
		ModifyResponse: func(resp *http.Response) error {
			// If this is an SSE stream, ensure proper headers
			if strings.Contains(resp.Header.Get("Content-Type"), "text/event-stream") {
				resp.Header.Set("Cache-Control", "no-cache")
				resp.Header.Set("Connection", "keep-alive")
			}
			return nil
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			writeError(w, http.StatusBadGateway, fmt.Sprintf("proxy error: %v", err))
		},
	}

	proxy.ServeHTTP(w, r)
}

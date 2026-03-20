package server

import (
	"encoding/json"
	"log"
	"net/http"

	"nhooyr.io/websocket"

	"github.com/cottrellashley/orbit/internal/domain"
)

// ---------------------------------------------------------------------------
// Terminal API handlers
// ---------------------------------------------------------------------------

// GET /api/managed-server — returns the URL of the managed OpenCode server.
func (s *Server) handleManagedServer(w http.ResponseWriter, r *http.Request) {
	if s.managedServerURL == "" {
		writeError(w, http.StatusServiceUnavailable, "no managed server running")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"url": s.managedServerURL})
}

// nil-guard helper for all terminal handlers.
func (s *Server) terminalReady(w http.ResponseWriter) bool {
	if s.termSvc == nil {
		writeError(w, http.StatusNotImplemented, "terminal service not configured")
		return false
	}
	return true
}

// spawnTerminalRequest is the JSON body for POST /api/terminal.
type spawnTerminalRequest struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
	Cols    uint16   `json:"cols"`
	Rows    uint16   `json:"rows"`
}

// terminalJSON is the JSON response for a terminal.
type terminalJSON struct {
	ID      string `json:"id"`
	Command string `json:"command"`
	Running bool   `json:"running"`
}

func toTerminalJSON(t domain.Terminal) terminalJSON {
	return terminalJSON{
		ID:      t.ID,
		Command: t.Command,
		Running: t.Running,
	}
}

// POST /api/terminal — spawn a new PTY-backed terminal.
func (s *Server) handleSpawnTerminal(w http.ResponseWriter, r *http.Request) {
	if !s.terminalReady(w) {
		return
	}

	var req spawnTerminalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.Command == "" {
		writeError(w, http.StatusBadRequest, "command is required")
		return
	}
	if req.Cols == 0 {
		req.Cols = 120
	}
	if req.Rows == 0 {
		req.Rows = 40
	}

	t, err := s.termSvc.Spawn(r.Context(), domain.TerminalSpawnOpts{
		Command: req.Command,
		Args:    req.Args,
		Cols:    req.Cols,
		Rows:    req.Rows,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, toTerminalJSON(*t))
}

// GET /api/terminal — list active terminals.
func (s *Server) handleListTerminals(w http.ResponseWriter, r *http.Request) {
	if !s.terminalReady(w) {
		return
	}

	terminals := s.termSvc.List(r.Context())
	out := make([]terminalJSON, len(terminals))
	for i, t := range terminals {
		out[i] = toTerminalJSON(t)
	}
	writeJSON(w, http.StatusOK, out)
}

// DELETE /api/terminal/{id} — kill a terminal.
func (s *Server) handleKillTerminal(w http.ResponseWriter, r *http.Request) {
	if !s.terminalReady(w) {
		return
	}

	id := r.PathValue("id")
	if err := s.termSvc.Kill(r.Context(), id); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// GET /api/terminal/{id}/ws — WebSocket for bidirectional PTY I/O.
//
// Protocol:
//   - Binary frames: raw PTY I/O (both directions)
//   - Text frames: JSON control messages, e.g. {"type":"resize","cols":120,"rows":40}
//   - Server → client: binary frames with PTY output
//   - Client → server: binary frames with keystrokes, text frames with control messages
func (s *Server) handleTerminalWS(w http.ResponseWriter, r *http.Request) {
	if !s.terminalReady(w) {
		return
	}

	id := r.PathValue("id")

	// Attach to the terminal's PTY before upgrading — fail fast if not found.
	conn, err := s.termSvc.Attach(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	defer conn.Close()

	// Upgrade to WebSocket.
	ws, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		// Allow all origins — Orbit is a local dev tool.
		InsecureSkipVerify: true,
	})
	if err != nil {
		log.Printf("terminal ws upgrade: %v", err)
		return
	}
	defer ws.Close(websocket.StatusNormalClosure, "")

	ctx := r.Context()

	// PTY → WebSocket (binary frames).
	go func() {
		buf := make([]byte, 32*1024)
		for {
			n, err := conn.Read(buf)
			if err != nil {
				// PTY closed — close the WebSocket.
				ws.Close(websocket.StatusNormalClosure, "process exited")
				return
			}
			if err := ws.Write(ctx, websocket.MessageBinary, buf[:n]); err != nil {
				return
			}
		}
	}()

	// WebSocket → PTY.
	for {
		msgType, data, err := ws.Read(ctx)
		if err != nil {
			// WebSocket closed by client.
			return
		}

		switch msgType {
		case websocket.MessageBinary:
			// Raw keystrokes → PTY input.
			if _, err := conn.Write(data); err != nil {
				return
			}

		case websocket.MessageText:
			// JSON control message.
			var ctrl struct {
				Type string `json:"type"`
				Cols uint16 `json:"cols"`
				Rows uint16 `json:"rows"`
			}
			if err := json.Unmarshal(data, &ctrl); err != nil {
				continue
			}
			switch ctrl.Type {
			case "resize":
				if ctrl.Cols > 0 && ctrl.Rows > 0 {
					_ = conn.Resize(ctrl.Cols, ctrl.Rows)
				}
			}
		}
	}
}

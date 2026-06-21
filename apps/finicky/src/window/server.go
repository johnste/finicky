package window

import (
	"encoding/json"
	"finicky/browser"
	"finicky/rules"
	"finicky/util"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"sync"
)

var (
	apiPort           int
	hub               = newSSEHub()
	GetVersionFunc    func() string
	GetConfigFunc     func() interface{}
	GetUpdateInfoFunc func() interface{}
	TestURLFunc       func(url string) (interface{}, error)
)

type sseHub struct {
	mu      sync.Mutex
	clients map[chan string]struct{}
}

func newSSEHub() *sseHub {
	return &sseHub{clients: make(map[chan string]struct{})}
}

func (h *sseHub) register(ch chan string) {
	h.mu.Lock()
	h.clients[ch] = struct{}{}
	h.mu.Unlock()
}

func (h *sseHub) unregister(ch chan string) {
	h.mu.Lock()
	delete(h.clients, ch)
	close(ch)
	h.mu.Unlock()
}

func (h *sseHub) broadcast(event string, data interface{}) {
	payload, err := json.Marshal(data)
	if err != nil {
		slog.Error("Failed to marshal SSE event", "event", event, "error", err)
		return
	}
	h.broadcastRaw(event, payload)
}

func (h *sseHub) broadcastRaw(event string, payload []byte) {
	msg := fmt.Sprintf("event: %s\ndata: %s\n\n", event, payload)
	h.mu.Lock()
	for ch := range h.clients {
		select {
		case ch <- msg:
		default:
		}
	}
	h.mu.Unlock()
}

// BroadcastSSE sends a named SSE event to all connected UI clients.
func BroadcastSSE(event string, data interface{}) {
	hub.broadcast(event, data)
}

// BroadcastSSERaw sends a named SSE event with pre-encoded JSON (e.g. from slog).
func BroadcastSSERaw(event string, rawJSON []byte) {
	hub.broadcastRaw(event, rawJSON)
}

// StartAPIServer binds to a random local port and serves the REST + SSE API.
func StartAPIServer() error {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return fmt.Errorf("failed to start API server: %w", err)
	}
	apiPort = ln.Addr().(*net.TCPAddr).Port
	slog.Debug("API server listening", "url", fmt.Sprintf("http://127.0.0.1:%d/api", apiPort))

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/initial-data", handleInitialData)
	mux.HandleFunc("GET /api/rules", handleGetRulesHTTP)
	mux.HandleFunc("POST /api/rules", handleSaveRulesHTTP)
	mux.HandleFunc("GET /api/browsers", handleGetBrowsersHTTP)
	mux.HandleFunc("GET /api/browser-profiles", handleGetBrowserProfilesHTTP)
	mux.HandleFunc("POST /api/test-url", handleTestURLHTTP)
	mux.HandleFunc("GET /api/events", handleSSE)

	go http.Serve(ln, corsMiddleware(mux)) //nolint:errcheck
	return nil
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("Failed to write JSON response", "error", err)
	}
}

func handleInitialData(w http.ResponseWriter, _ *http.Request) {
	result := map[string]interface{}{
		"rules":             getRulesData(),
		"installedBrowsers": browser.GetInstalledBrowsers(),
	}
	if GetVersionFunc != nil {
		result["version"] = GetVersionFunc()
	}
	if GetConfigFunc != nil {
		result["config"] = GetConfigFunc()
	}
	if GetUpdateInfoFunc != nil {
		if ui := GetUpdateInfoFunc(); ui != nil {
			result["updateInfo"] = ui
		}
	}
	writeJSON(w, result)
}

func handleGetRulesHTTP(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, getRulesData())
}

func handleSaveRulesHTTP(w http.ResponseWriter, r *http.Request) {
	var rf rules.RulesFile
	if err := json.NewDecoder(r.Body).Decode(&rf); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := rules.Save(rf); err != nil {
		slog.Error("Failed to save rules", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	slog.Debug("Rules saved", "count", len(rf.Rules))
	if SaveRulesHandler != nil {
		go SaveRulesHandler(rf)
	}
	writeJSON(w, getRulesData())
}

func handleGetBrowsersHTTP(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, browser.GetInstalledBrowsers())
}

func handleGetBrowserProfilesHTTP(w http.ResponseWriter, r *http.Request) {
	browserName := r.URL.Query().Get("browser")
	writeJSON(w, browser.GetProfilesForBrowser(browserName))
}

func handleTestURLHTTP(w http.ResponseWriter, r *http.Request) {
	var body struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if TestURLFunc == nil {
		http.Error(w, "test URL handler not initialized", http.StatusServiceUnavailable)
		return
	}
	result, err := TestURLFunc(body.URL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(w, result)
}

func handleSSE(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ch := make(chan string, 32)
	hub.register(ch)
	defer hub.unregister(ch)

	for {
		select {
		case msg, ok := <-ch:
			if !ok {
				return
			}
			fmt.Fprint(w, msg)
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}

func getRulesData() interface{} {
	rf, err := rules.Load()
	if err != nil {
		slog.Error("Failed to load rules", "error", err)
		return map[string]interface{}{"defaultBrowser": "", "rules": []interface{}{}}
	}
	path, _ := rules.GetPath()
	var rulesPath string
	if _, statErr := os.Stat(path); statErr == nil {
		rulesPath = path
	}
	type rulesResponse struct {
		rules.RulesFile
		Path string `json:"path,omitempty"`
	}
	return rulesResponse{RulesFile: rf, Path: util.ShortenPath(rulesPath)}
}

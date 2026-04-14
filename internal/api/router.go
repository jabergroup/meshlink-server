package api

import (
	"meshlink-server/internal/store"
	"net/http"
)

// corsMiddleware adds CORS headers for the web dashboard
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// NewRouter creates the HTTP router with all API routes
func NewRouter(s *store.Store) http.Handler {
	h := NewHandlers(s)

	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("/api/pair/create", h.CreatePair)
	mux.HandleFunc("/api/pair/join", h.JoinPair)
	mux.HandleFunc("/api/pair/status", h.GetPairStatus)
	mux.HandleFunc("/api/pair/update", h.UpdateStatus)
	mux.HandleFunc("/api/heartbeat", h.Heartbeat)
	mux.HandleFunc("/api/sessions", h.ListSessions)

	// Agent download endpoint
	mux.HandleFunc("/download/agent", ServeAgentDownload)

	// Web Dashboard
	mux.HandleFunc("/", ServeDashboard)

	// Health check
	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{
			"status":  "ok",
			"service": "meshlink-server",
		})
	})

	return corsMiddleware(mux)
}

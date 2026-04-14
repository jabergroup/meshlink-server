package api

import (
	"meshlink-server/internal/store"
	"net/http"
)

// corsMiddleware adds CORS headers
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

	// Initialize auth
	InitAuth()

	mux := http.NewServeMux()

	// Public routes (no auth)
	mux.HandleFunc("/api/login", HandleLogin)
	})
	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{
			"status":  "ok",
			"service": "meshlink-server",
		})
	})
	mux.HandleFunc("/download/agent", ServeAgentDownload)
	mux.HandleFunc("/", ServeDashboard)

	// Protected routes (require JWT)
	mux.HandleFunc("/api/pair/create", authMiddleware(h.CreatePair))
	mux.HandleFunc("/api/pair/join", authMiddleware(h.JoinPair))
	mux.HandleFunc("/api/pair/status", authMiddleware(h.GetPairStatus))
	mux.HandleFunc("/api/pair/update", authMiddleware(h.UpdateStatus))
	mux.HandleFunc("/api/heartbeat", authMiddleware(h.Heartbeat))
	mux.HandleFunc("/api/sessions", authMiddleware(h.ListSessions))

	return corsMiddleware(mux)
}

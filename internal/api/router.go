package api

import (
	"meshlink-server/internal/store"
	"net/http"
	"os"
	"strings"
)

// Global API key (set from environment or flag)
var apiKey string

// SetAPIKey sets the API key for authentication
func SetAPIKey(key string) {
	apiKey = key
}

// corsMiddleware adds CORS headers for the web dashboard
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// apiKeyMiddleware checks for valid API key on protected routes
func apiKeyMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// If no API key configured, allow all (development mode)
		if apiKey == "" {
			next(w, r)
			return
		}

		// Check X-API-Key header
		key := r.Header.Get("X-API-Key")

		// Fallback: check Authorization: Bearer <key>
		if key == "" {
			auth := r.Header.Get("Authorization")
			if strings.HasPrefix(auth, "Bearer ") {
				key = strings.TrimPrefix(auth, "Bearer ")
			}
		}

		// Fallback: check ?api_key= query parameter (for dashboard)
		if key == "" {
			key = r.URL.Query().Get("api_key")
		}

		if key != apiKey {
			writeJSON(w, http.StatusUnauthorized, map[string]string{
				"error": "invalid or missing API key",
			})
			return
		}

		next(w, r)
	}
}

// NewRouter creates the HTTP router with all API routes
func NewRouter(s *store.Store) http.Handler {
	h := NewHandlers(s)

	// Load API key from environment if not already set
	if apiKey == "" {
		apiKey = os.Getenv("MESHLINK_API_KEY")
	}

	mux := http.NewServeMux()

	// Protected API routes (require API key)
	mux.HandleFunc("/api/pair/create", apiKeyMiddleware(h.CreatePair))
	mux.HandleFunc("/api/pair/join", apiKeyMiddleware(h.JoinPair))
	mux.HandleFunc("/api/pair/status", apiKeyMiddleware(h.GetPairStatus))
	mux.HandleFunc("/api/pair/update", apiKeyMiddleware(h.UpdateStatus))
	mux.HandleFunc("/api/heartbeat", apiKeyMiddleware(h.Heartbeat))
	mux.HandleFunc("/api/sessions", apiKeyMiddleware(h.ListSessions))

	// Public routes (no API key needed)
	mux.HandleFunc("/download/agent", ServeAgentDownload)
	mux.HandleFunc("/", ServeDashboard)

	// Health check (public)
	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{
			"status":  "ok",
			"service": "meshlink-server",
		})
	})

	return corsMiddleware(mux)
}

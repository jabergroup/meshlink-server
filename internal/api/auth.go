package api

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	authUsername string
	authPassword string // stored as SHA-256 hash
	jwtSecret    []byte
)

// InitAuth sets up authentication credentials from environment variables
func InitAuth() {
	authUsername = os.Getenv("MESHLINK_USER")
	pass := os.Getenv("MESHLINK_PASS")

	fmt.Printf("  DEBUG: MESHLINK_USER='%s' (len=%d)\n", authUsername, len(authUsername))
	fmt.Printf("  DEBUG: MESHLINK_PASS set=%v (len=%d)\n", pass != "", len(pass))

	if authUsername == "" || pass == "" {
		fmt.Println("  WARNING: MESHLINK_USER / MESHLINK_PASS not set! Auth disabled.")
		return
	}

	// Hash the password
	h := sha256.Sum256([]byte(pass))
	authPassword = hex.EncodeToString(h[:])
	fmt.Printf("  DEBUG: passHash=%s\n", authPassword[:16])

	// Generate JWT secret
	jwtSecret = make([]byte, 32)
	rand.Read(jwtSecret)

	fmt.Printf("  Auth:       Enabled (user: %s)\n", authUsername)
}

// HandleLogin authenticates a user and returns a JWT token
func HandleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "POST only")
		return
	}

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	// Verify credentials
	h := sha256.Sum256([]byte(req.Password))
	passHash := hex.EncodeToString(h[:])

	if req.Username != authUsername || passHash != authPassword {
		time.Sleep(1 * time.Second) // Brute-force protection
		writeError(w, http.StatusUnauthorized, "invalid username or password")
		return
	}

	// Generate JWT token (valid for 30 days)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": req.Username,
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(30 * 24 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"token":    tokenString,
		"username": req.Username,
		"expires":  time.Now().Add(30 * 24 * time.Hour).Format(time.RFC3339),
	})
}

// authMiddleware checks for valid JWT token on protected routes
func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// If auth not configured, allow all
		if authUsername == "" {
			next(w, r)
			return
		}

		// Extract token from Authorization header or query param
		tokenStr := ""

		// Check Authorization: Bearer <token>
		auth := r.Header.Get("Authorization")
		if strings.HasPrefix(auth, "Bearer ") {
			tokenStr = strings.TrimPrefix(auth, "Bearer ")
		}

		// Fallback: ?token= query parameter
		if tokenStr == "" {
			tokenStr = r.URL.Query().Get("token")
		}

		if tokenStr == "" {
			writeJSON(w, http.StatusUnauthorized, map[string]string{
				"error": "authentication required",
			})
			return
		}

		// Validate JWT
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method")
			}
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			writeJSON(w, http.StatusUnauthorized, map[string]string{
				"error": "invalid or expired token",
			})
			return
		}

		next(w, r)
	}
}

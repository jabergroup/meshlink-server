package api

import (
	"encoding/json"
	"fmt"
	"meshlink-server/internal/store"
	"net/http"
	"os"
	"path/filepath"
)

// Handlers holds the API handlers
type Handlers struct {
	store *store.Store
}

// NewHandlers creates new API handlers
func NewHandlers(s *store.Store) *Handlers {
	return &Handlers{store: s}
}

// --- Request/Response types ---

type CreatePairResponse struct {
	Code      string `json:"code"`
	ExpiresIn int    `json:"expires_in"` // seconds
}

type JoinPairRequest struct {
	Code       string   `json:"code"`
	DeviceID   string   `json:"device_id"`
	Name       string   `json:"name"`
	IPv6       string   `json:"ipv6"`
	IPv4       string   `json:"ipv4"`
	PublicIP   string   `json:"public_ip"`
	PublicPort int      `json:"public_port"`
	NATType    string   `json:"nat_type"`
	Services   []string `json:"services"`
}

type JoinPairResponse struct {
	Role      string       `json:"role"`
	Session   *SessionInfo `json:"session"`
	AuthToken string       `json:"auth_token"`
	SecretKey string       `json:"secret_key"`
}

type SessionInfo struct {
	Code   string      `json:"code"`
	Status string      `json:"status"`
	Server *DeviceInfo `json:"server,omitempty"`
	Client *DeviceInfo `json:"client,omitempty"`
}

type DeviceInfo struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Role       string   `json:"role"`
	IPv6       string   `json:"ipv6"`
	IPv4       string   `json:"ipv4"`
	PublicIP   string   `json:"public_ip"`
	PublicPort int      `json:"public_port"`
	NATType    string   `json:"nat_type"`
	Services   []string `json:"services"`
	Status     string   `json:"status"`
}

type StatusUpdateRequest struct {
	Code     string `json:"code"`
	DeviceID string `json:"device_id"`
	Status   string `json:"status"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

// --- Handlers ---

// CreatePair generates a new pairing code
func (h *Handlers) CreatePair(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	code := h.store.GenerateCode()
	h.store.CreateSession(code)

	writeJSON(w, http.StatusOK, CreatePairResponse{
		Code:      code,
		ExpiresIn: 300, // 5 minutes
	})
}

// JoinPair adds a device to a pairing session
func (h *Handlers) JoinPair(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req JoinPairRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Code == "" || req.DeviceID == "" {
		writeError(w, http.StatusBadRequest, "code and device_id are required")
		return
	}

	// Generate security tokens
	authToken := store.GenerateToken("msl")
	secretKey := store.GenerateToken("msl_sk")

	device := &store.Device{
		ID:         req.DeviceID,
		Name:       req.Name,
		IPv6:       req.IPv6,
		IPv4:       req.IPv4,
		PublicIP:   req.PublicIP,
		PublicPort: req.PublicPort,
		NATType:    req.NATType,
		AuthToken:  authToken,
		SecretKey:  secretKey,
		Services:   req.Services,
		Status:     "online",
	}

	session, err := h.store.JoinSession(req.Code, device)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	resp := JoinPairResponse{
		Role:      device.Role,
		AuthToken: authToken,
		SecretKey: secretKey,
		Session:   toSessionInfo(session),
	}

	writeJSON(w, http.StatusOK, resp)
}

// GetPairStatus returns the status of a pairing session
func (h *Handlers) GetPairStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		writeError(w, http.StatusBadRequest, "code is required")
		return
	}

	session, ok := h.store.GetSession(code)
	if !ok {
		writeError(w, http.StatusNotFound, "session not found or expired")
		return
	}

	writeJSON(w, http.StatusOK, toSessionInfo(session))
}

// UpdateStatus updates the tunnel connection status
func (h *Handlers) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req StatusUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	h.store.UpdateStatus(req.Code, req.Status)
	h.store.UpdateDeviceStatus(req.DeviceID, req.Status)

	writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

// ListSessions returns all active sessions
func (h *Handlers) ListSessions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	sessions := h.store.GetAllSessions()
	var result []*SessionInfo
	for _, s := range sessions {
		result = append(result, toSessionInfo(s))
	}
	if result == nil {
		result = []*SessionInfo{}
	}

	writeJSON(w, http.StatusOK, result)
}

// Heartbeat receives periodic updates from agents (IP changes, keep-alive)
func (h *Handlers) Heartbeat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req struct {
		DeviceID string `json:"device_id"`
		IPv6     string `json:"ipv6"`
		IPv4     string `json:"ipv4"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	device, session := h.store.Heartbeat(req.DeviceID, req.IPv6, req.IPv4)
	if device == nil {
		writeError(w, http.StatusNotFound, "device not registered")
		return
	}

	// Return peer info so the agent knows the latest IP of the other device
	peer := h.store.GetPeerInfo(req.DeviceID)

	resp := map[string]interface{}{
		"status": "ok",
		"device": DeviceInfo{
			ID: device.ID, Name: device.Name, Role: device.Role,
			IPv6: device.IPv6, IPv4: device.IPv4, Status: device.Status,
		},
	}
	if peer != nil {
		resp["peer"] = DeviceInfo{
			ID: peer.ID, Name: peer.Name, Role: peer.Role,
			IPv6: peer.IPv6, IPv4: peer.IPv4, Status: peer.Status,
			Services: peer.Services,
		}
	}
	if session != nil {
		resp["session_status"] = session.Status
	}

	writeJSON(w, http.StatusOK, resp)
}

// --- Helpers ---

func toSessionInfo(s *store.PairSession) *SessionInfo {
	info := &SessionInfo{
		Code:   s.Code,
		Status: s.Status,
	}
	if s.Server != nil {
		info.Server = &DeviceInfo{
			ID:         s.Server.ID,
			Name:       s.Server.Name,
			Role:       s.Server.Role,
			IPv6:       s.Server.IPv6,
			IPv4:       s.Server.IPv4,
			PublicIP:   s.Server.PublicIP,
			PublicPort: s.Server.PublicPort,
			NATType:    s.Server.NATType,
			Services:   s.Server.Services,
			Status:     s.Server.Status,
		}
	}
	if s.Client != nil {
		info.Client = &DeviceInfo{
			ID:         s.Client.ID,
			Name:       s.Client.Name,
			Role:       s.Client.Role,
			IPv6:       s.Client.IPv6,
			IPv4:       s.Client.IPv4,
			PublicIP:   s.Client.PublicIP,
			PublicPort: s.Client.PublicPort,
			NATType:    s.Client.NATType,
			Services:   s.Client.Services,
			Status:     s.Client.Status,
		}
	}
	return info
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, ErrorResponse{Error: msg})
}

// ServeAgentDownload serves the MeshLink-Setup.exe installer for download
func ServeAgentDownload(w http.ResponseWriter, r *http.Request) {
	exe, err := os.Executable()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "cannot determine server path")
		return
	}
	baseDir := filepath.Dir(filepath.Dir(exe)) // go up from server/ to MeshLink/

	// Serve the installer exe (single file, professional)
	setupPath := filepath.Join(baseDir, "dist", "MeshLink-Setup.exe")
	if _, err := os.Stat(setupPath); os.IsNotExist(err) {
		setupPath = `C:\MeshLink\dist\MeshLink-Setup.exe`
	}

	if info, err := os.Stat(setupPath); err == nil {
		w.Header().Set("Content-Type", "application/vnd.microsoft.portable-executable")
		w.Header().Set("Content-Disposition", "attachment; filename=MeshLink-Setup.exe")
		w.Header().Set("Content-Length", fmt.Sprintf("%d", info.Size()))
		http.ServeFile(w, r, setupPath)
		return
	}

	writeError(w, http.StatusNotFound, "installer not found")
}

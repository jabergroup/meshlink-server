package store

import (
	"crypto/rand"
	"fmt"
	"sync"
	"time"
)

// Device represents a paired device
type Device struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Role       string    `json:"role"` // "server" or "client"
	IPv6       string    `json:"ipv6"`
	IPv4       string    `json:"ipv4"`
	PublicIP   string    `json:"public_ip"`
	PublicPort int       `json:"public_port"`
	NATType    string    `json:"nat_type"` // "none", "full_cone", "restricted", "symmetric", "unknown"
	AuthToken  string    `json:"auth_token"`
	SecretKey  string    `json:"secret_key"`
	Services   []string  `json:"services"`
	Status     string    `json:"status"` // "waiting", "paired", "connected", "offline"
	LastSeen   time.Time `json:"last_seen"`
}

// PairSession represents a pairing session between two devices
type PairSession struct {
	Code      string    `json:"code"`
	Server    *Device   `json:"server,omitempty"`
	Client    *Device   `json:"client,omitempty"`
	Status    string    `json:"status"` // "waiting", "paired", "connected", "expired"
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// Store is an in-memory store for pairing sessions and devices
type Store struct {
	mu       sync.RWMutex
	sessions map[string]*PairSession
	devices  map[string]*Device
}

// New creates a new Store
func New() *Store {
	s := &Store{
		sessions: make(map[string]*PairSession),
		devices:  make(map[string]*Device),
	}
	go s.cleanupLoop()
	return s
}

// GenerateCode creates a unique 6-digit pairing code
func (s *Store) GenerateCode() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	for {
		code := randomCode()
		if _, exists := s.sessions[code]; !exists {
			return code
		}
	}
}

// CreateSession creates a new pairing session
func (s *Store) CreateSession(code string) *PairSession {
	s.mu.Lock()
	defer s.mu.Unlock()

	session := &PairSession{
		Code:      code,
		Status:    "waiting",
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}
	s.sessions[code] = session
	return session
}

// GetSession returns a session by code
func (s *Store) GetSession(code string) (*PairSession, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session, ok := s.sessions[code]
	if !ok {
		return nil, false
	}
	if time.Now().After(session.ExpiresAt) {
		return nil, false
	}
	return session, true
}

// JoinSession adds a device to an existing session
func (s *Store) JoinSession(code string, device *Device) (*PairSession, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, ok := s.sessions[code]
	if !ok {
		return nil, fmt.Errorf("session not found")
	}
	if time.Now().After(session.ExpiresAt) {
		delete(s.sessions, code)
		return nil, fmt.Errorf("session expired")
	}

	if session.Server == nil {
		device.Role = "server"
		session.Server = device
		session.Status = "waiting"
	} else if session.Client == nil {
		device.Role = "client"
		session.Client = device
		session.Status = "paired"
	} else {
		return nil, fmt.Errorf("session is full")
	}

	s.devices[device.ID] = device
	return session, nil
}

// UpdateStatus updates the tunnel status
func (s *Store) UpdateStatus(code string, status string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if session, ok := s.sessions[code]; ok {
		session.Status = status
	}
}

// UpdateDeviceStatus updates a device's status and last seen time
func (s *Store) UpdateDeviceStatus(deviceID string, status string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if device, ok := s.devices[deviceID]; ok {
		device.Status = status
		device.LastSeen = time.Now()
	}
}

// GetAllSessions returns all active sessions
func (s *Store) GetAllSessions() []*PairSession {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*PairSession
	now := time.Now()
	for _, session := range s.sessions {
		if now.Before(session.ExpiresAt) || session.Status == "connected" {
			result = append(result, session)
		}
	}
	return result
}

// DeleteSession removes a session
func (s *Store) DeleteSession(code string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, code)
}

// Heartbeat updates a device's IP and marks it alive. Returns the paired device info.
func (s *Store) Heartbeat(deviceID, ipv6, ipv4 string) (*Device, *PairSession) {
	s.mu.Lock()
	defer s.mu.Unlock()

	device, ok := s.devices[deviceID]
	if !ok {
		return nil, nil
	}

	device.IPv6 = ipv6
	device.IPv4 = ipv4
	device.Status = "online"
	device.LastSeen = time.Now()

	// Find the session this device belongs to
	for _, session := range s.sessions {
		if (session.Server != nil && session.Server.ID == deviceID) ||
			(session.Client != nil && session.Client.ID == deviceID) {
			// Update the device in the session too
			if session.Server != nil && session.Server.ID == deviceID {
				session.Server.IPv6 = ipv6
				session.Server.IPv4 = ipv4
				session.Server.Status = "online"
				session.Server.LastSeen = time.Now()
			}
			if session.Client != nil && session.Client.ID == deviceID {
				session.Client.IPv6 = ipv6
				session.Client.IPv4 = ipv4
				session.Client.Status = "online"
				session.Client.LastSeen = time.Now()
			}
			return device, session
		}
	}

	return device, nil
}

// GetPeerInfo returns the paired device's current info
func (s *Store) GetPeerInfo(deviceID string) *Device {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, session := range s.sessions {
		if session.Server != nil && session.Server.ID == deviceID && session.Client != nil {
			return session.Client
		}
		if session.Client != nil && session.Client.ID == deviceID && session.Server != nil {
			return session.Server
		}
	}
	return nil
}

// SavePersistentPair stores a permanent pairing (survives restarts)
func (s *Store) SavePersistentPair(session *PairSession) {
	s.mu.Lock()
	defer s.mu.Unlock()
	// Make session permanent (no expiry)
	session.ExpiresAt = time.Now().Add(100 * 365 * 24 * time.Hour) // 100 years
	session.Status = "connected"
}

// cleanupLoop periodically removes expired sessions
func (s *Store) cleanupLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	for range ticker.C {
		s.mu.Lock()
		now := time.Now()
		for code, session := range s.sessions {
			if session.Status != "connected" && now.After(session.ExpiresAt) {
				delete(s.sessions, code)
			}
		}
		s.mu.Unlock()
	}
}

// randomCode generates a 6-digit numeric code
func randomCode() string {
	b := make([]byte, 3)
	rand.Read(b)
	num := (int(b[0])<<16 | int(b[1])<<8 | int(b[2])) % 1000000
	return fmt.Sprintf("%06d", num)
}

// GenerateToken generates a secure random token
func GenerateToken(prefix string) string {
	b := make([]byte, 24)
	rand.Read(b)
	return fmt.Sprintf("%s_%x", prefix, b)
}

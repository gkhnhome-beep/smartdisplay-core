// Package guest provides guest access request management.
// FAZ L2: Guest approval flow
package guest

import (
	"errors"
	"fmt"
	"smartdisplay-core/internal/logger"
	"sync"
	"time"
)

// Request status constants
const (
	StatusPending  = "pending"
	StatusApproved = "approved"
	StatusRejected = "rejected"
	StatusExpired  = "expired"
)

// GuestRequest represents a single guest access request
type GuestRequest struct {
	ID          string    `json:"id"`
	TargetUser  string    `json:"target_user"`
	Status      string    `json:"status"`
	RequestedAt time.Time `json:"requested_at"`
	ExpiresAt   time.Time `json:"expires_at"`
}

// Manager handles guest access requests with in-memory storage
type Manager struct {
	mu            sync.RWMutex
	activeRequest *GuestRequest
	timeout       time.Duration
	onApproved    func(*GuestRequest) error
	onRejected    func(*GuestRequest) error
	expireTimer   *time.Timer
}

// NewManager creates a new guest request manager with specified timeout
func NewManager(timeout time.Duration) *Manager {
	if timeout == 0 {
		timeout = 60 * time.Second // Default 60 seconds
	}
	return &Manager{
		timeout: timeout,
	}
}

// SetApprovedCallback sets handler for approved requests
func (m *Manager) SetApprovedCallback(fn func(*GuestRequest) error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onApproved = fn
}

// SetRejectedCallback sets handler for rejected requests
func (m *Manager) SetRejectedCallback(fn func(*GuestRequest) error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onRejected = fn
}

// CreateRequest creates a new guest access request
// Returns error if another request is already active
func (m *Manager) CreateRequest(targetUser string) (*GuestRequest, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Only one active request at a time
	if m.activeRequest != nil && m.activeRequest.Status == StatusPending {
		return nil, errors.New("guest request already pending")
	}

	now := time.Now()
	req := &GuestRequest{
		ID:          fmt.Sprintf("greq-%d", now.UnixNano()),
		TargetUser:  targetUser,
		Status:      StatusPending,
		RequestedAt: now,
		ExpiresAt:   now.Add(m.timeout),
	}

	m.activeRequest = req

	// Start expiration timer
	m.startExpirationTimer(req.ID)

	logger.Info("guest request created: id=" + req.ID + " target=" + targetUser)

	return req, nil
}

// GetActiveRequest returns the currently active request (if any)
func (m *Manager) GetActiveRequest() *GuestRequest {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.activeRequest == nil {
		return nil
	}

	// Return nil if request has expired
	if m.activeRequest.Status == StatusPending && time.Now().After(m.activeRequest.ExpiresAt) {
		return nil
	}

	return m.activeRequest
}

// ApproveRequest marks a request as approved
func (m *Manager) ApproveRequest(requestID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.activeRequest == nil || m.activeRequest.ID != requestID {
		return errors.New("request not found")
	}

	if m.activeRequest.Status != StatusPending {
		return errors.New("request is not pending")
	}

	m.activeRequest.Status = StatusApproved
	req := m.activeRequest

	// Stop expiration timer
	if m.expireTimer != nil {
		m.expireTimer.Stop()
		m.expireTimer = nil
	}

	logger.Info("guest request approved: id=" + req.ID)

	// Call approved callback if set
	if m.onApproved != nil {
		go func() {
			if err := m.onApproved(req); err != nil {
				logger.Error("approval callback failed: " + err.Error())
			}
		}()
	}

	return nil
}

// RejectRequest marks a request as rejected
func (m *Manager) RejectRequest(requestID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.activeRequest == nil || m.activeRequest.ID != requestID {
		return errors.New("request not found")
	}

	if m.activeRequest.Status != StatusPending {
		return errors.New("request is not pending")
	}

	m.activeRequest.Status = StatusRejected
	req := m.activeRequest

	// Stop expiration timer
	if m.expireTimer != nil {
		m.expireTimer.Stop()
		m.expireTimer = nil
	}

	logger.Info("guest request rejected: id=" + req.ID)

	// Call rejected callback if set
	if m.onRejected != nil {
		go func() {
			if err := m.onRejected(req); err != nil {
				logger.Error("rejection callback failed: " + err.Error())
			}
		}()
	}

	return nil
}

// ClearRequest clears the active request
func (m *Manager) ClearRequest() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.expireTimer != nil {
		m.expireTimer.Stop()
		m.expireTimer = nil
	}

	m.activeRequest = nil
}

// startExpirationTimer starts the expiration timer for a request
func (m *Manager) startExpirationTimer(requestID string) {
	if m.expireTimer != nil {
		m.expireTimer.Stop()
	}

	m.expireTimer = time.AfterFunc(m.timeout, func() {
		m.mu.Lock()
		defer m.mu.Unlock()

		// Only expire if the request is still pending
		if m.activeRequest != nil && m.activeRequest.ID == requestID && m.activeRequest.Status == StatusPending {
			m.activeRequest.Status = StatusExpired
			logger.Info("guest request expired: id=" + requestID)
		}
	})
}

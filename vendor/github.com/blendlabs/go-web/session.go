package web

import (
	"sync"
	"time"
)

// NewSession returns a new session object.
func NewSession(userID string, sessionID string) *Session {
	return &Session{
		UserID:     userID,
		SessionID:  sessionID,
		CreatedUTC: time.Now().UTC(),
		State:      map[string]interface{}{},
		Mutex:      &sync.RWMutex{},
	}
}

// Session is an active session
type Session struct {
	UserID     string                 `json:"userID" yaml:"userID"`
	SessionID  string                 `json:"sessionID" yaml:"sessionID"`
	CreatedUTC time.Time              `json:"createdUTC" yaml:"createdUTC"`
	State      map[string]interface{} `json:"state,omitempty" yaml:"state,omitempty"`
	Mutex      *sync.RWMutex          `json:"-" yaml:"-"`
}

func (s *Session) ensureMutex() {
	if s.Mutex == nil {
		s.Mutex = &sync.RWMutex{}
	}
}

// IsZero returns if the object is set or not.
func (s *Session) IsZero() bool {
	return len(s.UserID) == 0 || len(s.SessionID) == 0
}

// Lock locks the session.
func (s *Session) Lock() {
	s.ensureMutex()
	s.Mutex.Lock()
}

// Unlock unlocks the session.
func (s *Session) Unlock() {
	s.ensureMutex()
	s.Mutex.Unlock()
}

// RLock read locks the session.
func (s *Session) RLock() {
	s.ensureMutex()
	s.Mutex.RLock()
}

// RUnlock read unlocks the session.
func (s *Session) RUnlock() {
	s.ensureMutex()
	s.Mutex.RUnlock()
}

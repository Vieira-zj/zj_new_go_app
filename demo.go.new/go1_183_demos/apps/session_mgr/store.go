package main

import (
	"sync"
	"time"
)

type SessionContextKeyType string

const SessionContextKey SessionContextKeyType = "x-session"

type SessionStore interface {
	read(id string) (Session, error)
	write(session Session) error
	destroy(id string) error
	gc(idleExpiration, absoluteExpiration time.Duration) error
}

// InMemorySessionStore

type InMemorySessionStore struct {
	mu       sync.RWMutex
	sessions map[string]Session
}

func NewInMemorySessionStore() *InMemorySessionStore {
	return &InMemorySessionStore{
		sessions: make(map[string]Session),
	}
}

func (s *InMemorySessionStore) read(id string) (Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session := s.sessions[id]
	return session, nil
}

func (s *InMemorySessionStore) write(session Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.sessions[session.id] = session
	return nil
}

func (s *InMemorySessionStore) destroy(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.sessions, id)
	return nil
}

func (s *InMemorySessionStore) gc(idleExpiration, absoluteExpiration time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for id, session := range s.sessions {
		if time.Since(session.lastActivityAt) > idleExpiration ||
			time.Since(session.createdAt) > absoluteExpiration {
			delete(s.sessions, id)
		}
	}
	return nil
}

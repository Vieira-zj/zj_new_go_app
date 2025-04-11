package main

import (
	"context"
	"log"
	"net/http"
	"time"
)

type SessionManager struct {
	store              SessionStore
	idleExpiration     time.Duration
	absoluteExpiration time.Duration
	cookieName         string
}

func NewSessionManager(store SessionStore, gcInterval, idleExpiration, absoluteExpiration time.Duration, cookieName string) *SessionManager {
	m := &SessionManager{
		store:              store,
		idleExpiration:     idleExpiration,
		absoluteExpiration: absoluteExpiration,
		cookieName:         cookieName,
	}

	go m.gc(gcInterval) // remove expired session

	return m
}

func (m *SessionManager) gc(d time.Duration) {
	ticker := time.NewTicker(d)
	for range ticker.C {
		m.store.gc(m.idleExpiration, m.absoluteExpiration)
	}
}

func (m *SessionManager) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// start the session
		session, rws := m.start(r)

		// create a new response writer
		sw := SessionResponseWriter{
			ResponseWriter: w,
			sessionManager: m,
			request:        rws,
		}

		// add essential headers
		w.Header().Add("Vary", "Cookie")
		w.Header().Add("Cache-Control", `no-cache="Set-Cookie"`)

		// call the next handler and pass the new response writer and new request
		next.ServeHTTP(sw, rws)

		// save the session
		m.save(session)

		// write the session cookie to the response if not already written
		writeCookieIfNecessary(sw)
	})
}

func (m *SessionManager) validate(session Session) bool {
	if time.Since(session.createdAt) > m.absoluteExpiration ||
		time.Since(session.lastActivityAt) > m.idleExpiration {
		// delete the session from the store
		if err := m.store.destroy(session.id); err != nil {
			panic(err)
		}
		return false
	}

	return true
}

func (m *SessionManager) start(r *http.Request) (Session, *http.Request) {
	var session Session

	// read from cookie
	cookie, err := r.Cookie(m.cookieName)
	if err == nil {
		session, err = m.store.read(cookie.Value)
		if err != nil {
			log.Printf("failed to read session from store: %v", err)
		}
	}

	// generate a new session
	if len(session.id) == 0 || !m.validate(session) {
		session = NewSession()
	}

	// attach session to context
	ctx := context.WithValue(r.Context(), SessionContextKey, session)
	r = r.WithContext(ctx)

	return session, r
}

func (m *SessionManager) save(session Session) error {
	session.lastActivityAt = time.Now()
	if err := m.store.write(session); err != nil {
		return err
	}
	return nil
}

func (m *SessionManager) Migrate(session Session) error {
	if err := m.store.destroy(session.id); err != nil {
		return err
	}

	session.id = generateSessionId()
	return nil
}

func GetSession(r *http.Request) Session {
	session, ok := r.Context().Value(SessionContextKey).(Session)
	if !ok {
		panic("session not found in request context")
	}

	return session
}

package main

import (
	"crypto/rand"
	"encoding/base64"
	"io"
	"net/http"
	"time"
)

// Session

type Session struct {
	id             string
	createdAt      time.Time
	lastActivityAt time.Time
	data           map[string]any
}

func NewSession() Session {
	now := time.Now()
	return Session{
		id:             generateSessionId(),
		data:           make(map[string]any),
		createdAt:      now,
		lastActivityAt: now,
	}
}

func (s Session) Get(key string) any {
	return s.data[key]
}

func (s Session) Put(key string, value any) {
	s.data[key] = value
}

func (s Session) Delete(key string) {
	delete(s.data, key)
}

func generateSessionId() string {
	id := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, id); err != nil {
		panic("failed to generate session id")
	}

	return base64.RawURLEncoding.EncodeToString(id)
}

// SessionResponseWriter

type SessionResponseWriter struct {
	http.ResponseWriter
	sessionManager *SessionManager
	request        *http.Request
	done           bool
}

func (w SessionResponseWriter) Write(b []byte) (int, error) {
	writeCookieIfNecessary(w)
	return w.ResponseWriter.Write(b)
}

func (w SessionResponseWriter) WriteHeader(code int) {
	writeCookieIfNecessary(w)
	w.ResponseWriter.WriteHeader(code)
}

func (w SessionResponseWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

func writeCookieIfNecessary(w SessionResponseWriter) {
	if w.done {
		return
	}

	session, ok := w.request.Context().Value(SessionContextKey).(Session)
	if !ok {
		panic("session not found in request context")
	}

	cookie := http.Cookie{
		Name:     w.sessionManager.cookieName,
		Value:    session.id, // save session_id in cookie "value"
		Domain:   "mywebsite.com",
		HttpOnly: true,
		Path:     "/",
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(w.sessionManager.idleExpiration),
		MaxAge:   int(w.sessionManager.idleExpiration / time.Second),
	}

	http.SetCookie(w.ResponseWriter, &cookie)
	w.done = true
}

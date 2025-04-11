package main

import (
	"net/http"
	"time"
)

// Demo: http session manager

func main() {
	sessionMgr := NewSessionManager(NewInMemorySessionStore(), 30*time.Minute, time.Hour, 4*time.Hour, "x-session-mgr")

	mux := http.NewServeMux()

	mux.HandleFunc("/projects/switch/some-project-id", func(w http.ResponseWriter, r *http.Request) {
		session := GetSession(r)
		session.Put("current_project", "some-project-id")
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/project", func(w http.ResponseWriter, r *http.Request) {
		session := GetSession(r)
		curProject, ok := session.Get("current_project").(string)
		if ok {
			w.Write([]byte("current project: " + curProject))
		} else {
			w.Write([]byte("current project: unknown"))
		}
	})

	server := &http.Server{
		Addr:    ":8081",
		Handler: sessionMgr.Handle(mux),
	}

	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}
}

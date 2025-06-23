package main

import (
	"context"
	"log"
	"net/http"

	"github.com/casbin/casbin/v2"
)

type UserCtxKey struct{}

func UserMiddleware(users Users) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			user, err := getUser(r, users)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(err.Error()))
				return
			}

			ctx := context.WithValue(r.Context(), UserCtxKey{}, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		}

		return http.HandlerFunc(fn)
	}
}

func AuthMiddleware(e *casbin.Enforcer, users Users) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			user, err := getUserFromContext(r.Context())
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}

			ok, err := e.Enforce(user.Role, r.URL.Path, r.Method)
			if err != nil {
				log.Printf("enforce casbin rules failed: %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
				return
			}

			if !ok {
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte(http.StatusText(http.StatusForbidden)))
				return
			}

			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(fn)
	}
}

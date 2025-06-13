package main

import (
	"log"
	"net/http"

	"github.com/casbin/casbin/v2"
)

func Authorizer(e *casbin.Enforcer, users Users) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			user := User{
				Role: "anonymous",
			}

			if name := r.URL.Query().Get("name"); len(name) > 0 {
				var ok bool
				if user, ok = GetSession().FindUser(name); !ok {
					var err error
					user, err = users.FindByName(name) // if login user not in session, then get from db
					if err != nil {
						log.Println("user not found:", name)
						w.WriteHeader(http.StatusBadRequest)
						w.Write([]byte(http.StatusText(http.StatusBadRequest)))
						return
					}
				}
			}

			GetSession().SetCurUser(user) // set current access user to session

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
			return
		}

		return http.HandlerFunc(fn)
	}
}

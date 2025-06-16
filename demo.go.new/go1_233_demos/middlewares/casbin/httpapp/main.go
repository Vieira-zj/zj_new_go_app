package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/casbin/casbin/v2"
)

// curl "http://localhost:8081/login" -H "name:Sabine"
// curl "http://localhost:8081/logout" -H "name:Sabine"
// curl "http://localhost:8081/member/role" -H "name:Sabine"
// curl "http://localhost:8081/member/id" -H "name:Sabine"
// curl "http://localhost:8081/admin/total" -H "name:Sabine"

func main() {
	authEnforcer, err := casbin.NewEnforcer("./conf/auth_model.conf", "./conf/policy.csv")
	if err != nil {
		log.Fatal(err)
	}

	users := mockInitUsers()

	mux := http.NewServeMux()
	mux.HandleFunc("/login", loginHandler())
	mux.HandleFunc("/logout", logoutHandler())
	mux.HandleFunc("/member/id", getUserIdHandler())
	mux.HandleFunc("/member/role", getUserRoleHandler())
	mux.HandleFunc("/admin/total", getTotalHandler(users))

	handler := AuthMiddleware(authEnforcer, users)(mux)
	handler = UserMiddleware(users)(handler)

	log.Print("http server started on :8081")
	if err := http.ListenAndServe(":8081", handler); err != nil {
		log.Fatal(err)
	}
}

// Handlers

func loginHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if user, ok := GetSession().GetUser(r.Context()); ok { // get cached login user
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("hello, " + user.Name))
			return
		}

		user, err := getUserFromContext(r.Context())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		if user.isAnonymous() {
			w.Write([]byte("anonymous login"))
			return
		}

		w.WriteHeader(http.StatusOK)
		if mockCheckPwd(user) {
			GetSession().Save(user) // store login user in session
			w.Write([]byte("hello, " + user.Name))
		} else {
			w.Write([]byte("wrong password"))
		}
	}
}

func logoutHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if user, ok := GetSession().GetUser(r.Context()); ok {
			GetSession().Clear(user.Name)
			w.Write([]byte("bye, " + user.Name))
		} else {
			w.Write([]byte(http.StatusText(http.StatusOK)))
		}
	}
}

func getUserIdHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if user, ok := GetSession().GetUser(r.Context()); ok {
			w.Write([]byte(strconv.Itoa(user.ID)))
		} else {
			w.Write([]byte("user is not login"))
		}
	}
}

func getUserRoleHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if user, ok := GetSession().GetUser(r.Context()); ok {
			w.Write([]byte(user.Role))
		} else {
			w.Write([]byte("user is not login"))
		}
	}
}

func getTotalHandler(users Users) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, ok := GetSession().GetUser(r.Context()); ok {
			w.Write([]byte(strconv.Itoa(len(users))))
		} else {
			w.Write([]byte("user is not login"))
		}
	}
}

// Utils

func getUserFromContext(ctx context.Context) (User, error) {
	user, ok := ctx.Value(UserCtxKey{}).(User)
	if !ok {
		return User{}, fmt.Errorf("user is not set in context")
	}
	return user, nil
}

func getUser(r *http.Request, users Users) (User, error) {
	name := r.Header.Get("name")
	if len(name) == 0 {
		return User{Role: "anonymous"}, nil
	}

	// get user from session, if not login, try to get from db
	if user, ok := GetSession().FindUser(name); ok {
		return user, nil
	}
	return users.FindByName(name)
}

// Mock

func mockInitUsers() Users {
	return Users{
		{ID: 1, Name: "Admin", Role: "admin"},
		{ID: 2, Name: "Sabine", Role: "member"},
		{ID: 3, Name: "Sepp", Role: "member"},
	}
}

func mockCheckPwd(user User) bool {
	log.Println("check password for user:", user.Name, user.Role)
	return true
}

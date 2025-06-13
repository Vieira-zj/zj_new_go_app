package main

import (
	"log"
	"net/http"
	"strconv"

	"github.com/casbin/casbin/v2"
)

// curl "http://localhost:8081/login?name=Sabine"
// curl "http://localhost:8081/logout?name=Sabine"
// curl "http://localhost:8081/member/role?name=Sabine"
// curl "http://localhost:8081/member/id?name=Sabine"
// curl "http://localhost:8081/admin/total?name=Sabine"

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

	log.Print("http server started on :8080")
	handler := Authorizer(authEnforcer, users)(mux)
	if err := http.ListenAndServe(":8081", handler); err != nil {
		log.Fatal(err)
	}
}

func loginHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if user, ok := GetSession().GetUser(); ok { // get cached login user
			w.Write([]byte("hello, " + user.Name))
			return
		}

		user := GetSession().GetCurUser()
		if user.isAnonymous() {
			w.Write([]byte("anonymous login"))
			return
		}

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
		if user, ok := GetSession().GetUser(); ok {
			GetSession().Clear(user.Name)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(http.StatusText(http.StatusOK)))
	}
}

func getUserIdHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if user, ok := GetSession().GetUser(); ok {
			w.Write([]byte(strconv.Itoa(user.ID)))
		} else {
			w.Write([]byte("user is not login"))
		}
	}
}

func getUserRoleHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if user, ok := GetSession().GetUser(); ok {
			w.Write([]byte(user.Role))
		} else {
			w.Write([]byte("user is not login"))
		}
	}
}

func getTotalHandler(users Users) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, ok := GetSession().GetUser(); ok {
			w.Write([]byte(strconv.Itoa(len(users))))
		} else {
			w.Write([]byte("user is not login"))
		}
	}
}

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

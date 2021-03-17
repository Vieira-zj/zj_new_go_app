package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo"
)

func TestGetUsers(t *testing.T) {
	expect := "Users"
	ret := getUsers()
	if ret != expect {
		t.Errorf("expect %s and actual %s.", expect, ret)
	}
}

func TestUsers(t *testing.T) {
	req, err := http.NewRequest("GET", "/users", nil)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()

	c := echo.New().NewContext(req, w)
	if err = Users(c); err != nil {
		t.Fatal(err)
	}

	expect := "Users"
	ret := string(w.Body.Bytes())
	if ret != expect {
		t.Errorf("expect %s and actual %s.", expect, ret)
	}
}

func TestGetUsersNew(t *testing.T) {
	expect := "UsersNew"
	ret := getUsersNew()
	if ret != expect {
		t.Errorf("expect %s and actual %s.", expect, ret)
	}
}

func TestUsersNew(t *testing.T) {
	req, err := http.NewRequest("GET", "/users/new", nil)
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()

	c := echo.New().NewContext(req, w)
	if err = UsersNew(c); err != nil {
		t.Fatal(err)
	}

	expect := "UsersNew"
	ret := string(w.Body.Bytes())
	if ret != expect {
		t.Errorf("expect %s and actual %s.", expect, ret)
	}
}

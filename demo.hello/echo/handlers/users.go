package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/labstack/echo"
)

// echo 复杂路由

// User .
type User struct {
	Name string `json:"name"`
	Aage int    `json:"age"`
}

// Users return all users.
func Users(c echo.Context) error {
	return c.String(http.StatusOK, getUsers())
}

func getUsers() string {
	return "Users"
}

// UsersNew new a user.
func UsersNew(c echo.Context) error {
	// body set from deco
	body := c.Get("req_body")
	user := &User{}
	if err := json.Unmarshal(body.([]byte), user); err != nil {
		c.String(http.StatusInternalServerError, "Json Unmarshal error: "+err.Error())
	}
	return c.JSON(http.StatusOK, user)
}

func getUsersNew() string {
	return "UsersNew"
}

// UsersName return user name.
func UsersName(c echo.Context) error {
	name := c.Param("name")
	return c.String(http.StatusOK, fmt.Sprintf("%s, %s", "Hi", name))
}

// UsersFiles return user files.
func UsersFiles(c echo.Context) error {
	prefix := "/users/1/files/"
	subPath := c.Request().URL.Path[len(prefix):]
	return c.String(http.StatusOK, "Query file: "+subPath)
}

package handlers

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo"
)

// echo 复杂路由

// Users return all users.
func Users(c echo.Context) error {
	return c.String(http.StatusOK, getUsers())
}

func getUsers() string {
	return "Users"
}

// UsersNew new a user.
func UsersNew(c echo.Context) error {
	return c.String(http.StatusOK, getUsersNew())
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

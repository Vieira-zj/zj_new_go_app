package main

import (
	"fmt"
	"slices"
)

type User struct {
	ID   int
	Name string
	Role string
}

func (u User) isAnonymous() bool {
	return u.Role == "anonymous"
}

type Users []User

func (u Users) Exists(id int) bool {
	return slices.ContainsFunc(u, func(user User) bool {
		return user.ID == id
	})
}

func (u Users) FindByName(name string) (User, error) {
	for _, user := range u {
		if user.Name == name {
			return user, nil
		}
	}
	return User{}, fmt.Errorf("user not found")
}

package main

import "context"

type Session struct {
	store map[string]User // store login users
}

var session *Session

func NewSession() *Session {
	return &Session{
		store: make(map[string]User, 4),
	}
}

func GetSession() *Session {
	if session == nil {
		session = NewSession()
	}
	return session
}

func (s Session) Save(user User) {
	s.store[user.Name] = user
}

func (s Session) Clear(name string) {
	delete(s.store, name)
}

func (s Session) FindUser(name string) (User, bool) {
	user, ok := s.store[name]
	return user, ok
}

func (s Session) GetUser(ctx context.Context) (User, bool) {
	curUser, err := getUserFromContext(ctx)
	if err != nil {
		return User{}, false
	}

	user, ok := s.store[curUser.Name]
	return user, ok
}

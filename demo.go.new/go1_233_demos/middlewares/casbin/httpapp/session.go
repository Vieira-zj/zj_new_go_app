package main

type Session struct {
	user  User            // current user
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

func (s *Session) SetCurUser(user User) {
	s.user = user
}

func (s *Session) GetCurUser() User {
	return s.user
}

func (s Session) FindUser(name string) (User, bool) {
	user, ok := s.store[name]
	return user, ok
}

func (s Session) GetUser() (User, bool) {
	user, ok := s.store[s.user.Name]
	return user, ok
}

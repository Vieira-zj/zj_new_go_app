package xormmock

import (
	"fmt"

	"github.com/go-xorm/xorm"
)

// Repository person table repo.
type Repository interface {
	Get(id int) (*Person, error)
	Create(id int, name string) error
	Update(id int, name string) error
	Delete(id int) error
}

type repo struct {
	session *xorm.Session
}

// NewPersonRepo returns person repo instance.
func NewPersonRepo(session *xorm.Session) Repository {
	return &repo{session: session}
}

func (r *repo) Get(id int) (*Person, error) {
	person := &Person{ID: id}
	has, err := r.session.Get(person)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, fmt.Errorf("person[id=%d] not found", id)
	}
	return person, nil
}

func (r *repo) Create(id int, name string) error {
	person := &Person{ID: id, Name: name}
	affected, err := r.session.Insert(person)
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("insert err, because of 0 affected")
	}
	return nil
}

func (r *repo) Update(id int, name string) (err error) {
	_, err = r.session.ID(id).Cols("name").Update(&Person{Name: name})
	return
}

func (r *repo) Delete(id int) (err error) {
	_, err = r.session.ID(id).Delete(&Person{})
	return
}

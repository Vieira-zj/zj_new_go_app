package xormmock

import (
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
)

/*
xorm mock demo just for mysql db.
refer: https://github.com/mbyd916/blog-hg/issues/8
*/

func TestPersonGet(t *testing.T) {
	// setup
	session, mock, err := getSession()
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	repo := NewPersonRepo(session)

	// get case1: found
	id, name := 1, "John"
	mock.ExpectQuery("SELECT (.+) FROM `person`").WithArgs(id).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(id, name))
	person, err := repo.Get(id)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("get person: %+v\n", *person)

	if person.Name != name {
		t.Fatalf("want %s and get %s", name, person.Name)
	}

	// get case2: not found
	mock.ExpectQuery("SELECT (.+) FROM `person`").WithArgs(id).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name"}))
	person, err = repo.Get(id)
	if err == nil {
		t.Fatal("Should be not found error.")
	}
	fmt.Println(err.Error())

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPersonCreate(t *testing.T) {
	// setup
	session, mock, err := getSession()
	if err != nil {
		t.Fatal(err)
	}
	defer session.Clone()

	repo := NewPersonRepo(session)

	// create case1: success
	id, name := 1, "John"
	mock.ExpectExec("INSERT INTO `person`").WithArgs(id, name).
		WillReturnResult(sqlmock.NewResult(1, 1))
	if err := repo.Create(id, name); err != nil {
		t.Fatal(err)
	}

	// create case2: failed
	mock.ExpectExec("INSERT INTO `person`").WithArgs(id, name).
		WillReturnResult(sqlmock.NewResult(0, 0))
	err = repo.Create(id, name)
	if err == nil {
		t.Fatal("Should be create error.")
	}
	fmt.Println(err.Error())

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPersonUpdate(t *testing.T) {
	// setup
	session, mock, err := getSession()
	if err != nil {
		t.Fatal(err)
	}
	defer session.Clone()

	repo := NewPersonRepo(session)

	// update case1
	id, name := 1, "John"
	mock.ExpectExec("UPDATE `person`").WithArgs(name, id).
		WillReturnResult(sqlmock.NewResult(1, 1))
	if err := repo.Update(id, name); err != nil {
		t.Fatal(err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestPersonDelete(t *testing.T) {
	// setup
	session, mock, err := getSession()
	if err != nil {
		t.Fatal(err)
	}
	defer session.Clone()

	repo := NewPersonRepo(session)

	// delete case1
	id := 1
	mock.ExpectExec("DELETE FROM `person`").WithArgs(id).
		WillReturnResult(sqlmock.NewResult(1, 1))
	if err := repo.Delete(id); err != nil {
		t.Fatal(err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

// getSession returns a xorm session with mock db.
func getSession() (*xorm.Session, sqlmock.Sqlmock, error) {
	db, mock, err := sqlmock.New()
	if err != nil {
		return nil, nil, err
	}

	eng, err := xorm.NewEngine("mysql", "root:123@/test?charset=utf8")
	if err != nil {
		return nil, nil, err
	}

	eng.DB().DB = db
	eng.ShowSQL(true)

	return eng.NewSession(), mock, nil
}

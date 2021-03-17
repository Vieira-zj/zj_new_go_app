package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

type user struct {
	ID          int64  `db:"id"`
	Name        string `db:"name"`
	NickName    string `db:"nickname"`
	Password    string `db:"password"`
	IsSuperUser string `db:"issuperuser"`
	// Picture     string `db:"picture"`
	Picture sql.NullString `db:"picture"`
}

const (
	username = "root"
	password = "example"
	network  = "tcp"
	server   = "localhost"
	port     = 13306
	database = "test"
)

/*
go sql
*/

func buildDBConnection() (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@%s(%s:%d)/%s", username, password, network, server, port, database)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("Open mysql failed, error: %v", err)
	}

	db.SetConnMaxLifetime(time.Duration(60) * time.Second)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(3)
	return db, nil
}

func queryOne(db *sql.DB) {
	fmt.Println("queryOne")
	u := new(user)
	row := db.QueryRow("select * from users where id=?", 2)
	if err := row.Scan(&u.ID, &u.Name, &u.NickName, &u.Password, &u.IsSuperUser, &u.Picture); err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("Row is not found")
		} else {
			fmt.Printf("Scan failed, error: %v\n", err)
		}
		return
	}

	if result, err := json.Marshal(u); err != nil {
		fmt.Printf("Json marshal failed, error: %v\n", err)
	} else {
		fmt.Println(string(result))
	}
}

func queryMulti(db *sql.DB) {
	fmt.Println("queryMulti")
	u := new(user)
	rows, err := db.Query("select * from users where id < ?", 3)
	if err != nil {
		fmt.Printf("Query failed, error: %v\n", err)
		return
	}
	defer func() {
		if rows != nil {
			rows.Close()
		}
	}()

	for rows.Next() {
		if err := rows.Scan(&u.ID, &u.Name, &u.NickName, &u.Password, &u.IsSuperUser, &u.Picture); err != nil {
			fmt.Printf("Scan failed, error: %v", err)
			return
		}

		if result, err := json.Marshal(u); err != nil {
			fmt.Printf("Json marshal failed, error: %v\n", err)
		} else {
			fmt.Println(string(result))
		}
	}

	if err := rows.Err(); err != nil {
		fmt.Println("Row error:", err)
	}
}

func updateRecord(db *sql.DB) {
	result, err := db.Exec("update users set nickname=? where id=?", "nick_go_new", 2)
	if err != nil {
		fmt.Printf("Update failed, error: %v\n", err)
		return
	}

	count, err := result.RowsAffected()
	if err != nil {
		fmt.Printf("Get RowsAffected failed, error: %v\n", err)
		return
	}
	fmt.Println("RowsAffected:", count)
}

func updateWithTransaction(db *sql.DB, isMockFail bool) {
	tx, err := db.Begin()
	if err != nil {
		fmt.Printf("Start a transaction failed, error: %v", err)
		return
	}
	defer tx.Rollback()

	result, err := tx.Exec("update users set nickname=? where id=?", "nick_tx_new2", 2)
	if err != nil {
		fmt.Println("Update failed, error:", err)
		if err = tx.Rollback(); err != nil {
			fmt.Println("Rollback failed, error:", err)
		}
		return
	}
	if count, err := result.RowsAffected(); err == nil {
		fmt.Println("RowsAffected:", count)
	}

	if isMockFail {
		fmt.Println("Mock failed and rollback.")
		time.Sleep(time.Second)
		if err = tx.Rollback(); err != nil {
			fmt.Println("Rollback failed, error:", err)
		}
	} else {
		if err := tx.Commit(); err != nil {
			if err = tx.Rollback(); err != nil {
				fmt.Println("Rollback failed, error:", err)
			}
		}
	}
}

func runSelect(db *sql.DB, sql string) {
	fmt.Println("Run sql:", sql)

	rows, err := db.Query(sql)
	if err != nil {
		fmt.Printf("Query failed, error: %v\n", err)
		return
	}
	defer func() {
		if rows != nil {
			rows.Close()
		}
	}()

	var res [][]string
	cols, err := rows.Columns()
	if err != nil {
		fmt.Println("Get columns failed, error:", err)
	}
	res = append(res, cols)

	row := make([]interface{}, len(cols))
	rowRef := make([]interface{}, len(cols)) // Scan()函数入参为指针类型
	for i := range row {
		rowRef[i] = &row[i]
	}

	for rows.Next() {
		if err := rows.Scan(rowRef...); err != nil {
			fmt.Printf("Scan failed, error: %v\n", err)
			return
		}
		tmp := make([]string, len(cols))
		for i := range row {
			if v, ok := row[i].([]byte); ok {
				if len(v) > 0 {
					tmp[i] = string(v)
				} else {
					tmp[i] = "null"
				}
			}
		}
		res = append(res, tmp)
	}

	if err = rows.Err(); err != nil {
		fmt.Println("Rows error:", err)
		return
	}

	for i := range res {
		fmt.Println(strings.Join(res[i], ","))
	}
}

func rawSQLMain() {
	db, err := buildDBConnection()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer db.Close()

	// queryOne(db)
	// queryMulti(db)

	sql := "select * from users where id < 4"
	runSelect(db, sql)

	// updateRecord(db)
	// updateWithTransaction(db, true)
}

/*
sqlx https://github.com/jmoiron/sqlx
*/

func sqlxMain() {
	dsn := fmt.Sprintf("%s:%s@%s(%s:%d)/%s", username, password, network, server, port, database)
	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	sql := "select * from users where id between 10 and 25"
	users := []user{}
	if err = db.Select(&users, sql); err != nil {
		panic(err)
	}
	fmt.Println("users:")
	for _, u := range users {
		fmt.Printf("%+v\n", u)
	}

	fmt.Println("\nusers:")
	rows, err := db.Queryx(sql)
	if err != nil {
		panic(err)
	}
	u := user{}
	for rows.Next() {
		if err := rows.StructScan(&u); err != nil {
			panic(err)
		}
		pic := "null"
		if len(u.Picture.String) > 0 {
			pic = u.Picture.String
		}
		fmt.Printf("name=%s, issuperuser=%s, pic=%s\n", u.Name, u.IsSuperUser, pic)
	}
}

func main() {
	// rawSQLMain()
	sqlxMain()

	fmt.Println("mysql app Done.")
}

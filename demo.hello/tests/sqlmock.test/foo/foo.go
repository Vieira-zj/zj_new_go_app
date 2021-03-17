package main

import (
	"database/sql"
	"fmt"
)

// Refer: https://github.com/DATA-DOG/go-sqlmock

func main() {
	db, err := sql.Open("mysql", "root@/blog")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	if err = recordStats(db, 1 /*some user id*/, 5 /*some product id*/); err != nil {
		panic(err)
	}
}

func recordStats(db *sql.DB, userID, productID int64) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		switch err {
		case nil:
			err = tx.Commit()
		default:
			tx.Rollback()
		}
	}()

	result, err := tx.Exec("UPDATE products SET views = views + 1")
	if err != nil {
		return err
	}
	logSQLResults(result)

	result, err = tx.Exec("INSERT INTO product_viewers (user_id, product_id) VALUES (?, ?)", userID, productID)
	if err != nil {
		return err
	}
	logSQLResults(result)
	return nil
}

func logSQLResults(result sql.Result) error {
	lastInsertID, err := result.LastInsertId()
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	fmt.Printf("LastInsertId=%d, RowsAffected=%d\n", lastInsertID, rowsAffected)
	return nil
}

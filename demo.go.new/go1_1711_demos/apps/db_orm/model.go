package db_orm

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"
	"time"

	"go1_1711_demo/apps/db_orm/generate"
)

// Model

type User struct {
	Id    int64     `json:"id"`
	Name  string    `json:"name"`
	Age   int64     `json:"age"`
	Ctime time.Time `json:"ctime"`
	Mtime time.Time `json:"mtime"`
}

// Dao

func FindUsersByRawSql(ctx context.Context, db *sql.DB) ([]*User, error) {
	query := "SELECT `id`,`name`,`age`,`ctime`,`mtime` FROM user WHERE `age`<? ORDER BY `id` LIMIT 20"
	rows, err := db.QueryContext(ctx, query, 30)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := []*User{}
	for rows.Next() {
		a := &User{}
		if err := rows.Scan(&a.Id, &a.Name, &a.Age, &a.Ctime, &a.Mtime); err != nil {
			return nil, err
		}
		results = append(results, a)
	}

	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return results, nil
}

func FindUserByReflect(ctx context.Context, db *sql.DB) ([]*User, error) {
	b := SelectBuilder{builder: &strings.Builder{}}
	sql, args := b.Select("id", "name", "age", "ctime", "mtime").From("user").
		Where(GT("id", 0), GT("age", 30)).OrderBy("id").Limit(0, 20).Query()

	rows, err := db.QueryContext(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := []*User{}
	if err := ScanDbRows(rows, &results); err != nil {
		return nil, err
	}

	return results, nil
}

func FindUser(ctx context.Context, db *sql.DB, b *SelectBuilder) ([]*User, error) {
	sql, args := b.Query()
	rows, err := db.QueryContext(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := []*User{}
	if colsDeepEqual(b.columns, generate.Columns) { // 不使用反射
		for rows.Next() {
			a := &User{}
			if err := rows.Scan(&a.Id, &a.Name, &a.Age, &a.Ctime, &a.Mtime); err != nil {
				return nil, err
			}
			results = append(results, a)
		}

		if rows.Err() != nil {
			return nil, rows.Err()
		}
		return results, nil
	}

	// 使用反射
	if err := ScanDbRows(rows, &results); err != nil {
		return nil, err
	}
	return results, nil
}

func colsDeepEqual(srcCols, dstCols []string) bool {
	if len(srcCols) != len(dstCols) {
		return false
	}

	sort.Strings(srcCols)
	sort.Strings(dstCols)
	fmt.Printf("columns deep equal: src=%v, dst=%v\n", srcCols, dstCols)
	for i := 0; i < len(srcCols); i++ {
		if srcCols[i] != dstCols[i] {
			return false
		}
	}
	return true
}

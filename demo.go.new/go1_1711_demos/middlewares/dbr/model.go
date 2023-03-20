package dbr

import "time"

type User struct {
	ID    int32     `db:"id" json:"id"`       // id
	Name  string    `db:"name" json:"name"`   // 名称
	Age   int32     `db:"age" json:"age"`     // 年龄
	Ctime time.Time `db:"ctime" json:"ctime"` // 创建时间
	Mtime time.Time `db:"mtime" json:"mtime"` // 更新时间
}

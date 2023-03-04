package generate

// Auto generated go table definition from "user" table ddl.
import "time"

type User struct {
	Id    int64     `json:"id"`    // id字段
	Name  string    `json:"name"`  // 名称
	Age   int64     `json:"age"`   // 年龄
	Ctime time.Time `json:"ctime"` // 创建时间
	Mtime time.Time `json:"mtime"` // 更新时间
}

const (
	table = "user"
	Id    = "id"
	Name  = "name"
	Age   = "age"
	Ctime = "ctime"
	Mtime = "mtime"
)

var Columns = []string{
	"id",
	"name",
	"age",
	"ctime",
	"mtime",
}

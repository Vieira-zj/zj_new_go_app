package db_orm

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"go/format"
	"io"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"text/template"
)

// SQL Parse for Table DDL

type Table struct {
	PackageName string
	TableName   string
	GoTableName string
	Fields      []*Column
}

type Column struct {
	ColumnName    string
	GoColumnName  string
	GoColumnType  string
	ColumnComment string
}

func GenTableDef(tabName string) error {
	tmplPath := filepath.Join(getCurrentDirPath(), "generate", tabName+"_table.tmpl")
	b, err := readFile(tmplPath)
	if err != nil {
		return err
	}

	parse, err := template.New(tabName + "_table").Parse(string(b))
	if err != nil {
		return err
	}

	buf := &bytes.Buffer{}
	if err := parse.Execute(buf, mockTableDDLParse()); err != nil {
		return err
	}

	b, err = format.Source(buf.Bytes())
	if err != nil {
		return err
	}
	outPath := filepath.Join(getCurrentDirPath(), "generate", tabName+"_table.go")
	return os.WriteFile(outPath, b, 0644)
}

func mockTableDDLParse() Table {
	// refer: SQL 语句解析器 https://github.com/xwb1989/sqlparser
	return Table{
		PackageName: "generate",
		TableName:   "user",
		GoTableName: "User",
		Fields: []*Column{
			{"id", "Id", "int64", "id字段"},
			{"name", "Name", "string", "名称"},
			{"age", "Age", "int64", "年龄"},
			{"ctime", "Ctime", "time.Time", "创建时间"},
			{"mtime", "Mtime", "time.Time", "更新时间"},
		},
	}
}

func getCurrentDirPath() string {
	_, fpath, _, _ := runtime.Caller(0)
	return path.Dir(fpath)
}

func readFile(fpath string) ([]byte, error) {
	input, err := os.Open(fpath)
	if err != nil {
		return nil, err
	}
	defer input.Close()

	return io.ReadAll(input)
}

// SQL Builder

type SelectBuilder struct {
	builder   *strings.Builder
	columns   []string
	tableName string
	where     []func(s *SelectBuilder)
	args      []interface{}
	orderby   string
	offset    *int64
	limit     *int64
}

func (s *SelectBuilder) Select(fields ...string) *SelectBuilder {
	s.columns = append(s.columns, fields...)
	return s
}

func (s *SelectBuilder) From(name string) *SelectBuilder {
	s.tableName = name
	return s
}

func (s *SelectBuilder) Where(f ...func(s *SelectBuilder)) *SelectBuilder {
	s.where = append(s.where, f...)
	return s
}

func (s *SelectBuilder) OrderBy(field string) *SelectBuilder {
	s.orderby = field
	return s
}

func (s *SelectBuilder) Limit(offset, limit int64) *SelectBuilder {
	s.offset = &offset
	s.limit = &limit
	return s
}

func GT(field string, arg interface{}) func(s *SelectBuilder) {
	return func(s *SelectBuilder) {
		s.builder.WriteString("`" + field + "`" + " > ?")
		s.args = append(s.args, arg)
	}
}

func (s *SelectBuilder) Query() (string, []interface{}) {
	s.builder.WriteString("SELECT ")
	for i, v := range s.columns {
		if i > 0 {
			s.builder.WriteString(",")
		}
		s.builder.WriteString("`" + v + "`")
	}

	s.builder.WriteString(" FROM ")
	s.builder.WriteString("`" + s.tableName + "` ")

	if len(s.where) > 0 {
		s.builder.WriteString("WHERE ")
		for i, f := range s.where {
			if i > 0 {
				s.builder.WriteString(" AND ")
			}
			f(s)
		}
	}

	if s.orderby != "" {
		s.builder.WriteString(" ORDER BY " + s.orderby)
	}
	if s.limit != nil {
		s.builder.WriteString(" LIMIT ")
		s.builder.WriteString(strconv.FormatInt(*s.limit, 10))
	}
	if s.offset != nil {
		s.builder.WriteString(" OFFSET ")
		s.builder.WriteString(strconv.FormatInt(*s.offset, 10))
	}

	return s.builder.String(), s.args
}

// Scanner

func ScanDbRows(rows *sql.Rows, dst interface{}) error {
	// 通过反射获取 dst slice element 结构体类型
	val := reflect.ValueOf(dst) // &[]*main.User
	if val.Kind() != reflect.Ptr {
		return errors.New("dst not a pointer")
	}
	val = reflect.Indirect(val) // []*main.User
	if val.Kind() != reflect.Slice {
		return errors.New("dst not a pointer to slice")
	}

	// 获取 slice 中的类型
	struPrt := val.Type().Elem() // &main.User
	stru := struPrt.Elem()       // main.User
	fmt.Println("dst slice element type:", stru)

	cols, err := rows.Columns()
	if err != nil {
		return err
	}
	if stru.NumField() < len(cols) {
		return errors.New("NumField and cols not match")
	}

	// 获取结构体中字段的类型和 tag
	tagIdx := make(map[string]int) //tag name -> field index
	for i := 0; i < stru.NumField(); i++ {
		tagname := stru.Field(i).Tag.Get("json")
		if tagname != "" {
			tagIdx[tagname] = i
		}
	}

	index := make([]int, 0, len(cols))               // [0,1,2,3,4,5]
	resultType := make([]reflect.Type, 0, len(cols)) // [int64,string,int64,time.Time,time.Time]
	for _, col := range cols {
		if idx, ok := tagIdx[col]; ok {
			index = append(index, idx)
			resultType = append(resultType, stru.Field(idx).Type)
		}
	}

	for rows.Next() {
		result := make([]interface{}, 0, len(resultType)) // []
		for _, t := range resultType {
			result = append(result, reflect.New(t).Interface())
		}
		if err := rows.Scan(result...); err != nil {
			return err
		}

		// 创建结构体对象
		obj := reflect.New(stru).Elem() // main.User
		for i, v := range result {
			fieldIdx := index[i]
			obj.Field(fieldIdx).Set(reflect.ValueOf(v).Elem()) // 给obj的每个字段赋值
		}

		vv := reflect.Append(val, obj.Addr())
		val.Set(vv) // []*main.User
	}

	return rows.Err()
}

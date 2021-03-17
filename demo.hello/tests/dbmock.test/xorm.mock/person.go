package xormmock

// Person struct for mock.
type Person struct {
	ID   int    `xorm:"pk id"`
	Name string `xorm:"name"`
}

// TableName returns table name.
func (*Person) TableName() string {
	return "person"
}

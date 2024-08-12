package model

type PriPersonModel struct {
	Name string
	age  int
}

func NewPriPersonModel(name string, age int) PriPersonModel {
	return PriPersonModel{
		Name: name,
		age:  age, // unexported field
	}
}

type PubPersonModel struct {
	Name string
	Age  int // exported field
}

func NewPubPersonModel(name string, age int) PubPersonModel {
	return PubPersonModel{
		Name: name,
		Age:  age,
	}
}

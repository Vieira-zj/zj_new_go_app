package main

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"go.uber.org/fx"
)

type t3 struct {
	Input io.Reader
}

func main() {
	type t1 struct {
		Name string
	}
	type t2 struct {
		Age int
	}

	var (
		v1 *t1
		v2 *t2
		v3 *t3
	)

	app := fx.New(
		fx.Provide(func() *t1 { return &t1{Name: "foo"} }),
		fx.Provide(func() *t2 { return &t2{Age: 31} }),
		fx.Provide(func() *t3 { return &t3{Input: strings.NewReader("hello world")} }),

		fx.Populate(&v1),
		fx.Populate(&v2),
		fx.Populate(&v3),
	)

	if err := app.Start(context.Background()); err != nil {
		panic(err)
	}
	defer app.Stop(context.Background())

	b, err := ioutil.ReadAll(v3.Input)
	if err != nil {
		panic(err)
	}
	fmt.Printf("the reulst is %v , %v\n", v1.Name, v2.Age)
	fmt.Printf("read string: %s\n", b)
}

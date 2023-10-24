package funcproxy

import (
	"fmt"
	"testing"
)

type iImpl interface {
	SayHello(string) string
}

// struct: app

type App struct {
	Impl iImpl `autowire:"impl"`
}

func (a App) Run() {
	fmt.Println("app start")
	a.Impl.SayHello("bar")
}

// struct: impl

type Impl struct {
	Name string
}

func (i Impl) SayHello(from string) string {
	msg := fmt.Sprintf("%s: hello, %s", from, i.Name)
	fmt.Println(msg)
	return msg
}

// struct: proxy

type Impl_ struct {
	// proxy fn to be inject: SayHello_ -> SayHello
	SayHello_ func(string) string
}

func (i Impl_) SayHello(from string) string {
	return i.SayHello_(from)
}

func TestProxyFunc(t *testing.T) {
	impl_ := &Impl_{}
	if err := ProxyFunc(&Impl{Name: "foo"}, impl_); err != nil {
		t.Fatal(err)
	}

	_ = impl_.SayHello("bar")
	t.Log("make proxy done")
}

func TestInject(t *testing.T) {
	app := &App{}
	// auto init
	RegisterImpl("app", app)
	RegisterImpl("impl", &Impl{Name: "foo"})
	RegisterProxyImpl("impl", &Impl_{})

	if err := Inject(); err != nil {
		t.Fatal(err)
	}

	app.Run()
	t.Log("inject done")
}

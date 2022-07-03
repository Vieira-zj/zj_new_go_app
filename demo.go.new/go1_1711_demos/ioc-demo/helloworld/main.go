package main

import (
	"fmt"
	"time"

	ioc "github.com/alibaba/ioc-golang"
)

// +ioc:autowire=true
// +ioc:autowire:type=singleton

type App struct {
	// inject main.ServiceImpl1 pointer to Service interface with proxy wrapper
	ServiceImpl1 Service `singleton:"main.ServiceImpl1"`

	// inject main.ServiceImpl2 pointer to Service interface with proxy wrapper
	ServiceImpl2 Service `singleton:"main.ServiceImpl2"`

	// inject ServiceImpl1 pointer to Service1 's own interface with proxy wrapper
	// this interface belongs to ServiceImpl1, there is no need to mark 'main.ServiceImpl1' in tag
	Service1OwnInterface ServiceImpl1IOCInterface `singleton:""`

	// inject ServiceStruct struct pointer
	ServiceStruct *ServiceStruct `singleton:""`
}

func (a *App) Run() {
	for {
		time.Sleep(time.Second * 3)
		fmt.Println(a.ServiceImpl1.GetHelloString("laurence"))
		fmt.Println(a.ServiceImpl2.GetHelloString("laurence"))

		fmt.Println(a.Service1OwnInterface.GetHelloString("laurence"))

		fmt.Println(a.ServiceStruct.GetString("laurence"))
	}
}

type Service interface {
	GetHelloString(string) string
}

// +ioc:autowire=true
// +ioc:autowire:type=singleton

type ServiceImpl1 struct {
}

func (s *ServiceImpl1) GetHelloString(name string) string {
	return fmt.Sprintf("This is ServiceImpl1, hello %s", name)
}

// +ioc:autowire=true
// +ioc:autowire:type=singleton

type ServiceImpl2 struct {
}

func (s *ServiceImpl2) GetHelloString(name string) string {
	return fmt.Sprintf("This is ServiceImpl2, hello %s", name)
}

// +ioc:autowire=true
// +ioc:autowire:type=singleton

type ServiceStruct struct {
}

func (s *ServiceStruct) GetString(name string) string {
	return fmt.Sprintf("This is ServiceStruct, hello %s", name)
}

func main() {
	// start
	if err := ioc.Load(); err != nil {
		panic(err)
	}

	// app, err := GetAppIOCInterfaceSingleton is ok too
	app, err := GetAppSingleton()
	if err != nil {
		panic(err)
	}
	app.Run()
}

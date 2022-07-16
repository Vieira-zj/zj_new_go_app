package main

import (
	"fmt"

	"github.com/alibaba/ioc-golang"
	iocConfig "github.com/alibaba/ioc-golang/config"
	"github.com/alibaba/ioc-golang/extension/config"
)

// +ioc:autowire=true
// +ioc:autowire:type=singleton
// +ioc:autowire:alias=AppAlias

type App struct {
	DemoConfigString  *config.ConfigString  `config:",autowire.config.demo-config.string-value"`
	DemoConfigInt     *config.ConfigInt     `config:",autowire.config.demo-config.int-value"`
	DemoConfigMap     *config.ConfigMap     `config:",autowire.config.demo-config.map-value"`
	DemoConfigSlice   *config.ConfigSlice   `config:",autowire.config.demo-config.slice-value"`
	DemoConfigInt64   *config.ConfigInt64   `config:",autowire.config.demo-config.int64-value"`
	DemoConfigFloat64 *config.ConfigFloat64 `config:",autowire.config.demo-config.float64-value"`
}

func (a *App) Run() {
	fmt.Println(a.DemoConfigString.Value())
	fmt.Println(a.DemoConfigInt.Value())
	fmt.Println(a.DemoConfigMap.Value())
	fmt.Println(a.DemoConfigSlice.Value())
	fmt.Println(a.DemoConfigInt64.Value())
	fmt.Println(a.DemoConfigFloat64.Value())
}

func main() {
	if err := ioc.Load(
		iocConfig.WithSearchPath("../conf"),
		iocConfig.WithConfigName("ioc_golang")); err != nil {
		panic(err)
	}

	getImplByFullName()
	getImplByAlias() // +ioc:autowire:alias=AppAlias
}

func getImplByFullName() {
	// Use the full name of the struct instead of App-App(${interfaceName}-${structName})
	app, err := GetAppSingleton()
	if err != nil {
		panic(err)
	}
	app.Run()
}

func getImplByAlias() {
	app, err := GetAppSingleton()
	if err != nil {
		panic(err)
	}
	app.Run()
}

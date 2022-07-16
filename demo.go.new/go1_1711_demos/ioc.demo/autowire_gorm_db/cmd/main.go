package main

import (
	"fmt"
	"time"

	sdkMysql "go1_1711_demo/ioc.demo/autowire_gorm_db/sdk"

	"github.com/alibaba/ioc-golang"
	"github.com/alibaba/ioc-golang/config"
)

// +ioc:autowire=true
// +ioc:autowire:type=singleton
// +ioc:autowire:alias=AppAlias

type App struct {
	MyDB sdkMysql.GORMDBIOCInterface `normal:",my-mysql"`
}

type MyDataDO struct {
	Id    int32
	Value string
}

func (a *MyDataDO) TableName() string {
	return "mydata"
}

func (a *App) Run() {
	if err := a.MyDB.AutoMigrate(&MyDataDO{}); err != nil {
		panic(err)
	}

	toInsertMyData := &MyDataDO{
		Value: "first value",
	}
	if err := a.MyDB.Model(&MyDataDO{}).Create(toInsertMyData).Error(); err != nil {
		panic(err)
	}

	for {
		time.Sleep(3 * time.Second)
		myDataDOs := make([]MyDataDO, 0)
		if err := a.MyDB.Model(&MyDataDO{}).Where("id = ?", 1).Find(&myDataDOs).Error(); err != nil {
			panic(err)
		}
		fmt.Println(myDataDOs)
	}
}

func main() {
	if err := ioc.Load(
		config.WithSearchPath("../conf"),
		config.WithConfigName("ioc_golang")); err != nil {
		panic(err)
	}

	app, err := GetAppSingleton()
	if err != nil {
		panic(err)
	}
	app.Run()
}

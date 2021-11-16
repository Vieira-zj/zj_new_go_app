package main

import (
	"log"

	"demo.hello/plugins/casbin/routers"
	"demo.hello/plugins/casbin/utils"
)

func init() {
	utils.InitDB()
	utils.InitEnforcer()
}

func main() {
	log.Fatal(routers.R.Run())
}

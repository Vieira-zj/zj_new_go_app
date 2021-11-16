package utils

import (
	"github.com/casbin/casbin"
	gormadapter "github.com/casbin/gorm-adapter"
)

/*
Refer: https://casbin.org/docs/en/adapters
*/

// Enforcer .
var Enforcer *casbin.Enforcer

// InitEnforcer .
func InitEnforcer() {
	adapter := gormadapter.NewAdapterByDB(Mysql)
	Enforcer = casbin.NewEnforcer("config/keymatch2_model.conf", adapter)
	Enforcer.EnableLog(true)
}

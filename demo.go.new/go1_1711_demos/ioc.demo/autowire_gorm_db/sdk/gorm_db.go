package sdk

import (
	"github.com/alibaba/ioc-golang/autowire"
	"gorm.io/gorm"
)

// +ioc:autowire=true
// +ioc:autowire:type=normal
// +ioc:autowire:type=singleton
// +ioc:autowire:paramType=Param
// +ioc:autowire:constructFunc=New

type GORMDB struct {
	client *gorm.DB
}

func fromDB(db *gorm.DB) GORMDBIOCInterface {
	return autowire.GetProxyFunction()(&GORMDB{
		client: db,
	}).(GORMDBIOCInterface)
}

func (db *GORMDB) AutoMigrate(dst ...interface{}) error {
	return db.client.AutoMigrate(dst...)
}

func (db *GORMDB) Model(value interface{}) GORMDBIOCInterface {
	return fromDB(db.client.Model(value))
}

func (db *GORMDB) Create(value interface{}) GORMDBIOCInterface {
	return fromDB(db.client.Create(value))
}

func (db *GORMDB) Where(query interface{}, args ...interface{}) GORMDBIOCInterface {
	return fromDB(db.client.Where(query, args...))
}

func (db *GORMDB) Find(dest interface{}, conds ...interface{}) GORMDBIOCInterface {
	return fromDB(db.client.Find(dest, conds...))
}

func (db *GORMDB) Error() error {
	return db.client.Error
}

package sdk

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Param struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
}

func (c *Param) New(mysqlImpl *GORMDB) (*GORMDB, error) {
	dbClient, err := gorm.Open(mysql.Open(getMysqlLinkStr(c)), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	mysqlImpl.client = dbClient
	return mysqlImpl, nil
}

func getMysqlLinkStr(conf *Param) string {
	return conf.Username + ":" + conf.Password + "@tcp(" + conf.Host + ":" + conf.Port + ")/" + conf.DBName +
		"?charset=utf8&parseTime=True&loc=Local"
}

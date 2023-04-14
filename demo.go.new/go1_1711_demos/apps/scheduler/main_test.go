package main

import (
	"testing"

	mgorm "go1_1711_demo/middlewares/gorm"
)

/*
CREATE TABLE `test_scheduler` (
  `id` int unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(64) NOT NULL,
  `payload` varchar(255) NOT NULL,
  `run_at` bigint unsigned NOT NULL COMMENT 'timestamp millis',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
*/

func TestCreateSchedulerTable(t *testing.T) {
	t.Skip("run once")
	db := mgorm.NewDB()
	m := DBModelScheduler{}
	if err := db.AutoMigrate(&m); err != nil {
		t.Fatal(err)
	}
	t.Log("create table:", m.TableName())
}

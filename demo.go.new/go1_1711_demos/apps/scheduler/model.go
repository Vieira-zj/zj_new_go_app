package main

type DBModelScheduler struct {
	ID      uint32 `gorm:"primaryKey;auto_increment;column:id" json:"id"`
	Name    string `gorm:"size:64;column:name;not null" json:"name"`
	Payload string `gorm:"size:256;column:payload;not null" json:"payload"`
	RunAt   uint64 `gorm:"column:run_at;not null;comment:timestamp millis" json:"run_at"`
}

func (DBModelScheduler) TableName() string {
	return "test_scheduler"
}

package client

import (
	"errors"
	"math/rand"
	"time"

	log "github.com/sirupsen/logrus"
)

// Connection network connection api interface.
type Connection interface {
	Get(interface{}) (string, error)
	Create(interface{}) (string, error)
	Update(interface{}) (string, error)
}

// MockConnect mock connect with random sleep.
type MockConnect struct {
	Total  int32
	Failed int32

	Sleep      int
	IsRandom   bool
	IsError    bool
	ErrPercent int
}

// Get select/retrieve data action.
func (c *MockConnect) Get(input interface{}) (string, error) {
	log.Debug("Mock api Get process ...")
	c.Total++
	if c.isError() {
		c.Failed++
		return "", errors.New("mock get error")
	}

	time.Sleep(time.Duration(c.wait()) * time.Millisecond)
	return "ok", nil
}

// Create insert data action.
func (c *MockConnect) Create(input interface{}) (string, error) {
	log.Debug("Mock api Create process ...")
	c.Total++
	time.Sleep(time.Duration(c.wait()) * time.Millisecond)
	return "ok", nil
}

// Update modify data action.
func (c *MockConnect) Update(input interface{}) (string, error) {
	log.Debug("Mock api Update process ...")
	c.Total++
	time.Sleep(time.Duration(c.wait()) * time.Millisecond)
	return "ok", nil
}

func (c *MockConnect) isError() bool {
	if !c.IsError {
		return false
	}
	num := rand.Intn(100)
	return num <= c.ErrPercent
}

func (c *MockConnect) wait() int {
	if c.IsRandom {
		return rand.Intn(c.Sleep)
	}
	return c.Sleep
}

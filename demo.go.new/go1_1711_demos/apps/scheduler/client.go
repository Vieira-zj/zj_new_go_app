package main

import (
	"context"
	"fmt"
	mgorm "go1_1711_demo/middlewares/gorm"
	"log"
	"time"

	"gorm.io/gorm"
)

type ListenFunc func(string)

type Listeners map[string]ListenFunc

type Event struct {
	ID      uint32 `gorm:"column:id" json:"id"`
	Name    string `gorm:"column:name" json:"name"`
	Payload string `gorm:"column:payload" json:"payload"`
}

type Scheduler struct {
	db        *gorm.DB
	listeners Listeners
}

func NewScheduler(listeners Listeners) Scheduler {
	return Scheduler{
		db:        mgorm.NewDB(),
		listeners: listeners,
	}
}

func (s Scheduler) Schedule(name string, payload string, runAt time.Time) error {
	log.Println("Scheduling event " + name + " to run at " + runAt.String())
	row := DBModelScheduler{
		Name:    name,
		Payload: payload,
		RunAt:   uint64(runAt.Unix()),
	}
	if err := s.db.WithContext(context.Background()).Create(&row).Error; err != nil {
		return fmt.Errorf("schedule insert error: %v", err)
	}

	return nil
}

func (s Scheduler) AddListener(name string, listenFunc ListenFunc) {
	s.listeners[name] = listenFunc
}

func (s Scheduler) CheckEventsInInterval(ctx context.Context, duration time.Duration) {
	go func() {
		ticker := time.NewTicker(duration)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				log.Println("Get sginal and scheduler exit:", ctx.Err())
				return
			case <-ticker.C:
				log.Println("Ticks Received...")
				events, err := s.checkDueEvents()
				if err != nil {
					log.Println("Scheduler get error:", err.Error())
				}
				for _, e := range events {
					s.callListeners(e)
				}
			}
		}
	}()
}

// checkDueEvents checks and returns due events.
func (s Scheduler) checkDueEvents() ([]Event, error) {
	var rows []Event
	err := s.db.Table(DBModelScheduler{}.TableName()).Where("run_at < ?", time.Now().Unix()).Find(&rows).Error
	return rows, err
}

// callListeners calls the event listener of provided event.
func (s Scheduler) callListeners(event Event) error {
	eventFn, ok := s.listeners[event.Name]
	if !ok {
		return fmt.Errorf("error: couldn't find event listeners attached to: %v", event.Name)
	}

	go eventFn(event.Payload)

	if err := s.db.Where("id = ?", event.ID).Delete(&DBModelScheduler{}); err != nil {
		return fmt.Errorf("error: delete due event: %v", err)
	}
	return nil
}

package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"time"
)

var eventListeners = Listeners{
	"SendEmail": func(s string) {
		log.Println("SendEmail:", s)
	},
	"PayBills": func(s string) {
		log.Println("PayBills:", s)
	},
}

// Demo: 构建一个基本的事件调度系统，它将在一定时间间隔后调度事件。

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)

	scheduler := NewScheduler(eventListeners)
	scheduler.CheckEventsInInterval(ctx, 5*time.Second)

	if err := scheduler.Schedule("SendEmail", "mail: gopher@gmail.com", time.Now().Add(15*time.Second)); err != nil {
		log.Fatalln(err)
	}
	if err := scheduler.Schedule("PayBills", "paybills: $4,000 bill", time.Now().Add(30*time.Second)); err != nil {
		log.Fatalln(err)
	}

	<-ctx.Done()
	cancel()

	log.Println("Interrupt received and closing...")
	time.Sleep(time.Second)
}

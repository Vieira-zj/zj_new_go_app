package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/fsnotify/fsnotify"
)

func main() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal("NewWatcher failed:", err)
	}
	// fsnotify使用了操作系统接口，监听器中保存了系统资源的句柄，所以使用后需要关闭
	defer watcher.Close()

	done := make(chan bool)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	go func() {
		<-quit
		log.Println("App is closing ...")
		close(done)
	}()

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				log.Printf("%s %s\n", event.Name, event.Op)
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	if err := watcher.Add("/tmp/test/"); err != nil {
		log.Fatal("Add failed:", err)
	}
	<-done
}

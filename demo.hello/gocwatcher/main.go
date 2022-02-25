package main

import "flag"

var (
	gocWatcherWorkingDir string
	help                 bool
)

func main() {
	flag.StringVar(&gocWatcherWorkingDir, "d", "/app/gocwatcher", "goc watch dog working dir.")
	flag.BoolVar(&help, "h", false, "help.")

	if help {
		flag.Usage()
		return
	}
}

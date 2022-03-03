package main

import "flag"

var (
	gocAdapterWorkingDir string
	help                 bool
)

func init() {
	flag.StringVar(&gocAdapterWorkingDir, "d", "/app/gocadapter", "goc watch dog working dir.")
	flag.BoolVar(&help, "h", false, "help.")
	flag.Parse()
}

func main() {
	if help {
		flag.Usage()
		return
	}
}

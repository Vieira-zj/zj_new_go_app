package main

import (
	"cmp"
	"log"
	"runtime/debug"
)

var AppVersion string

func main() {
	log.Println("app version:", cmp.Or(AppVersion, "beta"))
	logBuildInfo()

	log.Println("finish")
}

func logBuildInfo() {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		log.Println("no build info")
	}

	settings := make(map[string]string, len(info.Settings))
	for _, s := range info.Settings {
		settings[s.Key] = s.Value
	}

	log.Println("module:", info.Main.Path)
	log.Println("version:", info.Main.Version)
	log.Println("vcs.revision:", settings["vcs.revision"])
	log.Println("vcs.time:", settings["vcs.time"])
	log.Println("vcs.modified:", settings["vcs.modified"])
}

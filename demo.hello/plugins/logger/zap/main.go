package main

import (
	"encoding/json"

	"go.uber.org/zap"
)

func main() {
	rawJSON := []byte(`{
		"level":"debug",
		"encoding":"json",
		"outputPaths": ["stdout", "/tmp/test/server.log"],
		"errorOutputPaths": ["stderr"],
		"initialFields":{"name":"dj"},
		"encoderConfig": {
		  "messageKey": "message",
		  "levelKey": "level",
		  "levelEncoder": "lowercase"
		}
	  }`)

	// custom a logger
	var cfg zap.Config
	if err := json.Unmarshal(rawJSON, &cfg); err != nil {
		panic(err)
	}

	logger, err := cfg.Build()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	logger.Info("server start work successfully!")
}

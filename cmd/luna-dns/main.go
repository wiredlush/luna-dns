package main

import (
	"io"
	"log"
	"os"

	"github.com/natefinch/lumberjack"
	"github.com/wiredlush/luna-dns/internal/config"
	"github.com/wiredlush/luna-dns/internal/engine"
)

func main() {
	args := os.Args[1:]
	if len(args) <= 0 {
		log.Fatal("No configuration file provided")
	}
	config, err := config.Load(args[0])
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Configuration file loaded: " + args[0])

	if config.LogFile != "" {
		logWriter := io.MultiWriter(os.Stdout,
			&lumberjack.Logger{
				Filename:   config.LogFile,
				MaxSize:    250,
				MaxBackups: 2,
				MaxAge:     7,
			},
		)
		log.SetOutput(logWriter)
	}

	engine, err := engine.NewEngine(config)
	if err != nil {
		log.Fatal(err)
	}
	if err := engine.Start(); err != nil {
		log.Fatal(err)
	}
}

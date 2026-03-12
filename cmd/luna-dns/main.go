package main

import (
	"io"
	"log"
	"os"

	"github.com/natefinch/lumberjack"
	"github.com/wiredlush/luna-dns/pkg/config"
	"github.com/wiredlush/luna-dns/pkg/engine"
	"github.com/wiredlush/luna-dns/pkg/web"
)

func main() {
	args := os.Args[1:]

	if web.StartFunc != nil {
		if err := web.StartFunc(args); err != nil {
			log.Fatal(err)
		}
		select {}
	}

	if len(args) <= 0 {
		log.Fatal("No configuration file provided")
	}
	cfg, err := config.Load(args[0])
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Configuration file loaded: " + args[0])

	if cfg.LogFile != "" {
		logWriter := io.MultiWriter(os.Stdout,
			&lumberjack.Logger{
				Filename:   cfg.LogFile,
				MaxSize:    250,
				MaxBackups: 2,
				MaxAge:     7,
			},
		)
		log.SetOutput(logWriter)
	}

	eng, err := engine.NewEngine(cfg)
	if err != nil {
		log.Fatal(err)
	}
	if err := eng.Start(); err != nil {
		log.Fatal(err)
	}
}

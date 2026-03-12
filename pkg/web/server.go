//go:build web

package web

import (
	"flag"
	"log"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
)

var StartFunc func([]string) error

type WebConfig struct {
	Addr string
	Cert string
	Key  string
	DB   string
}

func startFromFlags(args []string) error {
	fs := flag.NewFlagSet("luna-dns-web", flag.ExitOnError)

	cfg := &WebConfig{}
	fs.StringVar(&cfg.Addr, "web-addr", ":8080", "Web server listen address")
	fs.StringVar(&cfg.Cert, "web-cert", "", "TLS certificate path")
	fs.StringVar(&cfg.Key, "web-key", "", "TLS key path")
	fs.StringVar(&cfg.DB, "db", "luna-dns.db", "SQLite database path")

	if err := fs.Parse(args); err != nil {
		return err
	}

	return Start(cfg)
}

func Start(cfg *WebConfig) error {
	certFile := cfg.Cert
	keyFile := cfg.Key

	if certFile == "" || keyFile == "" {
		var err error
		certFile, keyFile, err = generateSelfSignedCert()
		if err != nil {
			return err
		}
		log.Println("Generated self-signed certificate for HTTPS")
	}

	db, err := OpenDB(cfg.DB)
	if err != nil {
		return err
	}

	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	registerAPI(app, db)

	app.Use("/", filesystem.New(filesystem.Config{
		Root:       http.FS(staticFS),
		PathPrefix: "static",
		Browse:     false,
	}))

	go func() {
		log.Printf("Web server starting on https://%s", cfg.Addr)
		if err := app.ListenTLS(cfg.Addr, certFile, keyFile); err != nil {
			log.Printf("Web server error: %v", err)
		}
	}()

	return nil
}

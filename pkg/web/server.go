//go:build web

package web

import (
	"flag"
	"log"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/wiredlush/luna-dns/pkg/database"
)

var StartFunc func([]string) error

type API struct {
	Addr string
	Cert string
	Key  string
	DB   string
	db   *database.Database
	app  *fiber.App
}

func startFromFlags(args []string) error {
	fs := flag.NewFlagSet("luna-dns-web", flag.ExitOnError)

	api := &API{}
	fs.StringVar(&api.Addr, "web-addr", ":8080", "Web server listen address")
	fs.StringVar(&api.Cert, "web-cert", "", "TLS certificate path")
	fs.StringVar(&api.Key, "web-key", "", "TLS key path")
	fs.StringVar(&api.DB, "db", "luna-dns.db", "SQLite database path")

	if err := fs.Parse(args); err != nil {
		return err
	}

	return api.Start()
}

func (a *API) Start() error {
	certFile := a.Cert
	keyFile := a.Key

	if certFile == "" || keyFile == "" {
		var err error
		certFile, keyFile, err = generateSelfSignedCert()
		if err != nil {
			return err
		}
		log.Println("Generated self-signed certificate for HTTPS")
	}

	var err error
	a.db, err = database.Open(a.DB)
	if err != nil {
		return err
	}

	a.app = fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	a.app.Get("/api/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	a.app.Use("/", filesystem.New(filesystem.Config{
		Root:       http.FS(staticFS),
		PathPrefix: "static",
		Browse:     false,
	}))

	go func() {
		log.Printf("Web server starting on https://%s", a.Addr)
		if err := a.app.ListenTLS(a.Addr, certFile, keyFile); err != nil {
			log.Printf("Web server error: %v", err)
		}
	}()

	return nil
}

//go:build web

package web

import (
	"crypto/tls"
	"flag"
	"log"
	"net"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/wiredlush/luna-dns/pkg/database"
)

var StartFunc func([]string) error

type Server struct {
	Addr     string
	Cert     string
	Key      string
	DB       string
	db       *database.Database
	app      *fiber.App
	sessions *sessionStore
}

func startFromFlags(args []string) error {
	fs := flag.NewFlagSet("luna-dns-web", flag.ExitOnError)

	server := &Server{}
	fs.StringVar(&server.Addr, "web-addr", ":8080", "Web server listen address")
	fs.StringVar(&server.Cert, "web-cert", "", "TLS certificate path")
	fs.StringVar(&server.Key, "web-key", "", "TLS key path")
	fs.StringVar(&server.DB, "db", "luna-dns.db", "SQLite database path")

	if err := fs.Parse(args); err != nil {
		return err
	}

	return server.Start()
}

func (s *Server) loadTLSCert() (tls.Certificate, error) {
	if s.Cert != "" && s.Key != "" {
		return tls.LoadX509KeyPair(s.Cert, s.Key)
	}
	cert, err := generateSelfSignedCert()
	if err != nil {
		return cert, err
	}
	log.Println("Generated self-signed certificate for HTTPS")
	return cert, nil
}

func (s *Server) Start() error {
	tlsCert, err := s.loadTLSCert()
	if err != nil {
		return err
	}

	s.db, err = database.Open(s.DB)
	if err != nil {
		return err
	}

	s.sessions = newSessionStore()

	s.app = fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	s.app.Use(logger.New())

	s.app.Post("/api/login", s.handleLogin)
	s.app.Post("/api/logout", s.handleLogout)

	s.app.Use("/api", s.authMiddleware)
	s.app.Get("/api/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	s.app.Use("/", filesystem.New(filesystem.Config{
		Root:       http.FS(staticFS),
		PathPrefix: "static",
		Browse:     false,
	}))

	ln, err := net.Listen("tcp", s.Addr)
	if err != nil {
		return err
	}

	tlsLn := tls.NewListener(ln, &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
	})

	go func() {
		log.Printf("Web server starting on https://%s", s.Addr)
		if err := s.app.Listener(tlsLn); err != nil {
			log.Printf("Web server error: %v", err)
		}
	}()

	return nil
}

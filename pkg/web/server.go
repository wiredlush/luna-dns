//go:build web

package web

import (
	"crypto/tls"
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/wiredlush/luna-dns/pkg/config"
	"github.com/wiredlush/luna-dns/pkg/database"
	"github.com/wiredlush/luna-dns/pkg/engine"
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
	engine   *engine.Engine
	mu       sync.Mutex
}

func defaultDBPath() string {
	exe, err := os.Executable()
	if err != nil {
		return "luna-dns.db"
	}
	return filepath.Join(filepath.Dir(exe), "luna-dns.db")
}

func startFromFlags(args []string) error {
	fs := flag.NewFlagSet("luna-dns-web", flag.ExitOnError)

	server := &Server{}
	fs.StringVar(&server.Addr, "web-addr", ":8080", "Web server listen address")
	fs.StringVar(&server.Cert, "web-cert", "", "TLS certificate path")
	fs.StringVar(&server.Key, "web-key", "", "TLS key path")
	fs.StringVar(&server.DB, "db", defaultDBPath(), "SQLite database path")

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

	if dnsCfg, err := s.db.GetDnsConfig(); err == nil && dnsCfg.AutoStart {
		dnsServers := s.loadForwarders()
		cfg := &config.Config{
			Addr:     dnsCfg.ListenAddr(),
			Network:  dnsCfg.Network,
			CacheTTL: dnsCfg.CacheTTL,
			DNS:      dnsServers,
			Hosts:    s.loadRecords(),
		}
		eng, err := engine.NewEngine(cfg)
		if err != nil {
			log.Printf("Failed to create DNS engine: %v", err)
		} else if err := eng.StartBackground(); err != nil {
			log.Printf("Failed to start DNS engine: %v", err)
		} else {
			s.engine = eng
		}
	}

	s.app = fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	s.app.Use(logger.New())

	s.app.Post("/api/login", s.handleLogin)
	s.app.Post("/api/logout", s.handleLogout)

	s.app.Use("/api", s.authMiddleware)
	s.app.Get("/api/status", s.handleStatus)

	s.app.Get("/api/users", s.listUsers)
	s.app.Post("/api/users", s.createUser)
	s.app.Delete("/api/users/:id", s.deleteUser)
	s.app.Post("/api/change-password", s.changePassword)

	s.app.Get("/api/sessions", s.listSessions)
	s.app.Post("/api/sessions/logout-all", s.logoutAll)
	s.app.Get("/api/audit-logs", s.listAuditLogs)

	s.app.Get("/api/dns/config", s.getDnsConfig)
	s.app.Post("/api/dns/config", s.saveDnsConfig)
	s.app.Post("/api/dns/start", s.startDns)
	s.app.Post("/api/dns/stop", s.stopDns)
	s.app.Post("/api/dns/restart", s.restartDns)

	s.app.Get("/api/dns/forwarders", s.listForwarders)
	s.app.Post("/api/dns/forwarders", s.createForwarder)
	s.app.Delete("/api/dns/forwarders/:id", s.deleteForwarder)

	s.app.Get("/api/dns/records", s.listRecords)
	s.app.Post("/api/dns/records", s.createRecord)
	s.app.Delete("/api/dns/records/:id", s.deleteRecord)

	s.app.Use("/", filesystem.New(filesystem.Config{
		Root:       http.FS(staticFS),
		PathPrefix: "static",
		Browse:     false,
	}))

	s.app.Use(func(c *fiber.Ctx) error {
		data, err := staticFS.ReadFile("static/index.html")
		if err != nil {
			return c.SendStatus(fiber.StatusNotFound)
		}
		c.Set("Content-Type", "text/html")
		return c.Send(data)
	})

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

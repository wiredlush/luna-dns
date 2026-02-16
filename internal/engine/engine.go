package engine

import (
	"fmt"
	"log"
	"time"

	"github.com/miekg/dns"
	"github.com/wiredlush/luna-dns/internal/blocklists"
	"github.com/wiredlush/luna-dns/internal/cache"
	"github.com/wiredlush/luna-dns/internal/config"
	"github.com/wiredlush/luna-dns/internal/entry"
	"github.com/wiredlush/luna-dns/internal/tree"
)

type Engine struct {
	Hosts        *tree.Tree
	Blocklists   *blocklists.Blocklists
	cache        *cache.Cache
	addr         string
	network      string
	dns          []config.DNS
	forwardIndex int
	server       *dns.Server
	cfg          *config.Config
}

func NewEngine(cfg *config.Config) (*Engine, error) {
	Hosts := tree.NewTree()
	for _, host := range cfg.Hosts {
		entry, err := entry.NewEntry(host.Host, host.IP)
		if err != nil {
			return nil, err
		}

		Hosts.Insert(entry)
	}

	return &Engine{
		Hosts:        Hosts,
		Blocklists:   blocklists.NewBlocklists(cfg.Blocklists, cfg.BlocklistUpdate),
		cache:        cache.NewCache(time.Duration(cfg.CacheTTL) * time.Second),
		addr:         cfg.Addr,
		network:      cfg.Network,
		dns:          cfg.DNS,
		forwardIndex: 0,
		cfg:          cfg,
	}, nil
}

func (e *Engine) Start() error {
	if e.server != nil {
		return fmt.Errorf("engine already running")
	}

	e.Blocklists = blocklists.NewBlocklists(e.cfg.Blocklists, e.cfg.BlocklistUpdate)
	e.cache = cache.NewCache(time.Duration(e.cfg.CacheTTL) * time.Second)

	go e.Blocklists.Routine()
	go e.cache.Routine()

	log.Printf("Listening on %s (%s)\n", e.addr, e.network)

	dns.HandleFunc(".", e.handler)
	e.server = &dns.Server{Addr: e.addr, Net: e.network}
	return e.server.ListenAndServe()
}

func (e *Engine) Stop() error {
	if e.server == nil {
		return fmt.Errorf("engine not running")
	}

	log.Println("Stopping engine...")

	e.Blocklists.Stop()
	e.cache.Stop()

	err := e.server.Shutdown()
	e.server = nil

	log.Println("Engine stopped")
	return err
}

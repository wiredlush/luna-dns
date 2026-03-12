package engine

import (
	"fmt"
	"log"
	"time"

	"github.com/miekg/dns"
	"github.com/wiredlush/luna-dns/pkg/cache"
	"github.com/wiredlush/luna-dns/pkg/config"
	"github.com/wiredlush/luna-dns/pkg/entry"
	"github.com/wiredlush/luna-dns/pkg/tree"
)

type Engine struct {
	hostTree        *tree.Tree
	blocklistTree   *tree.Tree
	blocklists      []string
	blocklistUpdate int64
	cache           *cache.Cache
	addr            string
	network         string
	dns             []config.DNS
	forwardIndex    int
	server          *dns.Server
}

func NewEngine(config *config.Config) (*Engine, error) {
	hosts := tree.NewTree()
	for _, host := range config.Hosts {
		entry, err := entry.NewEntry(host.Host, host.IP)
		if err != nil {
			return nil, err
		}
		hosts.Insert(entry)
	}

	blockListUpdate := config.BlocklistUpdate
	if blockListUpdate == 0 {
		blockListUpdate = 720
	}

	return &Engine{
		hostTree:        hosts,
		blocklistTree:   tree.NewTree(),
		blocklists:      config.Blocklists,
		blocklistUpdate: blockListUpdate,
		cache:           cache.NewCache(time.Duration(config.CacheTTL) * time.Second),
		addr:            config.Addr,
		network:         config.Network,
		dns:             config.DNS,
		forwardIndex:    0,
	}, nil
}

func (e *Engine) Start() error {
	if e.server != nil {
		return fmt.Errorf("engine is already running")
	}

	e.cache.Reset()
	go e.BlocklistsRoutine()
	go e.cache.CacheRoutine()

	log.Printf("Listening on %s (%s)\n", e.addr, e.network)

	dns.HandleFunc(".", e.handler)
	e.server = &dns.Server{Addr: e.addr, Net: e.network}
	return e.server.ListenAndServe()
}

func (e *Engine) Stop() error {
	if e.server == nil {
		return fmt.Errorf("engine is not running")
	}
	e.cache.Stop()
	err := e.server.Shutdown()
	e.server = nil
	return err
}

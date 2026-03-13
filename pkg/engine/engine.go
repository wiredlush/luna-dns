package engine

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/miekg/dns"
	"github.com/wiredlush/luna-dns/pkg/cache"
	"github.com/wiredlush/luna-dns/pkg/config"
	"github.com/wiredlush/luna-dns/pkg/entry"
	"github.com/wiredlush/luna-dns/pkg/tree"
)

type Engine struct {
	hostMu          sync.RWMutex
	hostTree        *tree.Tree
	blocklistTree   *tree.Tree
	blocklists      []string
	blocklistUpdate int64
	cache           *cache.Cache
	addr            string
	network         string
	dnsMu           sync.RWMutex
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

	cacheTTL := config.CacheTTL
	if cacheTTL == 0 {
		cacheTTL = 14400
	}

	return &Engine{
		hostTree:        hosts,
		blocklistTree:   tree.NewTree(),
		blocklists:      config.Blocklists,
		blocklistUpdate: blockListUpdate,
		cache:           cache.NewCache(time.Duration(cacheTTL) * time.Second),
		addr:            config.Addr,
		network:         config.Network,
		dns:             config.DNS,
		forwardIndex:    0,
	}, nil
}

func (e *Engine) Running() bool {
	return e.server != nil
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

func (e *Engine) StartBackground() error {
	if e.server != nil {
		return fmt.Errorf("engine is already running")
	}

	e.cache.Reset()
	go e.BlocklistsRoutine()
	go e.cache.CacheRoutine()

	dns.HandleFunc(".", e.handler)

	started := make(chan struct{})
	e.server = &dns.Server{
		Addr: e.addr,
		Net:  e.network,
		NotifyStartedFunc: func() {
			close(started)
		},
	}

	errCh := make(chan error, 1)
	go func() {
		if err := e.server.ListenAndServe(); err != nil {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		e.server = nil
		log.Printf("Failed to listen on %s (%s): %v", e.addr, e.network, err)
		return err
	case <-started:
		log.Printf("Listening on %s (%s)\n", e.addr, e.network)
		return nil
	}
}

func (e *Engine) SetForwarders(forwarders []config.DNS) {
	e.dnsMu.Lock()
	defer e.dnsMu.Unlock()

	old := make(map[string]bool, len(e.dns))
	for _, d := range e.dns {
		old[d.Addr+"|"+d.Network] = true
	}
	new := make(map[string]bool, len(forwarders))
	for _, d := range forwarders {
		new[d.Addr+"|"+d.Network] = true
	}

	for _, d := range forwarders {
		if !old[d.Addr+"|"+d.Network] {
			log.Printf("Forwarder added: %s (%s)", d.Addr, d.Network)
		}
	}
	for _, d := range e.dns {
		if !new[d.Addr+"|"+d.Network] {
			log.Printf("Forwarder removed: %s (%s)", d.Addr, d.Network)
		}
	}

	e.dns = forwarders
	e.forwardIndex = 0
}

func (e *Engine) SetHosts(hosts []config.Host) {
	newTree := tree.NewTree()
	for _, h := range hosts {
		ent, err := entry.NewEntry(h.Host, h.IP)
		if err != nil {
			log.Printf("Invalid host entry %s -> %s: %v", h.Host, h.IP, err)
			continue
		}
		newTree.Insert(ent)
	}

	e.hostMu.Lock()
	e.hostTree = newTree
	e.hostMu.Unlock()

	log.Printf("DNS records updated: %d entries", len(hosts))
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

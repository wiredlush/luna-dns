package engine

import (
	"sync"
	"sync/atomic"
)

type Stats struct {
	TotalQueries   atomic.Uint64
	BlockedQueries atomic.Uint64
	CustomQueries  atomic.Uint64
	CacheHits      atomic.Uint64
	CacheMisses    atomic.Uint64
	clients        sync.Map // key: string (IP), value: struct{}
	domains        sync.Map // key: string (domain), value: struct{}
}

type StatsSnapshot struct {
	TotalQueries   uint64  `json:"total_queries"`
	BlockedQueries uint64  `json:"blocked_queries"`
	CustomQueries  uint64  `json:"custom_queries"`
	CacheHits      uint64  `json:"cache_hits"`
	CacheMisses    uint64  `json:"cache_misses"`
	CacheHitRate   float64 `json:"cache_hit_rate"`
	UniqueClients  uint64  `json:"unique_clients"`
	UniqueDomains  uint64  `json:"unique_domains"`
}

func (s *Stats) TrackClient(ip string) {
	s.clients.Store(ip, struct{}{})
}

func (s *Stats) TrackDomain(domain string) {
	s.domains.Store(domain, struct{}{})
}

func (s *Stats) Snapshot() StatsSnapshot {
	var clients, domains uint64
	s.clients.Range(func(_, _ any) bool {
		clients++
		return true
	})
	s.domains.Range(func(_, _ any) bool {
		domains++
		return true
	})

	hits := s.CacheHits.Load()
	misses := s.CacheMisses.Load()
	var hitRate float64
	if total := hits + misses; total > 0 {
		hitRate = float64(int(float64(hits)/float64(total)*1000)) / 10
	}

	return StatsSnapshot{
		TotalQueries:   s.TotalQueries.Load(),
		BlockedQueries: s.BlockedQueries.Load(),
		CustomQueries:  s.CustomQueries.Load(),
		CacheHits:      hits,
		CacheMisses:    misses,
		CacheHitRate:   hitRate,
		UniqueClients:  clients,
		UniqueDomains:  domains,
	}
}

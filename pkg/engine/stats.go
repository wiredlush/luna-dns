package engine

import (
	"sync"
	"sync/atomic"
	"time"
)

const (
	timeSeriesBuckets  = 144 // 24h / 10min = 144 buckets
	timeSeriesInterval = 10 * time.Minute
)

type Stats struct {
	TotalQueries   atomic.Uint64
	BlockedQueries atomic.Uint64
	CustomQueries  atomic.Uint64
	CacheHits      atomic.Uint64
	CacheMisses    atomic.Uint64
	clients        sync.Map
	domains        sync.Map
	tsBuckets      [timeSeriesBuckets]tsBucketAtomic
	tsMu           sync.Mutex // only held by Snapshot, never by hot path
	tsLastAdvance  int64      // unix timestamp of last advance
}

type tsBucketAtomic struct {
	Total   atomic.Uint64
	Blocked atomic.Uint64
	Custom  atomic.Uint64
}

type TimeSeriesPoint struct {
	Time    string `json:"time"`
	Total   uint64 `json:"total"`
	Blocked uint64 `json:"blocked"`
	Custom  uint64 `json:"custom"`
}

type StatsSnapshot struct {
	TotalQueries   uint64            `json:"total_queries"`
	BlockedQueries uint64            `json:"blocked_queries"`
	CustomQueries  uint64            `json:"custom_queries"`
	CacheHits      uint64            `json:"cache_hits"`
	CacheMisses    uint64            `json:"cache_misses"`
	CacheHitRate   float64           `json:"cache_hit_rate"`
	UniqueClients  uint64            `json:"unique_clients"`
	UniqueDomains  uint64            `json:"unique_domains"`
	TimeSeries     []TimeSeriesPoint `json:"time_series"`
}

func (s *Stats) TrackClient(ip string) {
	s.clients.Store(ip, struct{}{})
}

func (s *Stats) TrackDomain(domain string) {
	s.domains.Store(domain, struct{}{})
}

func bucketIndex(t time.Time) int {
	return int(t.Unix()/int64(timeSeriesInterval.Seconds())) % timeSeriesBuckets
}

// RecordQuery is lock-free — called on every DNS query
func (s *Stats) RecordQuery() {
	s.tsBuckets[bucketIndex(time.Now())].Total.Add(1)
}

// RecordBlocked is lock-free — called when a query is blocked
func (s *Stats) RecordBlocked() {
	s.tsBuckets[bucketIndex(time.Now())].Blocked.Add(1)
}

// RecordCustom is lock-free — called when a query matches a custom record
func (s *Stats) RecordCustom() {
	s.tsBuckets[bucketIndex(time.Now())].Custom.Add(1)
}

// advanceAndCollect clears stale buckets and returns the time series.
// Called only from Snapshot() (every ~2s), holds the mutex briefly.
func (s *Stats) advanceAndCollect() []TimeSeriesPoint {
	now := time.Now()
	currentIdx := bucketIndex(now)

	s.tsMu.Lock()
	defer s.tsMu.Unlock()

	lastAdvance := s.tsLastAdvance
	nowUnix := now.Unix()
	s.tsLastAdvance = nowUnix

	if lastAdvance > 0 {
		lastIdx := int(lastAdvance/int64(timeSeriesInterval.Seconds())) % timeSeriesBuckets
		if lastIdx != currentIdx {
			steps := currentIdx - lastIdx
			if steps < 0 {
				steps += timeSeriesBuckets
			}
			if steps > timeSeriesBuckets {
				steps = timeSeriesBuckets
			}
			for i := 1; i <= steps; i++ {
				clearIdx := (lastIdx + i) % timeSeriesBuckets
				s.tsBuckets[clearIdx].Total.Store(0)
				s.tsBuckets[clearIdx].Blocked.Store(0)
				s.tsBuckets[clearIdx].Custom.Store(0)
			}
		}
	}

	points := make([]TimeSeriesPoint, timeSeriesBuckets)
	for i := range timeSeriesBuckets {
		bIdx := (currentIdx + 1 + i) % timeSeriesBuckets
		bucketTime := now.Add(-time.Duration(timeSeriesBuckets-1-i) * timeSeriesInterval).Truncate(timeSeriesInterval)
		points[i] = TimeSeriesPoint{
			Time:    bucketTime.Format("15:04"),
			Total:   s.tsBuckets[bIdx].Total.Load(),
			Blocked: s.tsBuckets[bIdx].Blocked.Load(),
			Custom:  s.tsBuckets[bIdx].Custom.Load(),
		}
	}

	return points
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
		TimeSeries:     s.advanceAndCollect(),
	}
}

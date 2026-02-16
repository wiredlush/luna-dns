package cache

import (
	"log"
	"sync"
	"time"

	"github.com/miekg/dns"
)

type Cache struct {
	sync.Mutex
	entries map[string]entry
	ttl     time.Duration
	stopCh  chan struct{}
}

type entry struct {
	createdAt time.Time
	answer    []dns.RR
}

func NewCache(ttl time.Duration) *Cache {
	return &Cache{
		ttl:     ttl,
		entries: map[string]entry{},
		stopCh:  make(chan struct{}),
	}
}

func (c *Cache) Stop() {
	close(c.stopCh)
}

func (c *Cache) Search(question []dns.Question) []dns.RR {
	c.Lock()
	defer c.Unlock()
	return c.entries[hashQuestion(question)].answer
}

func (c *Cache) Insert(question []dns.Question, answer []dns.RR) {
	c.Lock()
	defer c.Unlock()

	hash := hashQuestion(question)
	c.entries[hash] = entry{
		createdAt: time.Now(),
		answer:    answer,
	}

	log.Println("New entry in cache: " + hash)
}

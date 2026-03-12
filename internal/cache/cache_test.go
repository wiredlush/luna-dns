package cache

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/miekg/dns"
)

func TestCache(t *testing.T) {
	cache := NewCache(1 * time.Millisecond)

	rr, _ := dns.NewRR(fmt.Sprintf("%s A %s", "test", "127.0.0.1"))
	question := dns.Question{
		Name:   "test",
		Qtype:  1,
		Qclass: 1,
	}
	cache.Insert([]dns.Question{question}, []dns.RR{rr})
	found := cache.Search([]dns.Question{question})
	if found == nil {
		t.Fatal()
	}
	if !reflect.DeepEqual(found[0], rr) {
		t.Fatal()
	}

	time.Sleep(1 * time.Millisecond)
	deleted := cache.deleteOldEntries()
	if deleted == 0 {
		t.Fatal()
	}
}

func TestCacheRoutine(t *testing.T) {
	cache := NewCache(1 * time.Second)
	go cache.CacheRoutine()

	cache.Insert([]dns.Question{}, []dns.RR{})

	time.Sleep(5 * time.Second)

	cache.Lock()
	if len(cache.entries) > 0 {
		t.Fatal()
	}
	cache.Unlock()
}

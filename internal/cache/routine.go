package cache

import (
	"log"
	"time"
)

func (c *Cache) Routine() {
	for {
		time.Sleep(c.ttl + (1 * time.Second))
		log.Println("Cleaning old cache entries...")

		deletedEntries := c.deleteOldEntries()
		if deletedEntries > 0 {
			log.Printf("Deleted %d entries from cache\n", deletedEntries)
		}
	}
}

func (c *Cache) deleteOldEntries() int {
	c.Lock()
	defer c.Unlock()

	deletedEntries := 0
	for hash, entry := range c.entries {
		delta := time.Since(entry.createdAt)
		if delta > c.ttl {
			delete(c.entries, hash)
			deletedEntries++
		}
	}

	return deletedEntries
}

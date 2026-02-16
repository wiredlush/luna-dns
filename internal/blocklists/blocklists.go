package blocklists

import (
	"log"
	"strings"
	"time"

	"github.com/wiredlush/luna-dns/internal/tree"
)

type Blocklists struct {
	hosts      *tree.Tree
	blocklists []string
	updateTime int64
	stopCh     chan struct{}
}

func NewBlocklists(blocklists []string, updateTime int64) *Blocklists {
	if updateTime == 0 {
		updateTime = 720
	}

	return &Blocklists{
		hosts:      tree.NewTree(),
		blocklists: blocklists,
		updateTime: updateTime,
		stopCh:     make(chan struct{}),
	}
}

func (b *Blocklists) Stop() {
	close(b.stopCh)
}

func (b *Blocklists) Routine() {
	if len(b.blocklists) == 0 {
		return
	}

	b.update()

	timer := time.NewTimer(time.Duration(b.updateTime) * time.Minute)
	defer timer.Stop()

	for {
		select {
		case <-b.stopCh:
			return
		case <-timer.C:
			b.update()
			timer.Reset(time.Duration(b.updateTime) * time.Minute)
		}
	}
}

func (b *Blocklists) update() {
	log.Println("Updating blocklists...")

	newHosts := tree.NewTree()
	for _, blocklist := range b.blocklists {
		if strings.HasPrefix(blocklist, "file://") {
			b.processFile(blocklist, newHosts)
			continue
		}

		b.processRemote(blocklist, newHosts)
	}

	b.hosts = newHosts
	log.Printf("Blocklists updated, next update in %d minutes\n", b.updateTime)
}

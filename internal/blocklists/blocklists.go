package blocklists

import (
	"log"
	"strings"
	"time"

	"github.com/wiredlush/luna-dns/internal/tree"
)

// Blocklists - Blocklists strutct
type Blocklists struct {
	hosts      *tree.Tree
	blocklists []string
	updateTime int64
}

// NewBlocklists - Create a new Blocklists
func NewBlocklists(blocklists []string, updateTime int64) *Blocklists {
	if updateTime == 0 {
		updateTime = 720
	}

	return &Blocklists{
		hosts:      tree.NewTree(),
		blocklists: blocklists,
		updateTime: updateTime,
	}
}

// Routine - Start Blocklists update routine
func (b *Blocklists) Routine() {
	if len(b.blocklists) == 0 {
		return
	}

	for {
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
		log.Printf("Blocklists updated, next update in %d minutes\n",
			b.updateTime)

		time.Sleep(time.Duration(b.updateTime * int64(time.Minute)))
	}
}

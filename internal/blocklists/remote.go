package blocklists

import (
	"bufio"
	"log"
	"net/http"
	"strings"

	"github.com/wiredlush/luna-dns/internal/entry"
	"github.com/wiredlush/luna-dns/internal/tree"
)

func (b *Blocklists) processRemote(blocklist string, hosts *tree.Tree) {
	log.Printf("Downloading %s...\n", blocklist)

	resp, err := http.Get(blocklist)
	if err != nil {
		log.Printf("Unable to download remote blocklist %s: %s\n",
			blocklist, err)
		return
	}
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		entry, err := entry.NewEntry(line, "0.0.0.0")
		if err != nil {
			continue
		}
		hosts.Insert(entry)
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error processing blocklist %s: %s",
			blocklist, err)
		return
	}
}

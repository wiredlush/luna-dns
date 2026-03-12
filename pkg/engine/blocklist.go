package engine

import (
	"bufio"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/wiredlush/luna-dns/pkg/entry"
	"github.com/wiredlush/luna-dns/pkg/tree"
)

func (e *Engine) BlocklistsRoutine() {
	if len(e.blocklists) == 0 {
		return
	}

	for {
		log.Println("Updating blocklists...")

		newHosts := tree.NewTree()
		for _, blocklist := range e.blocklists {
			if strings.HasPrefix(blocklist, "file://") {
				e.processFile(blocklist, newHosts)
				continue
			}
			e.processRemote(blocklist, newHosts)
		}
		e.hostTree = newHosts

		log.Printf("Blocklists updated, next update in %d minutes\n", e.blocklistUpdate)
		time.Sleep(time.Duration(e.blocklistUpdate * int64(time.Minute)))
	}
}

func (e *Engine) processFile(blocklist string, blocklistTree *tree.Tree) {
	filepath := strings.TrimPrefix(blocklist, "file://")
	log.Printf("Processing blocklist file %s...\n", filepath)

	file, err := os.Open(filepath)
	if err != nil {
		log.Printf("Error opening blocklist file %s: %s", filepath, err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		entry, err := entry.NewEntry(line, "0.0.0.0")
		if err != nil {
			continue
		}
		blocklistTree.Insert(entry)
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error processing blocklist %s: %s",
			filepath, err)
		return
	}
}

func (e *Engine) processRemote(blocklist string, hosts *tree.Tree) {
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

package blocklists

import (
	"bufio"
	"log"
	"os"
	"strings"

	"github.com/wiredlush/luna-dns/internal/entry"
	"github.com/wiredlush/luna-dns/internal/tree"
)

func (b *Blocklists) processFile(blocklist string, hosts *tree.Tree) {
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
		hosts.Insert(entry)
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error processing blocklist %s: %s",
			filepath, err)
		return
	}
}

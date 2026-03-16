package engine

import (
	"bufio"
	"log"
	"os"

	"github.com/wiredlush/luna-dns/pkg/blocklist"
	"github.com/wiredlush/luna-dns/pkg/entry"
	"github.com/wiredlush/luna-dns/pkg/tree"
)

func (e *Engine) loadBlocklists() {
	if len(e.blocklists) == 0 {
		return
	}

	log.Println("Loading blocklists...")

	newTree := tree.NewTree()
	for _, path := range e.blocklists {
		e.processFile(path, newTree)
	}
	e.blocklistTree = newTree

	log.Println("Blocklists loaded")
}

func (e *Engine) processFile(path string, blocklistTree *tree.Tree) {
	log.Printf("Processing blocklist file %s...\n", path)

	file, err := os.Open(path)
	if err != nil {
		log.Printf("Error opening blocklist file %s: %s", path, err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		domain := blocklist.ParseLine(scanner.Text())
		if domain == "" {
			continue
		}
		ent, err := entry.NewEntry(domain, "0.0.0.0")
		if err != nil {
			continue
		}
		blocklistTree.Insert(ent)
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error processing blocklist %s: %s", path, err)
	}
}

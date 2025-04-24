package tree

import (
	"github.com/wiredlush/luna-dns/internal/entry"
)

// Search - Search for a domain in DNS tree
func (t *Tree) Search(domain string) (string, error) {
	entry, err := entry.NewEntry(domain, "")
	if err != nil {
		return "", err
	}

	return t.searchEntry(entry), nil
}

func (t *Tree) searchEntry(entry *entry.Entry) string {
	foundTLD, wildcard := searchNode(&t.tlds, entry.TLD)
	if wildcard {
		return foundTLD.ip
	}

	if foundTLD != nil {
		current := foundTLD
		for index, subdomain := range entry.Subdomains {
			foundNode, wildcard := searchNode(&current.childrens, subdomain)
			switch {
			case wildcard:
				return foundNode.ip
			case foundNode == nil:
				return ""
			case index == len(entry.Subdomains)-1:
				return foundNode.ip
			}
			current = foundNode
		}
	}

	return ""
}

func searchNode(nodes *map[string]*node, key string) (*node, bool) {
	wildcardNode, ok := (*nodes)["*"]
	if ok {
		return wildcardNode, true
	}

	node, ok := (*nodes)[key]
	if !ok {
		return nil, false
	}

	return node, false
}

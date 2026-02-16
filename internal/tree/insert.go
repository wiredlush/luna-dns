package tree

import (
	"github.com/wiredlush/luna-dns/internal/entry"
)

func (t *Tree) Insert(entry *entry.Entry) {
	foundTLD, _ := searchNode(&t.tlds, entry.TLD)
	if foundTLD == nil {
		switch entry.TLD {
		case "*":
			foundTLD = t.insertNode(&t.tlds, entry.TLD, entry.IP)
		default:
			foundTLD = t.insertNode(&t.tlds, entry.TLD, "")
		}
	}

	current := foundTLD
	for i, subdomain := range entry.Subdomains {
		if subdomain != "*" && i != len(entry.Subdomains)-1 {
			foundNode, _ := searchNode(&current.children, subdomain)
			if foundNode == nil {
				foundNode = t.insertNode(&current.children, subdomain, "")
			}
			current = foundNode
			continue
		}

		foundNode, _ := searchNode(&current.children, subdomain)
		if foundNode == nil {
			foundNode = t.insertNode(&current.children, subdomain, entry.IP)
		}
		current = foundNode

		if subdomain == "*" {
			break
		}
	}
}

func (t *Tree) insertNode(nodes *map[string]*node, host string, ip string) *node {
	(*nodes)[host] = &node{
		ip:       ip,
		children: map[string]*node{},
	}

	insertedNode, _ := searchNode(nodes, host)
	return insertedNode
}

package tree

import (
	"github.com/wiredlush/luna-dns/pkg/entry"
)

func (t *Tree) Delete(domain string) bool {
	e, err := entry.NewEntry(domain, "")
	if err != nil {
		return false
	}

	tldNode, ok := t.tlds[e.TLD]
	if !ok {
		return false
	}

	if len(e.Subdomains) == 0 {
		delete(t.tlds, e.TLD)
		return true
	}

	parents := []*map[string]*node{&t.tlds}
	keys := []string{e.TLD}
	current := tldNode

	for _, subdomain := range e.Subdomains {
		child, ok := current.children[subdomain]
		if !ok {
			return false
		}
		parents = append(parents, &current.children)
		keys = append(keys, subdomain)
		current = child
	}

	last := len(parents) - 1
	delete(*parents[last], keys[last])

	for i := last - 1; i >= 1; i-- {
		parent := *parents[i]
		node := parent[keys[i]]
		if node != nil && len(node.children) == 0 && node.ip == "" {
			delete(*parents[i], keys[i])
		} else {
			break
		}
	}

	return true
}

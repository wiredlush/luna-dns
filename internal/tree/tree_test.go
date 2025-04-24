package tree

import (
	"testing"

	"github.com/wiredlush/luna-dns/internal/entry"
)

func TestBasics(t *testing.T) {
	testEntries := func(tree *Tree, entries []map[string]string, t *testing.T) {
		for _, e := range entries {
			testEntry, _ := entry.NewEntry(e["host"], e["ip"])
			tree.Insert(testEntry)

			found, _ := tree.Search(testEntry.Host)
			if found == "" {
				t.Fatal()
			}

			if found != testEntry.IP {
				t.Fatal()
			}
		}
	}

	tree := NewTree()
	testEntries(tree, []map[string]string{
		{
			"host": "test.com",
			"ip":   "127.0.0.1",
		},
		{
			"host": "a.test.com",
			"ip":   "127.0.0.1",
		},
		{
			"host": "test.a",
			"ip":   "127.0.0.1",
		},
	}, t)
}

func TestOthers(t *testing.T) {
	insertEntries := func(tree *Tree, entries []map[string]string) {
		for _, e := range entries {
			testEntry, _ := entry.NewEntry(e["host"], e["ip"])
			tree.Insert(testEntry)
		}
	}

	searchDomains := func(tree *Tree, entries []map[string]any,
		t *testing.T) {
		for _, e := range entries {
			host := e["host"].(string)
			expected := e["expected"].(bool)

			found, _ := tree.Search(host)
			if found == "" && expected {
				t.Fatal()
			}
			if found != "" && !expected {
				t.Fatal()
			}
		}
	}

	tree := NewTree()
	insertEntries(tree, []map[string]string{
		{
			"host": "*.test.com",
			"ip":   "127.0.0.1",
		},
		{
			"host": "*.tld",
			"ip":   "127.0.0.1",
		},
	})
	searchDomains(tree, []map[string]any{
		{
			"host":     "unk.com",
			"expected": false,
		},
		{
			"host":     "aaa.test.com",
			"expected": true,
		},
		{
			"host":     "test.tld",
			"expected": true,
		},
		{
			"host":     "test.xxx",
			"expected": false,
		},
	}, t)

	insertEntries(tree, []map[string]string{
		{
			"host": "*",
			"ip":   "127.0.0.1",
		},
	})
	searchDomains(tree, []map[string]any{
		{
			"host":     "google.com",
			"expected": true,
		},
	}, t)

	_, err := tree.Search("xxx")
	if err == nil {
		t.Fatal()
	}
}

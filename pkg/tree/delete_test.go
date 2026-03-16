package tree

import (
	"testing"

	"github.com/wiredlush/luna-dns/pkg/entry"
)

func TestDeleteExistingDomain(t *testing.T) {
	tr := NewTree()
	e, _ := entry.NewEntry("example.com", "1.2.3.4")
	tr.Insert(e)

	if !tr.Delete("example.com") {
		t.Fatal("Expected Delete to return true")
	}
	ip, _ := tr.Search("example.com")
	if ip != "" {
		t.Fatalf("Expected empty, got %s", ip)
	}
}

func TestDeleteNonExistentDomain(t *testing.T) {
	tr := NewTree()
	if tr.Delete("noexist.com") {
		t.Fatal("Expected Delete to return false for non-existent domain")
	}
}

func TestDeleteDoesNotAffectSiblings(t *testing.T) {
	tr := NewTree()
	e1, _ := entry.NewEntry("a.example.com", "1.1.1.1")
	e2, _ := entry.NewEntry("b.example.com", "2.2.2.2")
	tr.Insert(e1)
	tr.Insert(e2)

	tr.Delete("a.example.com")

	ip, _ := tr.Search("b.example.com")
	if ip != "2.2.2.2" {
		t.Fatalf("Expected b.example.com -> 2.2.2.2, got %s", ip)
	}
}

func TestDeleteWildcard(t *testing.T) {
	tr := NewTree()
	e, _ := entry.NewEntry("*.ads.example.com", "0.0.0.0")
	tr.Insert(e)

	if !tr.Delete("*.ads.example.com") {
		t.Fatal("Expected Delete to return true for wildcard")
	}
	ip, _ := tr.Search("banner.ads.example.com")
	if ip != "" {
		t.Fatalf("Expected empty after wildcard delete, got %s", ip)
	}
}

func TestDeletePrunesEmptyIntermediateNodes(t *testing.T) {
	tr := NewTree()
	e, _ := entry.NewEntry("deep.sub.example.com", "1.1.1.1")
	tr.Insert(e)

	tr.Delete("deep.sub.example.com")

	// The intermediate nodes (example, sub) should be pruned
	ip, _ := tr.Search("deep.sub.example.com")
	if ip != "" {
		t.Fatal("Expected empty after delete")
	}
}

func TestDeleteInvalidDomain(t *testing.T) {
	tr := NewTree()
	if tr.Delete("x") {
		t.Fatal("Expected Delete to return false for invalid domain")
	}
}

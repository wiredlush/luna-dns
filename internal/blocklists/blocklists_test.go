package blocklists

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/wiredlush/luna-dns/internal/tree"
)

func TestNewBlocklists(t *testing.T) {
	blocklists := []string{"blocklist1", "blocklist2"}
	b := NewBlocklists(blocklists, 1)
	if b.hosts == nil {
		t.Errorf("Expected hosts to be initialized")
	}
	if len(b.blocklists) != len(blocklists) {
		t.Errorf("Expected %d blocklists, got %d",
			len(blocklists), len(b.blocklists))
	}
}

func TestProcessFile(t *testing.T) {
	b := &Blocklists{}
	hosts := tree.NewTree()

	tmpfile, err := os.CreateTemp("", "blocklist-*.txt")
	if err != nil {
		t.Fatalf("Error creating temporary file: %s", err)
	}
	defer os.Remove(tmpfile.Name())
	_, err = tmpfile.Write([]byte("example.com\nexample.net\nexample.org\n"))
	if err != nil {
		t.Fatalf("Error writing to temporary file: %s", err)
	}
	err = tmpfile.Close()
	if err != nil {
		t.Fatalf("Error closing temporary file: %s", err)
	}

	b.processFile("file://"+tmpfile.Name(), hosts)
	b.hosts = hosts

	expected := []string{"example.com", "example.net", "example.org"}
	for _, domain := range expected {
		_, err := b.Search(domain)
		if err != nil {
			t.Errorf("Expected domain %s not found in tree", domain)
		}
	}
}

func TestProcessRemote(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/valid.txt" {
			fmt.Fprint(w, `
				example.com
				example.net
				example.org
			`)
		} else if r.URL.Path == "/invalid.txt" {
			fmt.Fprint(w, `
				invalid entry				
			`)
		} else {
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	b := &Blocklists{}

	validHosts := tree.NewTree()
	b.processRemote(server.URL+"/valid.txt", validHosts)
	b.hosts = validHosts

	expectedValid := []string{"example.com", "example.net", "example.org"}
	for _, domain := range expectedValid {
		_, err := b.Search(domain)
		if err != nil {
			t.Errorf("Expected domain %s not found in valid hosts tree",
				domain)
		}
	}

	invalidHosts := tree.NewTree()
	b.processRemote(server.URL+"/invalid.txt", invalidHosts)
	b.hosts = invalidHosts

	expectedInvalid := []string{"invalid entry"}
	for _, domain := range expectedInvalid {
		_, err := b.Search(domain)
		if err == nil {
			t.Errorf("Invalid domain %s found in hosts tree",
				domain)
		}
	}
}

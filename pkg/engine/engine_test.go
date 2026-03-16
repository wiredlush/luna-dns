package engine

import (
	"fmt"
	"net"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/miekg/dns"
	"github.com/wiredlush/luna-dns/pkg/config"
)

type testResponseWriter struct {
	outMessage *dns.Msg
}

func (w *testResponseWriter) LocalAddr() net.Addr {
	return &net.UDPAddr{}
}

func (w *testResponseWriter) RemoteAddr() net.Addr {
	return &net.UDPAddr{}
}

func (w *testResponseWriter) WriteMsg(m *dns.Msg) error {
	w.outMessage = m
	return nil
}

func (w *testResponseWriter) Write([]byte) (int, error) {
	return 0, nil
}

func (w *testResponseWriter) Close() error {
	return nil
}

func (w *testResponseWriter) TsigStatus() error {
	return nil
}

func (w *testResponseWriter) TsigTimersOnly(bool) {}

func (w *testResponseWriter) Hijack() {}

func TestNewEngine(t *testing.T) {
	_, err := NewEngine(&config.Config{
		Hosts: []config.Host{
			{
				Host: "google.com",
				IP:   "127.0.0.1",
			},
		},
	})
	if err != nil {
		t.Fatal()
	}

	_, err = NewEngine(&config.Config{
		Hosts: []config.Host{
			{
				Host: "x",
			},
		},
	})
	if err == nil {
		t.Fatal()
	}
}

func TestEngineStart(t *testing.T) {
	e, _ := NewEngine(&config.Config{
		Addr:    "127.0.0.1:53555",
		Network: "tcp",
		Hosts: []config.Host{
			{
				Host: "google.com",
				IP:   "127.0.0.1",
			},
		},
	})

	go func() {
		err := e.Start()
		if err != nil {
			t.Errorf("Start returned an error: %v", err)
		}
	}()

	time.Sleep(100 * time.Millisecond)

	conn, err := net.Dial(e.network, e.addr)
	if err != nil {
		t.Fatalf("Failed to dial DNS server: %v", err)
	}
	conn.Close()

	if err := e.Stop(); err != nil {
		t.Fatalf("Stop returned an error: %v", err)
	}
}

func TestEngineStopNotRunning(t *testing.T) {
	e, _ := NewEngine(&config.Config{
		Addr:    "127.0.0.1:53556",
		Network: "udp",
	})

	err := e.Stop()
	if err == nil {
		t.Fatal("Expected error when stopping a non-running engine")
	}
}

func TestEngineStartAlreadyRunning(t *testing.T) {
	e, _ := NewEngine(&config.Config{
		Addr:    "127.0.0.1:53557",
		Network: "tcp",
	})

	go e.Start()
	time.Sleep(100 * time.Millisecond)
	defer e.Stop()

	err := e.Start()
	if err == nil {
		t.Fatal("Expected error when starting an already running engine")
	}
}

func TestEngineRestart(t *testing.T) {
	e, _ := NewEngine(&config.Config{
		Addr:    "127.0.0.1:53558",
		Network: "tcp",
		Hosts: []config.Host{
			{
				Host: "google.com",
				IP:   "127.0.0.1",
			},
		},
	})

	go e.Start()
	time.Sleep(100 * time.Millisecond)

	conn, err := net.Dial(e.network, e.addr)
	if err != nil {
		t.Fatalf("Failed to dial DNS server on first start: %v", err)
	}
	conn.Close()

	if err := e.Stop(); err != nil {
		t.Fatalf("Stop returned an error: %v", err)
	}

	_, err = net.Dial(e.network, e.addr)
	if err == nil {
		t.Fatal("Expected connection to fail after stop")
	}

	go e.Start()
	time.Sleep(100 * time.Millisecond)

	conn, err = net.Dial(e.network, e.addr)
	if err != nil {
		t.Fatalf("Failed to dial DNS server after restart: %v", err)
	}
	conn.Close()

	e.Stop()
}

func TestHandler(t *testing.T) {
	engine, _ := NewEngine(&config.Config{
		Addr:    "127.0.0.1:53555",
		Network: "udp",
		Hosts: []config.Host{
			{
				Host: "google.com",
				IP:   "127.0.0.1",
			},
		},
	})

	testW := testResponseWriter{}
	engine.handler(&testW, &dns.Msg{
		MsgHdr: dns.MsgHdr{
			Opcode: dns.OpcodeQuery,
		},
		Question: []dns.Question{{
			Name:  "google.com.",
			Qtype: dns.TypeA,
		}},
	})

	response := testW.outMessage.Answer[0]
	expected, _ := dns.NewRR(fmt.Sprintf("%s A %s",
		"google.com", "127.0.0.1"))
	if response.String() != expected.String() {
		t.Fail()
	}
}

func TestFormatMessage(t *testing.T) {
	originalHeader := dns.MsgHdr{
		Id:       100,
		Response: false,
		Opcode:   500,
	}
	original := dns.Msg{
		MsgHdr: originalHeader,
	}

	out := formatMessage(&original)
	if reflect.DeepEqual(originalHeader, out.MsgHdr) {
		t.Fail()
	}
}

func TestEngineBuildForwardChain(t *testing.T) {
	dns := []config.DNS{
		{Addr: "1.1.1.1:53", Network: "udp"},
		{Addr: "2.2.2.2:53", Network: "udp"},
		{Addr: "3.3.3.3:53", Network: "udp"},
	}
	engine := &Engine{dns: dns}

	engine.forwardIndex = 1
	expectedChain := []config.DNS{
		{Addr: "2.2.2.2:53", Network: "udp"},
		{Addr: "3.3.3.3:53", Network: "udp"},
		{Addr: "1.1.1.1:53", Network: "udp"},
	}
	actualChain := engine.buildForwardChain()
	if !reflect.DeepEqual(actualChain, expectedChain) {
		t.Errorf("Test case 1 failed. Expected %v, but got %v",
			expectedChain, actualChain)
	}

	engine.forwardIndex = len(dns)
	expectedChain = dns
	actualChain = engine.buildForwardChain()
	if !reflect.DeepEqual(actualChain, expectedChain) {
		t.Errorf("Test case 2 failed. Expected %v, but got %v",
			expectedChain, actualChain)
	}

	engine.forwardIndex = len(dns) + 1
	expectedChain = dns
	actualChain = engine.buildForwardChain()
	if !reflect.DeepEqual(actualChain, expectedChain) {
		t.Errorf("Test case 3 failed. Expected %v, but got %v",
			expectedChain, actualChain)
	}
}

func TestEngineForward(t *testing.T) {
	engine, _ := NewEngine(&config.Config{
		Addr:    "127.0.0.1:53555",
		Network: "udp",
		DNS: []config.DNS{
			{
				Addr:    "8.8.8.8:53",
				Network: "udp",
			},
			{
				Addr:    "8.8.4.4:53",
				Network: "udp",
			},
		},
	})

	msg := &dns.Msg{}
	msg.SetQuestion("www.example.com.", dns.TypeA)
	engine.cache.Insert([]dns.Question{msg.Question[0]}, []dns.RR{})
	engine.forward(msg)
	if len(msg.Answer) != 0 {
		t.Fail()
	}

	msg = &dns.Msg{}
	msg.SetQuestion("www.google.com.", dns.TypeA)
	engine.forward(msg)
	if len(msg.Answer) == 0 {
		t.Fail()
	}
}

func newTestEngine() *Engine {
	e, _ := NewEngine(&config.Config{
		Addr:    "127.0.0.1:53555",
		Network: "udp",
		DNS: []config.DNS{
			{Addr: "8.8.8.8:53", Network: "udp"},
		},
	})
	return e
}

func writeTmpFile(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp("", "blocklist-*.txt")
	if err != nil {
		t.Fatalf("Error creating temporary file: %s", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("Error writing to temporary file: %s", err)
	}
	f.Close()
	t.Cleanup(func() { os.Remove(f.Name()) })
	return f.Name()
}

func TestProcessFileValidDomains(t *testing.T) {
	path := writeTmpFile(t, "example.com\nexample.net\nexample.org\n")
	engine := newTestEngine()
	engine.processFile(path, engine.blocklistTree)

	for _, domain := range []string{"example.com", "example.net", "example.org"} {
		ip, err := engine.blocklistTree.Search(domain)
		if err != nil || ip == "" {
			t.Errorf("Expected domain %s to be blocked", domain)
		}
	}
}

func TestProcessFileCommentsAndEmptyLines(t *testing.T) {
	content := "# this is a comment\n! another comment\n\n  \nexample.com\n# ignored\ntest.org\n"
	path := writeTmpFile(t, content)
	engine := newTestEngine()
	engine.processFile(path, engine.blocklistTree)

	for _, domain := range []string{"example.com", "test.org"} {
		ip, err := engine.blocklistTree.Search(domain)
		if err != nil || ip == "" {
			t.Errorf("Expected domain %s to be blocked", domain)
		}
	}

	for _, bad := range []string{"this", "another"} {
		ip, _ := engine.blocklistTree.Search(bad)
		if ip != "" {
			t.Errorf("Comment-derived entry %s should not be in tree", bad)
		}
	}
}

func TestProcessFileInvalidEntries(t *testing.T) {
	content := "example.com\ninvalid\n.bad.\nok.net\n"
	path := writeTmpFile(t, content)
	engine := newTestEngine()
	engine.processFile(path, engine.blocklistTree)

	ip, err := engine.blocklistTree.Search("example.com")
	if err != nil || ip == "" {
		t.Error("Expected example.com to be blocked")
	}
	ip, err = engine.blocklistTree.Search("ok.net")
	if err != nil || ip == "" {
		t.Error("Expected ok.net to be blocked")
	}
}

func TestProcessFileWildcards(t *testing.T) {
	content := "*.ads.example.com\ntracking.net\n"
	path := writeTmpFile(t, content)
	engine := newTestEngine()
	engine.processFile(path, engine.blocklistTree)

	ip, err := engine.blocklistTree.Search("banner.ads.example.com")
	if err != nil || ip == "" {
		t.Error("Expected banner.ads.example.com to be blocked by wildcard")
	}

	ip, err = engine.blocklistTree.Search("tracking.net")
	if err != nil || ip == "" {
		t.Error("Expected tracking.net to be blocked")
	}
}

func TestProcessFileNotFound(t *testing.T) {
	engine := newTestEngine()
	engine.processFile("/nonexistent/blocklist.txt", engine.blocklistTree)
}

func TestLoadBlocklistsEmpty(t *testing.T) {
	engine := newTestEngine()
	engine.loadBlocklists()
}

func TestLoadBlocklistsWithFiles(t *testing.T) {
	path1 := writeTmpFile(t, "one.com\ntwo.com\n")
	path2 := writeTmpFile(t, "three.org\n")

	e, _ := NewEngine(&config.Config{
		Addr:       "127.0.0.1:53555",
		Network:    "udp",
		DNS:        []config.DNS{{Addr: "8.8.8.8:53", Network: "udp"}},
		Blocklists: []string{path1, path2},
	})
	e.loadBlocklists()

	for _, domain := range []string{"one.com", "two.com", "three.org"} {
		ip, err := e.blocklistTree.Search(domain)
		if err != nil || ip == "" {
			t.Errorf("Expected domain %s to be blocked after loadBlocklists", domain)
		}
	}
}

func TestProcessFileBlocksToZeroIP(t *testing.T) {
	path := writeTmpFile(t, "blocked.com\n")
	engine := newTestEngine()
	engine.processFile(path, engine.blocklistTree)

	ip, _ := engine.blocklistTree.Search("blocked.com")
	if ip != "0.0.0.0" {
		t.Errorf("Expected blocked IP to be 0.0.0.0, got %s", ip)
	}
}

func TestNewEngineDefaultCacheTTL(t *testing.T) {
	e, err := NewEngine(&config.Config{
		Addr:    "127.0.0.1:53600",
		Network: "udp",
	})
	if err != nil {
		t.Fatal(err)
	}
	if e.cache == nil {
		t.Fatal("Expected cache to be initialized")
	}
}

func TestNewEngineCustomCacheTTL(t *testing.T) {
	e, err := NewEngine(&config.Config{
		Addr:     "127.0.0.1:53601",
		Network:  "udp",
		CacheTTL: 3600,
	})
	if err != nil {
		t.Fatal(err)
	}
	if e.cache == nil {
		t.Fatal("Expected cache to be initialized")
	}
}

func TestRunningFalseWhenNotStarted(t *testing.T) {
	e := newTestEngine()
	if e.Running() {
		t.Fatal("Expected Running() to be false before start")
	}
}

func TestRunningTrueWhenStarted(t *testing.T) {
	e, _ := NewEngine(&config.Config{
		Addr:    "127.0.0.1:53602",
		Network: "udp",
	})
	if err := e.StartBackground(); err != nil {
		t.Fatal(err)
	}
	defer e.Stop()

	if !e.Running() {
		t.Fatal("Expected Running() to be true after start")
	}
}

func TestStartBackgroundSuccess(t *testing.T) {
	e, _ := NewEngine(&config.Config{
		Addr:    "127.0.0.1:53603",
		Network: "udp",
	})
	if err := e.StartBackground(); err != nil {
		t.Fatalf("StartBackground failed: %v", err)
	}
	defer e.Stop()

	if !e.Running() {
		t.Fatal("Expected engine to be running")
	}
}

func TestStartBackgroundAlreadyRunning(t *testing.T) {
	e, _ := NewEngine(&config.Config{
		Addr:    "127.0.0.1:53604",
		Network: "udp",
	})
	if err := e.StartBackground(); err != nil {
		t.Fatal(err)
	}
	defer e.Stop()

	err := e.StartBackground()
	if err == nil {
		t.Fatal("Expected error when starting already running engine")
	}
}

func TestStartBackgroundBindError(t *testing.T) {
	e1, _ := NewEngine(&config.Config{
		Addr:    "127.0.0.1:53605",
		Network: "udp",
	})
	if err := e1.StartBackground(); err != nil {
		t.Fatal(err)
	}
	defer e1.Stop()

	e2, _ := NewEngine(&config.Config{
		Addr:    "127.0.0.1:53605",
		Network: "udp",
	})
	err := e2.StartBackground()
	if err == nil {
		t.Fatal("Expected bind error on same port")
	}
	if e2.Running() {
		t.Fatal("Engine should not be running after bind failure")
	}
}

func TestStopSuccess(t *testing.T) {
	e, _ := NewEngine(&config.Config{
		Addr:    "127.0.0.1:53606",
		Network: "udp",
	})
	if err := e.StartBackground(); err != nil {
		t.Fatal(err)
	}

	if err := e.Stop(); err != nil {
		t.Fatalf("Stop failed: %v", err)
	}
	if e.Running() {
		t.Fatal("Expected engine to not be running after stop")
	}
}

func TestSetForwardersAddAndRemove(t *testing.T) {
	e := newTestEngine()

	initial := []config.DNS{
		{Addr: "8.8.8.8:53", Network: "udp"},
		{Addr: "8.8.4.4:53", Network: "udp"},
	}
	e.SetForwarders(initial)

	if len(e.dns) != 2 {
		t.Fatalf("Expected 2 forwarders, got %d", len(e.dns))
	}

	updated := []config.DNS{
		{Addr: "1.1.1.1:53", Network: "udp"},
	}
	e.SetForwarders(updated)

	if len(e.dns) != 1 {
		t.Fatalf("Expected 1 forwarder, got %d", len(e.dns))
	}
	if e.dns[0].Addr != "1.1.1.1:53" {
		t.Fatalf("Expected 1.1.1.1:53, got %s", e.dns[0].Addr)
	}
	if e.forwardIndex != 0 {
		t.Fatal("Expected forwardIndex to be reset to 0")
	}
}

func TestSetForwardersEmpty(t *testing.T) {
	e := newTestEngine()
	e.dns = []config.DNS{{Addr: "8.8.8.8:53", Network: "udp"}}

	e.SetForwarders(nil)

	if len(e.dns) != 0 {
		t.Fatalf("Expected 0 forwarders, got %d", len(e.dns))
	}
}

func TestSetHostsValid(t *testing.T) {
	e := newTestEngine()

	e.SetHosts([]config.Host{
		{Host: "test.com", IP: "10.0.0.1"},
		{Host: "*.example.com", IP: "10.0.0.2"},
	})

	ip, err := e.hostTree.Search("test.com")
	if err != nil || ip != "10.0.0.1" {
		t.Errorf("Expected test.com -> 10.0.0.1, got %s", ip)
	}

	ip, err = e.hostTree.Search("sub.example.com")
	if err != nil || ip != "10.0.0.2" {
		t.Errorf("Expected sub.example.com -> 10.0.0.2, got %s", ip)
	}
}

func TestSetHostsWithInvalidEntry(t *testing.T) {
	e := newTestEngine()

	e.SetHosts([]config.Host{
		{Host: "valid.com", IP: "10.0.0.1"},
		{Host: "x", IP: "10.0.0.2"},
	})

	ip, err := e.hostTree.Search("valid.com")
	if err != nil || ip != "10.0.0.1" {
		t.Errorf("Expected valid.com -> 10.0.0.1, got %s", ip)
	}
}

func TestSetHostsReplacesTree(t *testing.T) {
	e := newTestEngine()

	e.SetHosts([]config.Host{{Host: "first.com", IP: "1.1.1.1"}})
	e.SetHosts([]config.Host{{Host: "second.com", IP: "2.2.2.2"}})

	ip, _ := e.hostTree.Search("first.com")
	if ip != "" {
		t.Error("Expected first.com to be gone after replacement")
	}
	ip, _ = e.hostTree.Search("second.com")
	if ip != "2.2.2.2" {
		t.Errorf("Expected second.com -> 2.2.2.2, got %s", ip)
	}
}

func TestSetBlocklistValid(t *testing.T) {
	e := newTestEngine()

	e.SetBlocklist([]string{"ads.com", "tracker.net", "*.spam.org"})

	ip, _ := e.blocklistTree.Search("ads.com")
	if ip != "0.0.0.0" {
		t.Errorf("Expected ads.com -> 0.0.0.0, got %s", ip)
	}
	ip, _ = e.blocklistTree.Search("tracker.net")
	if ip != "0.0.0.0" {
		t.Errorf("Expected tracker.net -> 0.0.0.0, got %s", ip)
	}
	ip, _ = e.blocklistTree.Search("sub.spam.org")
	if ip != "0.0.0.0" {
		t.Errorf("Expected sub.spam.org -> 0.0.0.0 via wildcard, got %s", ip)
	}
}

func TestSetBlocklistSkipsInvalid(t *testing.T) {
	e := newTestEngine()

	e.SetBlocklist([]string{"valid.com", "x", "also-valid.net"})

	ip, _ := e.blocklistTree.Search("valid.com")
	if ip != "0.0.0.0" {
		t.Errorf("Expected valid.com blocked, got %s", ip)
	}
	ip, _ = e.blocklistTree.Search("also-valid.net")
	if ip != "0.0.0.0" {
		t.Errorf("Expected also-valid.net blocked, got %s", ip)
	}
}

func TestSetBlocklistReplacesTree(t *testing.T) {
	e := newTestEngine()

	e.SetBlocklist([]string{"first.com"})
	e.SetBlocklist([]string{"second.com"})

	ip, _ := e.blocklistTree.Search("first.com")
	if ip != "" {
		t.Error("Expected first.com to be gone after replacement")
	}
	ip, _ = e.blocklistTree.Search("second.com")
	if ip != "0.0.0.0" {
		t.Errorf("Expected second.com -> 0.0.0.0, got %s", ip)
	}
}

func TestSetBlocklistEmpty(t *testing.T) {
	e := newTestEngine()
	e.SetBlocklist([]string{"ads.com"})
	e.SetBlocklist(nil)

	ip, _ := e.blocklistTree.Search("ads.com")
	if ip != "" {
		t.Error("Expected empty blocklist after setting nil")
	}
}

package engine

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/miekg/dns"
	"github.com/wiredlush/luna-dns/pkg/config"
	"github.com/wiredlush/luna-dns/pkg/tree"
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

func TestProcessFile(t *testing.T) {
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

	engine.processFile("file://"+tmpfile.Name(), engine.blocklistTree)

	expected := []string{"example.com", "example.net", "example.org"}
	for _, domain := range expected {
		_, err := engine.blocklistTree.Search(domain)
		if err != nil {
			t.Errorf("Expected domain %s not found in tree", domain)
		}
	}
}

func TestProcessRemote(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/valid.txt":
			fmt.Fprint(w, `
				example.com
				example.net
				example.org
			`)
		case "/invalid.txt":
			fmt.Fprint(w, `
				invalid entry				
			`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

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

	engine.processRemote(server.URL+"/valid.txt", engine.blocklistTree)

	expectedValid := []string{"example.com", "example.net", "example.org"}
	for _, domain := range expectedValid {
		_, err := engine.blocklistTree.Search(domain)
		if err != nil {
			t.Errorf("Expected domain %s not found in valid hosts tree",
				domain)
		}
	}

	engine.blocklistTree = tree.NewTree()
	engine.processRemote(server.URL+"/invalid.txt", engine.blocklistTree)

	expectedInvalid := []string{"invalid entry"}
	for _, domain := range expectedInvalid {
		_, err := engine.blocklistTree.Search(domain)
		if err == nil {
			t.Errorf("Invalid domain %s found in hosts tree", domain)
		}
	}
}

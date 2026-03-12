package config

import (
	"os"
	"reflect"
	"testing"
)

func TestConfig(t *testing.T) {
	writeFile := func(t *testing.T, content []byte) string {
		tmpFile, err := os.CreateTemp(".", "luna-dns_test")
		if err != nil {
			t.Fatal(err)
		}

		_, err = tmpFile.Write(content)
		if err != nil {
			t.Fatal(err)
		}

		return tmpFile.Name()
	}

	filePath := writeFile(t, []byte(`addr: 0.0.0.0:5353
network: tcp
debug: true
dns:
  - addr: "8.8.8.8:53"
    network: "udp"
hosts:
  - host: google.com
    ip: 127.0.0.1
  - host: "*.test.com"
    ip: 127.0.0.1
blocklists:
  - http://127.0.0.1/test.txt`))
	defer os.Remove(filePath)

	config, err := Load(filePath)
	if err != nil {
		t.Fatal(err)
	}

	if config.Addr != "0.0.0.0:5353" {
		t.Fatal()
	}
	if config.Network != "tcp" {
		t.Fatal()
	}
	if !reflect.DeepEqual(config.DNS, []DNS{
		{
			Addr:    "8.8.8.8:53",
			Network: "udp",
		},
	}) {
		t.Fatal()
	}
	if !reflect.DeepEqual(config.Hosts, []Host{
		{
			Host: "google.com",
			IP:   "127.0.0.1",
		},
		{
			Host: "*.test.com",
			IP:   "127.0.0.1",
		},
	}) {
		t.Fatal()
	}
	if !reflect.DeepEqual(config.Blocklists, []string{
		"http://127.0.0.1/test.txt",
	}) {
		t.Fatal()
	}

	_, err = Load("ne/not_existent_path_i_hope")
	if err == nil {
		t.Fatal()
	}
}

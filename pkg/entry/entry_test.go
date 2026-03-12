package entry

import (
	"testing"
)

func TestEntry(t *testing.T) {
	_, err := NewEntry("google.com", "127.0.0.1")
	if err != nil {
		t.Fatal(err)
	}

	_, err = NewEntry("test", "127.0.0.1")
	if err == nil {
		t.Fatal()
	}

	_, err = NewEntry(".", "127.0.0.1")
	if err == nil {
		t.Fatal()
	}
}

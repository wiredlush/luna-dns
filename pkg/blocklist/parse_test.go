package blocklist

import "testing"

func TestParseLinePlainDomain(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"example.com", "example.com"},
		{"ads.tracker.net", "ads.tracker.net"},
		{"*.ads.example.com", "*.ads.example.com"},
		{"  example.com  ", "example.com"},
	}
	for _, tt := range tests {
		if got := ParseLine(tt.input); got != tt.expected {
			t.Errorf("ParseLine(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestParseLineAdBlockPlus(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"||example.com^", "example.com"},
		{"||ads.tracker.net^", "ads.tracker.net"},
		{"||0.beer^", "0.beer"},
		{"  ||example.com^  ", "example.com"},
	}
	for _, tt := range tests {
		if got := ParseLine(tt.input); got != tt.expected {
			t.Errorf("ParseLine(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestParseLineAdBlockPlusWithModifiers(t *testing.T) {
	tests := []string{
		"||example.com^$third-party",
		"||example.com/path",
		"||example.com*",
	}
	for _, input := range tests {
		if got := ParseLine(input); got != "" {
			t.Errorf("ParseLine(%q) = %q, want empty", input, got)
		}
	}
}

func TestParseLineAdBlockPlusWildcard(t *testing.T) {
	if got := ParseLine("||*.ads.example.com^"); got != "*.ads.example.com" {
		t.Errorf("ParseLine wildcard = %q, want *.ads.example.com", got)
	}
}

func TestParseLineHostsFile(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"0.0.0.0 ads.example.com", "ads.example.com"},
		{"127.0.0.1 ads.example.com", "ads.example.com"},
		{"0.0.0.0 tracker.net", "tracker.net"},
		{"  0.0.0.0  ads.example.com  ", "ads.example.com"},
	}
	for _, tt := range tests {
		if got := ParseLine(tt.input); got != tt.expected {
			t.Errorf("ParseLine(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestParseLineHostsFileLocalhost(t *testing.T) {
	tests := []string{
		"0.0.0.0 localhost",
		"127.0.0.1 localhost",
		"0.0.0.0 localhost.localdomain",
	}
	for _, input := range tests {
		if got := ParseLine(input); got != "" {
			t.Errorf("ParseLine(%q) = %q, want empty", input, got)
		}
	}
}

func TestParseLineComments(t *testing.T) {
	tests := []string{
		"# comment",
		"! adblock comment",
		"! Title: some blocklist",
		"  # indented comment",
		"",
		"   ",
	}
	for _, input := range tests {
		if got := ParseLine(input); got != "" {
			t.Errorf("ParseLine(%q) = %q, want empty", input, got)
		}
	}
}

func TestParseLineGarbage(t *testing.T) {
	tests := []string{
		"some random text with spaces",
		"not a valid	entry with tabs",
	}
	for _, input := range tests {
		if got := ParseLine(input); got != "" {
			t.Errorf("ParseLine(%q) = %q, want empty", input, got)
		}
	}
}

package blocklist

import "strings"

// ParseLine normalizes a blocklist line from any supported format
// (plain domain, AdBlock Plus, hosts file) to a plain domain string.
// Returns empty string for comments, empty lines, or unparseable entries.
func ParseLine(raw string) string {
	line := strings.TrimSpace(raw)

	if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "!") {
		return ""
	}

	// AdBlock Plus format: ||domain^
	if domain, ok := strings.CutPrefix(line, "||"); ok {
		domain = strings.TrimSuffix(domain, "^")
		domain = strings.TrimSpace(domain)

		if strings.ContainsAny(domain, "$/") {
			return ""
		}

		if strings.Contains(domain, "*") && !strings.HasPrefix(domain, "*.") {
			return ""
		}

		return domain
	}

	// Hosts file format: 0.0.0.0 domain or 127.0.0.1 domain
	if strings.HasPrefix(line, "0.0.0.0") || strings.HasPrefix(line, "127.0.0.1") {
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			domain := fields[1]
			if domain == "localhost" || domain == "localhost.localdomain" {
				return ""
			}

			return domain
		}

		return ""
	}

	// Plain domain (no spaces)
	if strings.Contains(line, " ") || strings.Contains(line, "\t") {
		return ""
	}

	return line
}

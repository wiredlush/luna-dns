package entry

import (
	"fmt"
	"regexp"
	"strings"
)

type Entry struct {
	Host       string
	IP         string
	TLD        string
	Subdomains []string
}

func NewEntry(host, ip string) (*Entry, error) {
	domainRegex := regexp.MustCompile(`^(?:(?:(?:[\w_-]+|\*)\.)+[\w_-]+)|\*$`)
	if !domainRegex.MatchString(host) {
		return nil, fmt.Errorf("invalid host format: %s",
			host)
	}

	elements := strings.Split(host, ".")
	subdomains := []string{}
	for i := len(elements) - 2; i >= 0; i-- {
		subdomains = append(subdomains, elements[i])
	}

	return &Entry{
		Host:       host,
		IP:         ip,
		TLD:        elements[len(elements)-1],
		Subdomains: subdomains,
	}, nil
}

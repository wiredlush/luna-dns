package blocklists

func (b *Blocklists) Search(domain string) (string, error) {
	return b.hosts.Search(domain)
}

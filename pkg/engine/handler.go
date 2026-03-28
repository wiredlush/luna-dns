package engine

import (
	"net"

	"github.com/miekg/dns"
)

func (e *Engine) handler(w dns.ResponseWriter, r *dns.Msg) {
	e.stats.TotalQueries.Add(1)

	if addr := w.RemoteAddr(); addr != nil {
		host, _, err := net.SplitHostPort(addr.String())
		if err == nil {
			e.stats.TrackClient(host)
		}
	}

	for _, q := range r.Question {
		e.stats.TrackDomain(q.Name)
	}

	message := dns.Msg{}
	message.SetReply(r)

	switch r.Opcode {
	case dns.OpcodeQuery:
		e.query(&message)
	default:
		e.forward(&message)
	}

	w.WriteMsg(&message)
}

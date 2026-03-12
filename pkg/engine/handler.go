package engine

import (
	"github.com/miekg/dns"
)

func (e *Engine) handler(w dns.ResponseWriter, r *dns.Msg) {
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

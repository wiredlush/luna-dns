package engine

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"log"

	"github.com/miekg/dns"
	"github.com/wiredlush/luna-dns/internal/config"
)

func (e *Engine) forward(message *dns.Msg) {
	for _, q := range message.Question {
		ip, err := e.Blocklists.Search(q.Name[:len(q.Name)-1])
		if ip == "" || err != nil {
			continue
		}
		log.Printf("Blocked: %s: %s\n", q.Name[:len(q.Name)-1], ip)

		rr, err := dns.NewRR(fmt.Sprintf("%s A %s", q.Name, ip))
		if err == nil {
			message.Answer = append(message.Answer, rr)
			return
		}
	}

	cachedAnswer := e.cache.Search(message.Question)
	if cachedAnswer != nil {
		log.Printf("Entry found in cache: %v\n", cachedAnswer)
		message.Answer = cachedAnswer
		return
	}

	forwardChain := e.buildForwardChain()
	for _, server := range forwardChain {
		err := e.forwardRequest(server, message)
		if err == nil {
			break
		}

		log.Printf("%s (%s): %s\n", server.Addr,
			server.Network, err)
	}

	e.forwardIndex = (e.forwardIndex + 1) % len(e.dns)
}

func (e *Engine) buildForwardChain() []config.DNS {
	if e.forwardIndex >= len(e.dns) {
		e.forwardIndex = 0
		return e.dns
	}

	return append(e.dns[e.forwardIndex:],
		e.dns[:e.forwardIndex]...)
}

func (e *Engine) forwardRequest(server config.DNS, message *dns.Msg) error {
	client := &dns.Client{Net: server.Network}
	request := formatMessage(message)

	response, _, err := client.Exchange(request, server.Addr)
	if err != nil {
		return err
	}
	if response == nil || response.Rcode != dns.RcodeSuccess {
		return fmt.Errorf("failed to get a valid response")
	}

	if len(response.Answer) > 0 {
		message.Answer = response.Answer
		log.Printf("%s (%s) -> %s\n", server.Addr, server.Network,
			response.Answer)
		go e.cache.Insert(message.Question, response.Answer)
	}

	return nil
}

func formatMessage(original *dns.Msg) *dns.Msg {
	id := func() uint16 {
		var output uint16
		binary.Read(rand.Reader, binary.BigEndian, &output)
		return output
	}

	message := &dns.Msg{}
	original.CopyTo(message)

	message.MsgHdr = dns.MsgHdr{}
	message.Id = id()
	message.RecursionDesired = true

	return message
}

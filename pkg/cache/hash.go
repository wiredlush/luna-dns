package cache

import (
	"crypto/sha1"
	"encoding/hex"

	"github.com/miekg/dns"
)

func hashQuestion(question []dns.Question) string {
	hashSTR := ""
	for _, q := range question {
		hashSTR += q.String()
	}

	h := sha1.New()
	h.Write([]byte(hashSTR))

	return hex.EncodeToString(h.Sum(nil))
}

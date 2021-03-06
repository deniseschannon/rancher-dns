package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/miekg/dns"
)

func Respond(w dns.ResponseWriter, req *dns.Msg, m *dns.Msg) {
	// Figure out the max response size
	bufsize := uint16(512)
	tcp := isTcp(w)

	if o := req.IsEdns0(); o != nil {
		bufsize = o.UDPSize()
	}

	if tcp {
		bufsize = dns.MaxMsgSize - 1
	} else if bufsize < 512 {
		bufsize = 512
	}

	// Make sure the payload fits the buffer size. If the message is too large we strip the Extra section.
	// If it's still too large we return a truncated message for UDP queries and ServerFailure for TCP queries.
	if m.Len() > int(bufsize) {
		fqdn := dns.Fqdn(req.Question[0].Name)
		log.WithFields(log.Fields{"fqdn": fqdn}).Debug("Response too big, dropping Authority and Extra")
		m.Extra = nil
		if m.Len() > int(bufsize) {
			if tcp {
				log.WithFields(log.Fields{"fqdn": fqdn}).Debug("Response still too big, return ServerFailure")
				m = new(dns.Msg)
				m.SetRcode(req, dns.RcodeServerFailure)
			} else {
				log.WithFields(log.Fields{"fqdn": fqdn}).Debug("Response still too big, return truncated message")
				m.Answer = nil
				m.Truncated = true
			}
		}
	}

	err := w.WriteMsg(m)
	if err != nil {
		log.Warn("Failed to return reply: ", err, m.Len())
	}

}

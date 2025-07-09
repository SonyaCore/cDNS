package dns

import (
	"github.com/miekg/dns"
	"strings"
)

func ParseRecord(ans dns.RR, recordType string) ParsedRecord {
	parsed := ParsedRecord{
		TTL:  ans.Header().Ttl,
		Type: recordType,
	}
	switch rr := ans.(type) {
	case *dns.A:
		parsed.Address = rr.A.String()
	case *dns.AAAA:
		parsed.Address = rr.AAAA.String()
	case *dns.CNAME:
		parsed.Address = rr.Target
	case *dns.NS:
		parsed.Host = rr.Ns
	case *dns.MX:
		parsed.Host = rr.Mx
		parsed.Pref = rr.Preference
	case *dns.TXT:
		parsed.Text = strings.Join(rr.Txt, " ")
	case *dns.PTR:
		parsed.Host = rr.Ptr
	case *dns.SRV:
		parsed.Target = rr.Target
		parsed.Port = rr.Port
		parsed.Priority = rr.Priority
		parsed.Weight = rr.Weight
	case *dns.SOA:
		parsed.MName = rr.Ns
		parsed.RName = rr.Mbox
		parsed.Serial = rr.Serial
		parsed.Refresh = rr.Refresh
		parsed.Retry = rr.Retry
		parsed.Expire = rr.Expire
		parsed.Minimum = rr.Minttl
	case *dns.CAA:
		parsed.Tag = rr.Flag
		parsed.Value = rr.Value
	default:
		parsed.RawData = ans.String()
	}
	return parsed
}

package dns

import "time"

var RecordTypes = map[string]uint16{
	"A":      1,
	"AAAA":   28,
	"MX":     15,
	"NS":     2,
	"CNAME":  5,
	"TXT":    16,
	"PTR":    12,
	"SRV":    33,
	"SOA":    6,
	"CAA":    257,
	"DNSKEY": 48,
	"DS":     43,
	"RRSIG":  46,
	"NSEC":   47,
}

var PopularDNSServers = []string{
	"8.8.8.8",
	"8.8.4.4",
	"1.1.1.1",
	"1.0.0.1",
	"9.9.9.9",
	"149.112.112.112",
	"208.67.222.222",
	"208.67.220.220",
	"76.76.19.19",
	"76.223.100.101",
	"94.140.14.14",
	"94.140.15.15",
	"77.88.8.8",
	"77.88.8.1",
}

type Result struct {
	Nameserver string                    `json:"nameserver"`
	Domain     string                    `json:"domain"`
	QueryTime  time.Time                 `json:"query_time"`
	Records    map[string][]ParsedRecord `json:"records"`
	Errors     map[string]string         `json:"errors,omitempty"`
	Statistics Statistics                `json:"statistics"`
}

type ParsedRecord struct {
	TTL      uint32 `json:"ttl"`
	Type     string `json:"type"`
	Address  string `json:"address,omitempty"`
	Host     string `json:"host,omitempty"`
	Pref     uint16 `json:"pref,omitempty"`
	Text     string `json:"text,omitempty"`
	Target   string `json:"target,omitempty"`
	Port     uint16 `json:"port,omitempty"`
	Priority uint16 `json:"priority,omitempty"`
	Weight   uint16 `json:"weight,omitempty"`
	Serial   uint32 `json:"serial,omitempty"`
	Refresh  uint32 `json:"refresh,omitempty"`
	Retry    uint32 `json:"retry,omitempty"`
	Expire   uint32 `json:"expire,omitempty"`
	Minimum  uint32 `json:"minimum,omitempty"`
	MName    string `json:"mname,omitempty"`
	RName    string `json:"rname,omitempty"`
	Tag      uint8  `json:"tag,omitempty"`
	Value    string `json:"value,omitempty"`
	RawData  string `json:"raw_data,omitempty"`
}

type Statistics struct {
	TotalQueries        int           `json:"total_queries"`
	SuccessfulQueries   int           `json:"successful_queries"`
	FailedQueries       int           `json:"failed_queries"`
	AverageResponseTime time.Duration `json:"average_response_time"`
	TotalResponseTime   time.Duration `json:"total_response_time"`
}

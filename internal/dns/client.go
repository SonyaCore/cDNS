package dns

import (
	"cDNS/internal/logger"
	"go.uber.org/zap"
	"net"
	"strings"
)

func PrepareNameservers(nameservers []string) []string {
	var prepared []string
	for _, ns := range nameservers {
		if strings.HasPrefix(ns, "-") {
			continue
		}
		if !strings.Contains(ns, ":") {
			ns += ":53"
		}
		host, port, err := net.SplitHostPort(ns)
		if err != nil {
			logger.GetLogger().Warn("Invalid nameserver format", zap.String("nameserver", ns))
			continue
		}
		if net.ParseIP(host) == nil {
			if _, err := net.LookupHost(host); err != nil {
				logger.GetLogger().Warn("Cannot resolve nameserver", zap.String("host", host))
				continue
			}
		}
		prepared = append(prepared, host+":"+port)
	}
	return prepared
}

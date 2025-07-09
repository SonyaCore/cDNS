package dns

import (
	"cDNS/internal/config"
	"cDNS/internal/logger"
	"fmt"
	"github.com/miekg/dns"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"sort"
	"strings"
	"time"
)

func Query(cmd *cobra.Command, args []string) {
	cfg := config.GetConfigFromFlags(cmd)
	logger.InitLogger(cfg.LogLevel)

	if len(args) < 1 {
		logger.GetLogger().Error("Domain is required")
		err := cmd.Usage()
		if err != nil {
			return
		}
		return
	}

	domain := dns.Fqdn(args[0])
	nameservers := args[1:]

	if !IsValidDomain(domain) {
		logger.GetLogger().Fatal("Invalid domain")
	}
	nameservers = PrepareNameservers(nameservers)
	if len(nameservers) == 0 {
		logger.GetLogger().Fatal("No valid nameservers provided")
	}

	logger.GetLogger().Info("Starting DNS query", zap.String("domain", domain), zap.Strings("nameservers", nameservers))
	var allResults []Result
	for _, ns := range nameservers {
		result := Nameserver(domain, ns, cfg)
		allResults = append(allResults, result)
		if !cfg.JSONOutput {
			printHumanReadableResult(result, cfg)
		}
	}
	if cfg.JSONOutput {
		JsonOutput(allResults, cfg)
	}
	if cfg.VerboseOutput {
		summary(allResults)
	}
	logger.GetLogger().Info("DNS query completed", zap.Int("total_nameservers", len(nameservers)))
}

func Nameserver(domain, nameserver string, cfg config.Config) Result {
	originalNS := strings.Split(nameserver, ":")[0]
	result := Result{
		Nameserver: originalNS,
		Domain:     domain,
		QueryTime:  time.Now(),
		Records:    make(map[string][]ParsedRecord),
		Errors:     make(map[string]string),
		Statistics: Statistics{},
	}
	recordTypesToQuery := RecordTypes
	if len(cfg.RecordFilter) > 0 {
		recordTypesToQuery = make(map[string]uint16)
		for _, recordName := range cfg.RecordFilter {
			if recordType, exists := RecordTypes[strings.ToUpper(recordName)]; exists {
				recordTypesToQuery[strings.ToUpper(recordName)] = recordType
			}
		}
	}
	var sortedRecordNames []string
	for recordName := range recordTypesToQuery {
		sortedRecordNames = append(sortedRecordNames, recordName)
	}
	sort.Strings(sortedRecordNames)
	for _, recordName := range sortedRecordNames {
		recordType := recordTypesToQuery[recordName]
		result.Statistics.TotalQueries++
		startTime := time.Now()
		records, err := QueryDNSWithRetry(domain, nameserver, recordType, cfg)
		responseTime := time.Since(startTime)
		result.Statistics.TotalResponseTime += responseTime
		if err != nil {
			logger.GetLogger().Debug("DNS query failed", zap.String("record_type", recordName), zap.String("nameserver", nameserver), zap.Error(err))
			result.Errors[recordName] = err.Error()
			result.Statistics.FailedQueries++
			continue
		}
		result.Statistics.SuccessfulQueries++
		if len(records) == 0 {
			continue
		}
		for _, ans := range records {
			parsed := ParseRecord(ans, recordName)
			result.Records[recordName] = append(result.Records[recordName], parsed)
		}
	}
	if result.Statistics.TotalQueries > 0 {
		result.Statistics.AverageResponseTime = result.Statistics.TotalResponseTime / time.Duration(result.Statistics.TotalQueries)
	}
	return result
}

func QueryDNSWithRetry(domain, nameserver string, recordType uint16, cfg config.Config) ([]dns.RR, error) {
	var lastErr error
	for attempt := 0; attempt < cfg.Retries; attempt++ {
		records, err := QueryDNS(domain, nameserver, recordType, cfg.Timeout)
		if err == nil {
			return records, nil
		}
		lastErr = err
		if attempt < cfg.Retries-1 {
			time.Sleep(time.Duration(attempt+1) * time.Second)
		}
	}
	return nil, fmt.Errorf("failed after %d attempts: %v", cfg.Retries, lastErr)
}

func QueryDNS(domain, nameserver string, recordType uint16, timeout time.Duration) ([]dns.RR, error) {
	c := new(dns.Client)
	c.Timeout = timeout
	m := new(dns.Msg)

	// Ensure domain is fully qualified
	fqdn := dns.Fqdn(domain)
	m.SetQuestion(fqdn, recordType)
	m.RecursionDesired = true

	// Ensure nameserver has port
	if !strings.Contains(nameserver, ":") {
		nameserver = nameserver + ":53"
	}

	r, _, err := c.Exchange(m, nameserver)
	if err != nil {
		return nil, fmt.Errorf("exchange failed: %v", err)
	}
	if r.Rcode != dns.RcodeSuccess {
		return nil, fmt.Errorf("DNS error: %s", dns.RcodeToString[r.Rcode])
	}
	return r.Answer, nil
}

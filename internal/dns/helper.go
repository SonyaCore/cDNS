package dns

import (
	"cDNS/internal/config"
	"cDNS/internal/logger"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"os"
	"strings"
	"time"
)

func IsValidDomain(domain string) bool {
	if len(domain) == 0 || len(domain) > 253 {
		return false
	}
	domain = strings.TrimSuffix(domain, ".")
	parts := strings.Split(domain, ".")
	if len(parts) < 2 {
		return false
	}
	for _, part := range parts {
		if len(part) == 0 || len(part) > 63 {
			return false
		}
	}
	return true
}

func JsonOutput(results []Result, cfg config.Config) {
	jsonOutput, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		logger.GetLogger().Fatal("Failed to marshal results", zap.Error(err))
	}
	if cfg.OutputFile != "" {
		err := os.WriteFile(cfg.OutputFile, jsonOutput, 0644)
		if err != nil {
			logger.GetLogger().Fatal("Failed to write output file", zap.Error(err))
		}
		fmt.Printf("Results written to: %s\n", cfg.OutputFile)
	} else {
		fmt.Println(string(jsonOutput))
	}
}

func printHumanReadableResult(result Result, cfg config.Config) {
	fmt.Printf("\n📊 Results for %s via %s:\n", result.Domain, result.Nameserver)
	fmt.Printf("🕐 Query time: %s\n", result.QueryTime.Format(time.RFC3339))
	if len(result.Records) == 0 {
		fmt.Println("❌ No records found")
		return
	}
	for recordType, records := range result.Records {
		fmt.Printf("\n🔍 %s Records (%d found):\n", recordType, len(records))
		for i, record := range records {
			fmt.Printf("  %d. ", i+1)
			printRecord(record)
		}
	}
	if len(result.Errors) > 0 {
		fmt.Printf("\n❌ Errors:\n")
		for recordType, err := range result.Errors {
			fmt.Printf("  %s: %s\n", recordType, err)
		}
	}
	fmt.Printf("\n📈 Statistics:\n")
	fmt.Printf("  Total queries: %d\n", result.Statistics.TotalQueries)
	fmt.Printf("  Successful: %d\n", result.Statistics.SuccessfulQueries)
	fmt.Printf("  Failed: %d\n", result.Statistics.FailedQueries)
	fmt.Printf("  Average response time: %v\n", result.Statistics.AverageResponseTime)
}

func printRecord(record ParsedRecord) {
	fmt.Printf("TTL: %d", record.TTL)
	switch record.Type {
	case "A", "AAAA":
		fmt.Printf(" | Address: %s", record.Address)
	case "CNAME":
		fmt.Printf(" | Target: %s", record.Address)
	case "MX":
		fmt.Printf(" | Host: %s | Priority: %d", record.Host, record.Pref)
	case "NS":
		fmt.Printf(" | Nameserver: %s", record.Host)
	case "TXT":
		fmt.Printf(" | Text: %s", record.Text)
	case "PTR":
		fmt.Printf(" | Pointer: %s", record.Host)
	case "SRV":
		fmt.Printf(" | Target: %s | Port: %d | Priority: %d | Weight: %d", record.Target, record.Port, record.Priority, record.Weight)
	case "SOA":
		fmt.Printf(" | Master: %s | Email: %s | Serial: %d", record.MName, record.RName, record.Serial)
	case "CAA":
		fmt.Printf(" | Tag: %d | Value: %s", record.Tag, record.Value)
	default:
		if record.RawData != "" {
			fmt.Printf(" | Data: %s", record.RawData)
		}
	}
	fmt.Println()
}

func summary(results []Result) {
	fmt.Printf("\n📋 Summary:\n")
	fmt.Printf("  Nameservers queried: %d\n", len(results))
	totalQueries := 0
	totalSuccessful := 0
	totalFailed := 0
	for _, result := range results {
		totalQueries += result.Statistics.TotalQueries
		totalSuccessful += result.Statistics.SuccessfulQueries
		totalFailed += result.Statistics.FailedQueries
	}
	fmt.Printf("  Total queries: %d\n", totalQueries)
	fmt.Printf("  Successful: %d (%.1f%%)\n", totalSuccessful, float64(totalSuccessful)/float64(totalQueries)*100)
	fmt.Printf("  Failed: %d (%.1f%%)\n", totalFailed, float64(totalFailed)/float64(totalQueries)*100)
}

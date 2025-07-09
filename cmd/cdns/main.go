package main

import (
	"cDNS/internal/config"
	"cDNS/internal/logger"
	"cDNS/internal/server"
	"fmt"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"os"
	"runtime"

	"cDNS/internal/dns"
)

var (
	Version   = "dev"
	Commit    = "unknown"
	Date      = "unknown"
	GoVersion = runtime.Version()
	Platform  = runtime.GOOS + "/" + runtime.GOARCH
)

func PrintVersion() {
	commitHash := Commit
	if len(Commit) > 8 {
		commitHash = Commit[:8]
	}
	fmt.Printf("cDNS %s (%s) built with %s on %s at %s\n", Version, commitHash, GoVersion, Platform, Date)
}

func ShowDNSList(cmd *cobra.Command, args []string) {
	fmt.Println("Popular DNS Servers:")
	fmt.Println("====================")
	dnsInfo := map[string]string{
		"8.8.8.8":         "Google DNS",
		"8.8.4.4":         "Google DNS",
		"1.1.1.1":         "Cloudflare DNS",
		"1.0.0.1":         "Cloudflare DNS",
		"9.9.9.9":         "Quad9 DNS",
		"149.112.112.112": "Quad9 DNS",
		"208.67.222.222":  "OpenDNS",
		"208.67.220.220":  "OpenDNS",
		"76.76.19.19":     "Alternate DNS",
		"76.223.100.101":  "Alternate DNS",
		"94.140.14.14":    "AdGuard DNS",
		"94.140.15.15":    "AdGuard DNS",
		"77.88.8.8":       "Yandex DNS",
		"77.88.8.1":       "Yandex DNS",
	}
	for _, dnsServer := range dns.PopularDNSServers {
		fmt.Printf("%-15s - %s\n", dnsServer, dnsInfo[dnsServer])
	}
	fmt.Println("\nUsage Examples:")
	fmt.Println("cdns query google.com 8.8.8.8 1.1.1.1")
	fmt.Println("cdns query -j example.com 8.8.8.8")
	fmt.Println("cdns query --filter A,AAAA cloudflare.com 1.1.1.1")
}

func main() {
	PrintVersion()
	logger.InitLogger("info")

	rootCmd := &cobra.Command{
		Use:   "cdns",
		Short: "Enhanced DNS Query Tool",
		Long:  `A powerful DNS query tool with support for multiple nameservers, background processing, and API endpoints.`,
	}

	queryCmd := &cobra.Command{
		Use:   "query [domain] [nameservers...]",
		Short: "Query DNS records",
		Long:  `Query DNS records from specified nameservers`,
		Args:  cobra.MinimumNArgs(1), // Require at least 1 argument (domain)
		Run:   dns.Query,
		Example: `  cdns query google.com 8.8.8.8 1.1.1.1
  cdns query -j example.com 8.8.8.8
  cdns query --filter A,AAAA cloudflare.com 1.1.1.1`,
	}

	apiCmd := &cobra.Command{
		Use:   "api",
		Short: "Start API server",
		Long:  `Start the API server for background DNS checking`,
		Run:   server.RunAPI,
	}

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Run: func(cmd *cobra.Command, args []string) {
			PrintVersion()
		},
	}

	dnsListCmd := &cobra.Command{
		Use:   "dns-list",
		Short: "Show popular DNS servers",
		Run:   ShowDNSList,
	}

	// Global flags
	config.AddGlobalFlags(rootCmd)
	apiCmd.Flags().IntP("port", "p", 8080, "API server port")

	// Add flags for query command
	queryCmd.Flags().BoolP("json", "j", false, "Output in JSON format")
	queryCmd.Flags().StringP("filter", "f", "", "Filter record types (e.g., A,AAAA,MX)")
	queryCmd.Flags().IntP("timeout", "t", 5, "Query timeout in seconds")
	queryCmd.Flags().BoolP("verbose", "v", false, "Verbose output")

	rootCmd.AddCommand(queryCmd, apiCmd, versionCmd, dnsListCmd)

	if err := rootCmd.Execute(); err != nil {
		logger.GetLogger().Fatal("Failed to execute command", zap.Error(err))
		os.Exit(1)
	}
}

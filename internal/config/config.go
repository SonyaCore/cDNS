package config

import (
	"github.com/spf13/cobra"
	"time"
)

type Config struct {
	Timeout       time.Duration
	Retries       int
	JSONOutput    bool
	VerboseOutput bool
	RecordFilter  []string
	OutputFile    string
	APIPort       int
	LogLevel      string
}

func AddGlobalFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().DurationP("timeout", "t", 5*time.Second, "Query timeout")
	cmd.PersistentFlags().IntP("retries", "r", 3, "Number of retries")
	cmd.PersistentFlags().BoolP("json", "j", false, "Output in JSON format")
	cmd.PersistentFlags().BoolP("verbose", "v", false, "Verbose output")
	cmd.PersistentFlags().StringSliceP("filter", "f", []string{}, "Filter specific record types")
	cmd.PersistentFlags().StringP("output", "o", "", "Output file")
	cmd.PersistentFlags().StringP("log-level", "l", "info", "Log level (debug, info, warn, error)")
}

func GetConfigFromFlags(cmd *cobra.Command) Config {
	timeout, _ := cmd.Flags().GetDuration("timeout")
	retries, _ := cmd.Flags().GetInt("retries")
	json, _ := cmd.Flags().GetBool("json")
	verbose, _ := cmd.Flags().GetBool("verbose")
	filter, _ := cmd.Flags().GetStringSlice("filter")
	output, _ := cmd.Flags().GetString("output")
	logLevel, _ := cmd.Flags().GetString("log-level")

	return Config{
		Timeout:       timeout,
		Retries:       retries,
		JSONOutput:    json,
		VerboseOutput: verbose,
		RecordFilter:  filter,
		OutputFile:    output,
		LogLevel:      logLevel,
	}
}

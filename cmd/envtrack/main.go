package main

import (
	"fmt"
	"os"

	"github.com/envtrack/envtrack-cli/internal/commands"
	"github.com/envtrack/envtrack-cli/internal/config"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "envtrack",
		Short: "EnvTrack CLI - Manage your EnvTrack resources",
		Long:  `EnvTrack CLI is a command-line tool for interacting with the EnvTrack service.`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Set the output format from flag or config
			format, _ := cmd.Flags().GetString("format")
			if format == "" {
				format = config.GetConfig("default_format")
			}
			if format == "" {
				format = "json" // Default to JSON if not set
			}
			cmd.SetContext(config.WithOutputFormat(cmd.Context(), format))
		},
	}

	rootCmd.PersistentFlags().StringP("format", "f", "", "Output format (json, yaml, csv, bash)")

	commands.AddCommands(rootCmd)
	rootCmd.AddCommand(commands.ConfigureCommand())
	rootCmd.AddCommand(commands.VersionCommand())

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

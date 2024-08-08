package main

import (
	"fmt"
	"log"
	"os"

	"github.com/envtrack/envtrack-cli/internal/commands"
	"github.com/envtrack/envtrack-cli/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
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
				format = config.GlobalConf.GetDefaultFormat()
			}
			if format == "" {
				format = "json" // Default to JSON if not set
			}
			cmd.SetContext(config.GlobalConf.WithOutputFormat(cmd.Context(), format))
		},
	}

	rootCmd.PersistentFlags().StringP("format", "f", "", "Output format (json, yaml, csv, bash)")

	commands.AddCommands(rootCmd)

	err := doc.GenMarkdownTree(rootCmd, "./docs")
	if err != nil {
		log.Fatal(err)
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

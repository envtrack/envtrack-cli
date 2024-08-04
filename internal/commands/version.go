package commands

import (
	"fmt"
	"runtime/debug"

	"github.com/spf13/cobra"
)

var (
	// Version is the current version of the CLI
	Version = "dev"
	// CommitHash is the git commit hash of the build
	CommitHash = "unknown"

	BuildTime = "unknown"

	LocalBuildTime = "unknown"
)

func init() {
	if info, ok := debug.ReadBuildInfo(); ok {
		// log info.Settings
		for _, setting := range info.Settings {
			switch setting.Key {
			case "vcs.revision":
				CommitHash = setting.Value
			case "vcs.time":
				BuildTime = setting.Value
			}
		}
	}
}

func VersionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version number of EnvTrack CLI",
		Long:  `All software has versions. This is EnvTrack's`,
		Run:   runVersion,
	}

	return cmd
}

func runVersion(cmd *cobra.Command, args []string) {
	data := map[string]string{
		"version":          Version,
		"commit_hash":      CommitHash,
		"build_time":       BuildTime,
		"local_build_time": LocalBuildTime,
	}
	formatter, err := getFormatter(cmd.Context())
	if err != nil {
		fmt.Printf("Error getting formatter: %v\n", err)
		return
	}
	formattedOutput, _ := formatter.Format(data)
	fmt.Printf(formattedOutput)
	// fmt.Printf("EnvTrack CLI v%s\n", Version)
	// fmt.Printf("Git Commit: %s\n", CommitHash)
	// fmt.Printf("Build Time: %s\n", VersionTime)
}

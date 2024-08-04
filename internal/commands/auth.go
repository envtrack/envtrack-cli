package commands

import (
	"fmt"
	"os"

	"github.com/envtrack/envtrack-cli/internal/api"
	"github.com/envtrack/envtrack-cli/internal/config"
	"github.com/spf13/cobra"
)

func authCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "auth <token>",
		Short: "Authenticate with the EnvTrack service",
		Args:  cobra.ExactArgs(1),
		Run:   runAuth,
	}
}

func runAuth(cmd *cobra.Command, args []string) {
	token := args[0]

	// Validate the token by making a test API call
	client := api.NewClient(token)
	_, err := client.GetOrganizations()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Authentication failed: %v\n", err)
		return
	}

	// If the API call succeeds, store the token
	err = config.SetAuthToken(token)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to store auth token: %v\n", err)
		return
	}

	fmt.Println("Authentication successful. Token stored.")
}

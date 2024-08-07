package commands

import (
	"fmt"

	"github.com/envtrack/envtrack-cli/internal/config"
	"github.com/spf13/cobra"
)

func configureCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "configure",
		Aliases: []string{"conf"},
		GroupID: "conf",
		Short:   "Configure EnvTrack CLI settings",
		Long:    `Configure EnvTrack CLI settings such as API endpoint and authentication token.`,
	}

	cmd.AddCommand(
		configureSetCommand(),
		configureGetCommand(),
	)

	return cmd
}

func configureSetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a configuration value",
		Args:  cobra.ExactArgs(2),
		Run:   runConfigureSet,
	}
}

func configureGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get <key>",
		Short: "Get a configuration value",
		Args:  cobra.ExactArgs(1),
		Run:   runConfigureGet,
	}
}

func runConfigureSet(cmd *cobra.Command, args []string) {
	key := args[0]
	value := args[1]

	var err error
	if key == "auth_token" {
		err = config.GlobalConf.SetAuthToken(value)
	} else {
		err = config.GlobalConf.Set(key, value)
	}

	if err != nil {
		fmt.Printf("Error setting configuration: %v\n", err)
		return
	}

	fmt.Printf("Successfully set %s\n", key)
}

func runConfigureGet(cmd *cobra.Command, args []string) {
	key := args[0]

	var value string
	var err error

	if key == "auth_token" {
		value, err = config.GlobalConf.GetAuthToken()
	} else {
		value = config.GlobalConf.Get(key)
	}

	if err != nil {
		fmt.Printf("Error getting configuration: %v\n", err)
		return
	}

	fmt.Printf("%s: %s\n", key, value)
}

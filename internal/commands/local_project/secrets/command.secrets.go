package secrets

import "github.com/spf13/cobra"

func SecretsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "secrets",
		Aliases: []string{"sec"},
		Short:   "Manage environment secrets",
	}
	cmd.AddCommand(
		listSecretsCommand(),
		addSecretCommand(),
		getSecretsCommand(),
		removeSecretCommand(),
		updateSecretCommand(),
	)
	return cmd
}

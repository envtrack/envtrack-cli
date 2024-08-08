package variables

import (
	"github.com/spf13/cobra"
)

func LocalVariablesCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "variables",
		Aliases: []string{"vars"},
		Short:   "Manage environment variables",
	}

	cmd.AddCommand(
		listVariablesCommand(),
		addVariableCommand(),
		updateVariableCommand(),
		removeVariableCommand(),
	)

	return cmd
}

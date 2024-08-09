package varfile

import (
	"github.com/spf13/cobra"
)

func LocalVariablesCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "varfile",
		Aliases: []string{"vf"},
		Args:    cobra.NoArgs,
		Short:   "Manage variable local files",
	}

	cmd.AddCommand(
		linkVarfileCommand(),
	// addVariableCommand(),
	// updateVariableCommand(),
	// removeVariableCommand(),
	)

	return cmd
}

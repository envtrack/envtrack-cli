package varfile

import (
	"github.com/spf13/cobra"
)

func LocalVariablesCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "var-files",
		Aliases: []string{"vf"},
		Args:    cobra.NoArgs,
		Short:   "Manage variable local files",
	}

	cmd.AddCommand(
		linkVarfileCommand(),
		readVarfileCommand(),
	)

	return cmd
}

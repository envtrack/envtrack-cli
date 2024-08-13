package samplefile

import (
	"github.com/spf13/cobra"
)

func LocalSampleFileCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "sample-file",
		Aliases: []string{"sf"},
		Args:    cobra.NoArgs,
		Short:   "Manage sample files",
		RunE:    runList,
	}

	cmd.AddCommand(
		addSampleCommand(),
		remapCommand(),
		getCommand(),
	)

	return cmd
}

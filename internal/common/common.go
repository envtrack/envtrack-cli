package common

import (
	"context"
	"fmt"

	"github.com/envtrack/envtrack-cli/internal/config"
	"github.com/envtrack/envtrack-cli/internal/output"
	"github.com/spf13/cobra"
)

func GetFormatter(ctx context.Context) (output.Formatter, error) {
	format := config.GlobalConf.GetOutputFormat(ctx)
	formatter, err := output.GetFormatter(format)
	if err != nil {
		return nil, err
	}
	return formatter, nil
}

func RequireAuth(run func(cmd *cobra.Command, args []string)) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		token, err := config.GlobalConf.GetAuthToken()
		if err != nil || token == "" {
			fmt.Println("Authentication required. Please run 'envtrack auth <token>' to authenticate.")
			return
		}
		run(cmd, args)
	}
}

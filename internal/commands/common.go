package commands

import (
	"context"

	"github.com/envtrack/envtrack-cli/internal/config"
	"github.com/envtrack/envtrack-cli/internal/output"
)

func getFormatter(ctx context.Context) (output.Formatter, error) {
	format := config.GlobalConf.GetOutputFormat(ctx)
	formatter, err := output.GetFormatter(format)
	if err != nil {
		return nil, err
	}

	return formatter, nil

}

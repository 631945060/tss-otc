package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "tss-wallet",
	Short: "TSS wallet demonstration service",
	RunE: func(command *cobra.Command, _ []string) error {
		return runHTTPServer(command.Context())
	},
}

func ExecuteContext(ctx context.Context) error {
	rootCmd.SetArgs(os.Args[1:])
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		return fmt.Errorf("execute command: %w", err)
	}
	return nil
}

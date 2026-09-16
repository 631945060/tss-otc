package command

import (
	"github.com/spf13/cobra"

	"mpc-wallet-demo/apps/tss-wallet-service/internal/worker/cron"
)

var cronCmd = &cobra.Command{
	Use:   "cron",
	Short: "start scheduled operational jobs",
	RunE:  func(command *cobra.Command, _ []string) error { return cron.Run(command.Context()) },
}

func init() { rootCmd.AddCommand(cronCmd) }

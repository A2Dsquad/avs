package aggregator

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func AggregatorCommand(zLogger *zap.Logger) *cobra.Command {
	aggregatorCmd := &cobra.Command{
		Use:   "aggregator",
		Short: "aggregator command for avs",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	// // Add operator-specific subcommands here
	aggregatorCmd.AddCommand(
		Start(zLogger), // Example: 'operator start'
	)

	return aggregatorCmd
}

func Start(logger *zap.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "start",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			
			return nil
			// client.SubmitTransaction()
		},
	}
	return cmd
}

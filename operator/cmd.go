package operator

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

const (
	flagAptosNetwork      = "aptos-network"
	flagAptosConfigPath   = "aptos-config"
	flagAvsOperatorConfig = "avs-operator-config"
	flagAccountProfile    = "account-profile"
)

func OperatorCommand(zLogger *zap.Logger) *cobra.Command {
	operatorCmd := &cobra.Command{
		Use:   "operator",
		Short: "operator command for avs",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	// Add operator-specific subcommands here
	operatorCmd.AddCommand(
		Start(zLogger), // Example: 'operator start'
	)

	return operatorCmd
}

func Start(logger *zap.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "start",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get all the flags
			aptosPath, err := cmd.Flags().GetString(flagAptosConfigPath)
			if err != nil {
				return errors.Wrap(err, flagAptosConfigPath)
			}
			accountProfile, err := cmd.Flags().GetString(flagAccountProfile)
			if err != nil {
				return errors.Wrap(err, flagAccountProfile)
			}
			network, err := cmd.Flags().GetString(flagAptosNetwork)
			if err != nil {
				return errors.Wrap(err, flagAptosNetwork)
			}
			operatorConfigPath, err := cmd.Flags().GetString(flagAvsOperatorConfig)
			if err != nil {
				return errors.Wrap(err, flagAvsOperatorConfig)
			}

			networkConfig, err := extractNetwork(network)
			if err != nil {
				return fmt.Errorf("wrong config: %s", err)
			}
			operatorConfig, err := loadOperatorConfig(operatorConfigPath)
			if err != nil {
				return fmt.Errorf("can not load operator config: %s", err)
			}

			_, err = NewOperator(
				networkConfig,
				*operatorConfig,
				AptosAccountConfig{
					configPath: aptosPath,
					profile:    accountProfile,
				},
			)
			if err != nil {
				return fmt.Errorf("can not create new operator: %s", err)
			}

			return nil
			// client.SubmitTransaction()
		},
	}
	cmd.Flags().String(flagAptosConfigPath, ".aptos/config.yaml", "the path to your operator priv and pub key")
	cmd.Flags().String(flagAccountProfile, "default", "the account profile to use")
	cmd.Flags().String(flagAptosNetwork, "devnet", "choose network to connect to: mainnet, testnet, devnet, localnet")
	cmd.Flags().String(flagAvsOperatorConfig, "config/config.json", "see the example at config/example.json")
	return cmd
}

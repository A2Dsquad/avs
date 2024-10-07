package aggregator

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/aptos-labs/aptos-go-sdk"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

const (
	flagAptosNetwork     = "aptos-network"
	flagAggregatorConfig = "aggregator-config"
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
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get all the flags
			network, err := cmd.Flags().GetString(flagAptosNetwork)
			if err != nil {
				return errors.Wrap(err, flagAptosNetwork)
			}
			aggregatorConfigPath, err := cmd.Flags().GetString(flagAggregatorConfig)
			if err != nil {
				return errors.Wrap(err, flagAggregatorConfig)
			}

			aggregatorConfig, err := loadAggregatorConfig(aggregatorConfigPath)
			if err != nil {
				return fmt.Errorf("can not load aggregator config: %s", err)
			}

			networkConfig, err := extractNetwork(network)
			if err != nil {
				return fmt.Errorf("wrong config: %s", err)
			}

			aggregator, err := NewAggregator(*aggregatorConfig, logger)
			if err != nil {
				logger.Error("Cannot create aggregator", zap.Any("err", err))
				return err
			}

			// Listen for new task created in the ServiceManager contract in a separate goroutine
			go func() {
				listenErr := aggregator.SubscribeToNewTasks(networkConfig)
				if listenErr != nil {
					aggregator.logger.Fatal("Error subscribing for new tasks", zap.Any("err", listenErr))
				}
			}()

			aggregator.Start(context.Background())
			return nil
			// client.SubmitTransaction()
		},
	}
	return cmd
}

func loadAggregatorConfig(filename string) (*AggregatorConfig, error) {
	// Open the config file
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("error opening config file: %v", err)
	}
	defer file.Close()

	// Read the file contents
	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %v", err)
	}

	// Unmarshal the JSON data into the Config struct
	var config AggregatorConfig
	err = json.Unmarshal(bytes, &config)
	if err != nil {
		return nil, fmt.Errorf("error parsing config file: %v", err)
	}

	return &config, nil
}

func extractNetwork(network string) (aptos.NetworkConfig, error) {
	switch network {
	case "devnet":
		return aptos.DevnetConfig, nil
	case "localnet":
		return aptos.LocalnetConfig, nil
	case "testnet":
		return aptos.TestnetConfig, nil
	case "mainnet":
		return aptos.MainnetConfig, nil
	default:
		return aptos.NetworkConfig{}, fmt.Errorf("Choose one of: mainnet, testnet, devnet, localnet")
	}
}
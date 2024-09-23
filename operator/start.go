package operator

import (
	"fmt"

	"github.com/spf13/cobra"

	"go.uber.org/zap"

	aptos "github.com/aptos-labs/aptos-go-sdk"
)

func Start(logger *zap.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "start",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := AptosClient(aptos.DevnetConfig)

			receiver := aptos.AccountAddress{}
			err := receiver.ParseStringRelaxed("0x972290dd0c7b1b95312bd3efea78a4552d14d47dcd6f08248ba69e78366d051b")
			var start uint64 = 0
			var end uint64 = 10
			a, err := client.AccountTransactions(receiver, &start, &end)
			if err != nil {
				panic("Failed to PollForEvents:" + err.Error())
			}
			fmt.Println("Event:", a)

			event, err := client.AccountsEvents("0x972290dd0c7b1b95312bd3efea78a4552d14d47dcd6f08248ba69e78366d051b", "0")
			if err != nil {
				panic("Failed to PollForEvents:" + err.Error())
			}
			fmt.Println("Event:", event)

			price := getCMCPrice("BTC", "825")
			fmt.Println("price:", price)
			return nil
			// client.SubmitTransaction()
		},
	}
	return cmd
}

func AptosClient(networkConfig aptos.NetworkConfig) *aptos.Client {
	// Create a client for Aptos
	client, err := aptos.NewClient(networkConfig)
	if err != nil {
		panic("Failed to create client:" + err.Error())
	}
	return client
}

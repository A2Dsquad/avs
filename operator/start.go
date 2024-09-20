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

			event, err := client.AccountsEvents("0xf42eba44c19bec7a086683b446b0a3dcc3131a7e47cbf8bec1878bc6f5f3b9f2", "0")
			if err != nil {
				panic("Failed to PollForEvents:" + err.Error())
			}
			fmt.Println("Event:", event)
			return nil
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

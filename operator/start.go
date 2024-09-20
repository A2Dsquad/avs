package operator

import (
	"embed"
	"fmt"

	"github.com/spf13/cobra"

	"go.uber.org/zap"

	aptos "github.com/aptos-labs/aptos-go-sdk"
)

var migrationsFs embed.FS

func Start(logger *zap.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start [grpc-endpoint]",
		Short: "start",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			connectAptosDev(aptos.DevnetConfig)
			return nil
		},
	}
	return cmd
}

func connectAptosDev(networkConfig aptos.NetworkConfig) {
	// Create a client for Aptos
	client, err := aptos.NewClient(networkConfig)
	if err != nil {
		panic("Failed to create client:" + err.Error())
	}

	event, err := client.AccountsEvents("0xf42eba44c19bec7a086683b446b0a3dcc3131a7e47cbf8bec1878bc6f5f3b9f2", "0")
	if err != nil {
		panic("Failed to PollForEvents:" + err.Error())
	}
	fmt.Println("Event:", event)
}

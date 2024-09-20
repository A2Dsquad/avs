// transfer_coin is an example of how to make a coin transfer transaction in the simplest possible way
package main

import (
	"fmt"

	aptos "github.com/aptos-labs/aptos-go-sdk"
	"github.com/aptos-labs/aptos-go-sdk/bcs"
)

const FundAmount = 100_000_000
const TransferAmount = 1_000

// example This example shows you how to make an APT transfer transaction in the simplest possible way
func example(networkConfig aptos.NetworkConfig) {
	// Create a client for Aptos
	client, err := aptos.NewClient(networkConfig)
	if err != nil {
		panic("Failed to create client:" + err.Error())
	}

	event, err := client.AccountsEvents("0xf42eba44c19bec7a086683b446b0a3dcc3131a7e47cbf8bec1878bc6f5f3b9f2", "0")
	if err != nil {
		panic("Failed to PollForEvents:" + err.Error())
	}

	fmt.Println("event: ", event)

	// Create accounts locally for alice and bob
	alice, err := aptos.NewEd25519Account()
	if err != nil {
		panic("Failed to create alice:" + err.Error())
	}
	bob, err := aptos.NewEd25519Account()
	if err != nil {
		panic("Failed to create bob:" + err.Error())
	}

	fmt.Printf("\n=== Addresses ===\n")
	fmt.Printf("Alice: %s\n", alice.Address.String())
	fmt.Printf("Bob:%s\n", bob.Address.String())

	// Fund the sender with the faucet to create it on-chain
	err = client.Fund(alice.Address, FundAmount)
	if err != nil {
		panic("Failed to fund alice:" + err.Error())
	}

	aliceBalance, err := client.AccountAPTBalance(alice.Address)
	if err != nil {
		panic("Failed to retrieve alice balance:" + err.Error())
	}
	bobBalance, err := client.AccountAPTBalance(bob.Address)
	if err != nil {
		panic("Failed to retrieve bob balance:" + err.Error())
	}
	fmt.Printf("\n=== Initial Balances ===\n")
	fmt.Printf("Alice: %d\n", aliceBalance)
	fmt.Printf("Bob:%d\n", bobBalance)

	// Build transaction
	rawTxn, err := aptos.APTTransferTransaction(client, alice, bob.Address, TransferAmount)
	if err != nil {
		panic("Failed to build transaction:" + err.Error())
	}

	// Sign transaction
	signedTxn, err := rawTxn.SignedTransaction(alice)
	if err != nil {
		panic("Failed to sign transaction:" + err.Error())
	}

	// Submit and wait for it to complete
	submitResult, err := client.SubmitTransaction(signedTxn)
	if err != nil {
		panic("Failed to submit transaction:" + err.Error())
	}
	txnHash := submitResult.Hash

	// fmt.Println("submitResult : ", submitResult)

	// Wait for the transaction
	tx, err := client.WaitForTransaction(txnHash)
	if err != nil {
		panic("Failed to wait for transaction:" + err.Error())
	}

	fmt.Println("tx events: ", tx.Events[0].Guid)
	aliceBalance, err = client.AccountAPTBalance(alice.Address)
	if err != nil {
		panic("Failed to retrieve alice balance:" + err.Error())
	}
	bobBalance, err = client.AccountAPTBalance(bob.Address)
	if err != nil {
		panic("Failed to retrieve bob balance:" + err.Error())
	}
	fmt.Printf("\n=== Intermediate Balances ===\n")
	fmt.Printf("Alice: %d\n", aliceBalance)
	fmt.Printf("Bob:%d\n", bobBalance)

	// Now submit as a single call, with a custom payload
	accountBytes, err := bcs.Serialize(&bob.Address)
	if err != nil {
		panic("Failed to serialize bob's address:" + err.Error())
	}

	amountBytes, err := bcs.SerializeU64(TransferAmount)
	if err != nil {
		panic("Failed to serialize transfer amount:" + err.Error())
	}

	resp, err := client.BuildSignAndSubmitTransaction(alice, aptos.TransactionPayload{
		Payload: &aptos.EntryFunction{
			Module: aptos.ModuleId{
				Address: aptos.AccountOne,
				Name:    "aptos_account",
			},
			Function: "transfer",
			ArgTypes: []aptos.TypeTag{},
			Args: [][]byte{
				accountBytes,
				amountBytes,
			},
		}},
	)
	if err != nil {
		panic("Failed to sign transaction:" + err.Error())
	}

	_, err = client.WaitForTransaction(resp.Hash)
	if err != nil {
		panic("Failed to wait for transaction:" + err.Error())
	}

	aliceBalance, err = client.AccountAPTBalance(alice.Address)
	if err != nil {
		panic("Failed to retrieve alice balance:" + err.Error())
	}
	bobBalance, err = client.AccountAPTBalance(bob.Address)
	if err != nil {
		panic("Failed to retrieve bob balance:" + err.Error())
	}
	fmt.Printf("\n=== Final Balances ===\n")
	fmt.Printf("Alice: %d\n", aliceBalance)
	fmt.Printf("Bob:%d\n", bobBalance)
}

func main() {
	example(aptos.DevnetConfig)
}

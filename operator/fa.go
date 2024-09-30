package operator

import (
	"encoding/json"
	"fmt"

	aptos "github.com/aptos-labs/aptos-go-sdk"
)

const FAContract = "0x8a3ef52a6db858766859f9947abe749b0f2140878b34dd5c2c122c94ababe2bb"

func GetMetadata(
	client *aptos.Client,
) Metadata {
	contractAcc := aptos.AccountAddress{}
	err := contractAcc.ParseStringRelaxed(FAContract)
	if err != nil {
		panic("Could not ParseStringRelaxed:" + err.Error())
	}

	var noTypeTags []aptos.TypeTag
	viewResponse, err := client.View(&aptos.ViewPayload{
		Module: aptos.ModuleId{
			Address: contractAcc,
			Name:    "fungible_asset",
		},
		Function: "get_metadata",
		ArgTypes: noTypeTags,
		Args:     [][]byte{},
	})
	if err != nil {
		panic("Failed to view fa address:" + err.Error())
	}
	metadataMap := viewResponse[0].(map[string]interface{})
	metadataBz, err := json.Marshal(metadataMap)
	if err != nil {
		panic("Failed to marshal metadata to json:" + err.Error())
	}

	var metadataStr MetadataStr
	err = json.Unmarshal(metadataBz, &metadataStr)
	if err != nil {
		panic("Failed to unmarshal metadata from json:" + err.Error())
	}
	metadataAcc := aptos.AccountAddress{}
	err = metadataAcc.ParseStringRelaxed(metadataStr.Inner)
	if err != nil {
		panic("Could not ParseStringRelaxed:" + err.Error())
	}

	fmt.Println("metadata: ", metadataAcc.String())
	return Metadata{
		Inner: metadataAcc,
	}
}

func FAMetdataClient(
	client *aptos.Client,
) *aptos.FungibleAssetClient {
	contractAcc := aptos.AccountAddress{}
	err := contractAcc.ParseStringRelaxed(FAContract)
	if err != nil {
		panic("Could not ParseStringRelaxed:" + err.Error())
	}

	var noTypeTags []aptos.TypeTag
	viewResponse, err := client.View(&aptos.ViewPayload{
		Module: aptos.ModuleId{
			Address: contractAcc,
			Name:    "fungible_asset",
		},
		Function: "get_metadata",
		ArgTypes: noTypeTags,
		Args:     [][]byte{},
	})
	if err != nil {
		panic("Failed to view fa address:" + err.Error())
	}

	metadataMap := viewResponse[0].(map[string]interface{})
	metadataBz, err := json.Marshal(metadataMap)
	if err != nil {
		panic("Failed to marshal metadata to json:" + err.Error())
	}

	var metadataStr MetadataStr
	err = json.Unmarshal(metadataBz, &metadataStr)
	if err != nil {
		panic("Failed to unmarshal metadata from json:" + err.Error())
	}
	metadataAcc := aptos.AccountAddress{}
	err = metadataAcc.ParseStringRelaxed(metadataStr.Inner)
	if err != nil {
		panic("Could not ParseStringRelaxed:" + err.Error())
	}

	faClient, err := aptos.NewFungibleAssetClient(client, &metadataAcc)
	if err != nil {
		panic("Failed to create fa client:" + err.Error())
	}
	return faClient
}

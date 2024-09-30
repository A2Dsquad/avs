#! /bin/bash

# current seed 1009
seed=$1

aptos move compile 

address=$(yq -r '.profiles.default.account' ./.aptos/config.yaml)

resource_account="0x$(aptos account derive-resource-account-address --address $address --seed $seed | jq -r '.Result')"

echo $resource_account

sed -i -E "s|^\(oracle = \).*|\1\"$resource_account\"|" ./Move.toml

aptos move create-resource-account-and-publish-package --seed $seed --address-name default --assume-yes 
aptos move run-script --compiled-script-path ./build/oracle/bytecode_scripts/initialize_avs_modules.mv  --assume-yes 

echo "Deployed to $resource_account"

go run ./cmd/main.go operator config $resource_account 123   
go run ./cmd/main.go operator initialize-quorum 1 "1"
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

go run ./cmd/main.go operator config $resource_account localhost:26657   
go run ./cmd/main.go operator initialize-quorum 1 "1"

go run ./cmd/main.go aggregator config $resource_account localhost:26657

aptos move run --function-id $resource_account::service_manager::create_new_task --args address:0x7a31ea2f1026154178a41c45e298c82f4033ae6608bffa0642c54d64e5c4faaa u64:0 string:lmao2 u64:0 --assume-yes
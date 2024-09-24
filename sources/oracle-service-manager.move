module oracle::service_manager{
    use aptos_framework::event;
    use aptos_framework::fungible_asset::{
        Self, FungibleAsset, FungibleStore, Metadata,
    };
    use aptos_framework::object::{Self, Object};
    use aptos_framework::primary_fungible_store;
    use aptos_framework::timestamp;
    use aptos_framework::account::{Self, SignerCapability};
    use aptos_std::aptos_hash;
    use aptos_std::smart_table::{Self, SmartTable};
    
    use std::string::{Self, String};
    use std::bcs;
    use std::vector;
    use std::signer;

    use oracle::oracle_manager; 
    use oracle::registry_coordinator;
    use oracle::service_manager_base;
    use oracle::bls_apk_registry;
    use oracle::stake_registry;
    use oracle::fee_pool;

    const SERVICE_MANAGER_NAME: vector<u8> = b"SERVICE_MANAGER_NAME";
    const SERVICE_PREFIX: vector<u8> = b"SERVICE_PREFIX";

    const EBATCH_ALREADY_SUBMITTED: u64 = 1300;
    const EINSUFFICIENT_FUNDS: u64 = 1301;

    struct BatchState has store, drop, copy {
        task_created_timestamp: u64,
        responded: bool,
        respond_fee_token: Object<Metadata>,
        respond_fee_limit: u64
    }

    struct ServiceManagerStore has key {
        batches_state: SmartTable<vector<u8>, BatchState>,
    }

    struct FeePool has key {
        token_store: Object<FungibleStore>
    }
    struct BatcherBalanceStore has key, store {
        balances: SmartTable<Object<Metadata>, u64>,
    }

    struct ServiceManagerConfigs has key {
        signer_cap: SignerCapability,
    }

    #[event]
    struct BatcherBalanceUpdated has drop, store {
        batcher: address,
        token: Object<Metadata>,
        amount: u64,
    }

    #[event]
    struct BatchCreated has drop, store {
        batch_merkle_root: vector<u8>,
        batcher: address,
        timestamp: u64,
        batch_data_pointer: String,
        respond_fee_token: Object<Metadata>,
        respond_fee_limit: u64
    }
    public entry fun initialize() {
        if (is_initialized()) {
            return
        };

        // derive a resource account from signer to manage User share Account
        let oracle_signer = &oracle_manager::get_signer();
        let (service_manager_signer, signer_cap) = account::create_resource_account(oracle_signer, SERVICE_MANAGER_NAME);
        oracle_manager::add_address(string::utf8(SERVICE_MANAGER_NAME), signer::address_of(&service_manager_signer));
        move_to(&service_manager_signer, ServiceManagerConfigs {
            signer_cap,
        });
    }

    public entry fun create_new_task(
        batcher: &signer,
        token: Object<Metadata>,
        amount: u64,
        batch_merkle_root: vector<u8>,
        batch_data_pointer: String,
        respond_fee_limit: u64
    ) acquires ServiceManagerConfigs, ServiceManagerStore, BatcherBalanceStore {
        let batcher_address = signer::address_of(batcher);
       
        let hash_data = vector<u8>[];
        vector::append(&mut hash_data, batch_merkle_root);
        vector::append(&mut hash_data, batcher_store_seeds(batcher_address));

        let batch_identifier = aptos_hash::keccak256(hash_data);
        assert!(smart_table::contains(&service_manager_store().batches_state, batch_identifier), EBATCH_ALREADY_SUBMITTED);

        if (amount > 0) {
            let store = primary_fungible_store::ensure_primary_store_exists(batcher_address, token);
            let fa = fungible_asset::withdraw(batcher, store, amount);
        
            let pool = fee_pool::ensure_fee_pool(token);
            fee_pool::deposit(pool, fa);

            // Ensure batcher balance store is created
            ensure_batcher_balance_store(batcher_address);
            let store_mut = batcher_balance_store_mut(batcher_address);
            let current_balance = smart_table::borrow_mut_with_default(&mut store_mut.balances, token, 0);

            *current_balance = *current_balance + amount;

            event::emit(BatcherBalanceUpdated {
                batcher: batcher_address, 
                token,
                amount: *current_balance
            });
        };

        let store = batcher_balance_store(batcher_address);
        let current_balance = smart_table::borrow_with_default(&store.balances, token, &0);

        assert!(*current_balance >= respond_fee_limit, EINSUFFICIENT_FUNDS);

        let now = timestamp::now_seconds();
        let store_mut = service_manager_store_mut();
        smart_table::upsert(&mut store_mut.batches_state, batch_identifier, BatchState{
            task_created_timestamp: now,
            responded: false,
            respond_fee_token: token, 
            respond_fee_limit,
        });
        
        event::emit(BatchCreated{
            batch_merkle_root,
            batcher: batcher_address, 
            timestamp: now,
            batch_data_pointer,
            respond_fee_token: token,
            respond_fee_limit
        })
    }

    // public entry fun respond_to_task(
    //     aggregator: &signer,
    //     batch_merkle_root: vector<u8>,
    //     non_signer_stakes_and_signatur: NonSignerStakesAndSignature
    // ) acquires ServiceManagerStore, BatcherBalanceStore {
        
    // }

    #[view]
    public fun is_initialized(): bool{
        oracle_manager::address_exists(string::utf8(SERVICE_MANAGER_NAME))
    }

    #[view]
    /// Return the address of the resource account that stores pool manager configs.
    public fun service_manager_address(): address {
        oracle_manager::get_address(string::utf8(SERVICE_MANAGER_NAME))
    }

    public fun create_service_manager_store() acquires ServiceManagerConfigs{
        let service_manager_signer = service_manager_signer();
        move_to(service_manager_signer, ServiceManagerStore{
            batches_state: smart_table::new()
        })
    }

    fun ensure_service_manager_store() acquires ServiceManagerConfigs{
        if(!exists<ServiceManagerStore>(service_manager_address())){
            create_service_manager_store();
        }
    }

    fun ensure_batcher_balance_store(batcher_address: address) acquires ServiceManagerConfigs{
        if(!exists<BatcherBalanceStore>(batcher_address)){
            create_batcher_balance_store(batcher_address);
        }
    }

    public fun create_batcher_balance_store(batcher_address: address) acquires ServiceManagerConfigs{
        let service_manager_signer = service_manager_signer();
        let ctor = &object::create_named_object(service_manager_signer, batcher_store_seeds(batcher_address));
        let batcher_balance_store_signer = object::generate_signer(ctor);
    
        move_to(&batcher_balance_store_signer, ServiceManagerStore{
            batches_state: smart_table::new()
        })
    }

    inline fun service_manager_store(): &ServiceManagerStore  acquires ServiceManagerStore {
        borrow_global<ServiceManagerStore>(service_manager_address())
    }

    inline fun service_manager_store_mut(): &mut ServiceManagerStore  acquires ServiceManagerStore {
        borrow_global_mut<ServiceManagerStore>(service_manager_address())
    }

    inline fun batcher_balance_store(batcher_address: address): &BatcherBalanceStore acquires BatcherBalanceStore {
        borrow_global<BatcherBalanceStore>(batcher_balance_store_address(batcher_address))
    }

    inline fun batcher_balance_store_mut(batcher_address: address): &mut BatcherBalanceStore acquires BatcherBalanceStore {
        borrow_global_mut<BatcherBalanceStore>(batcher_balance_store_address(batcher_address))
    }

    inline fun batcher_balance_store_address(batcher_address: address): address {
        object::create_object_address(&service_manager_address(), batcher_store_seeds(batcher_address))
    }

    inline fun service_manager_signer(): &signer acquires ServiceManagerConfigs{
        &account::create_signer_with_capability(&borrow_global<ServiceManagerConfigs>(service_manager_address()).signer_cap)
    }

    inline fun batcher_store_seeds(batcher_address: address): vector<u8>{
        let seeds = vector<u8>[];
        vector::append(&mut seeds, SERVICE_PREFIX);
        vector::append(&mut seeds, bcs::to_bytes(&batcher_address));
        seeds
    }
}
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
    use oracle::bls_sig_checker;

    const SERVICE_MANAGER_NAME: vector<u8> = b"SERVICE_MANAGER_NAME";
    const SERVICE_PREFIX: vector<u8> = b"SERVICE_PREFIX";

    const THRESHOLD_DENOMINATOR: u128 = 100; 
    const QUORUM_THRESHOLD_PERCENTAGE: u128 = 67;
    const PRECISION: u64 = 1000000000;

    const ETASK_ALREADY_SUBMITTED: u64 = 1300;
    const ETASK_DOES_NOT_EXIST: u64 = 1301;
    const ETASK_ALREADY_RESPONDED: u64 = 1302;
    const ETASK_HAS_NO_BALANCE: u64 = 1303;
    const EINSUFFICIENT_FUNDS: u64 = 1304;
    const ETHRESHOLD_NOT_MEET: u64 = 1305;
    const EINVALID_TASK_ID: u64 = 1306;
    const ESIGNER_SIGNATURE_AMOUNT_NOT_MATCH: u64 = 1307;
    const ESIGNER_RESPONSE_AMOUNT_NOT_MATCH: u64 = 1308;

    struct TaskState has store, drop, copy {
        task_created_timestamp: u64,
        responded: bool,
        respond_fee_token: Object<Metadata>,
        respond_fee_limit: u64
    }

    struct ServiceManagerStore has key {
        tasks_state: SmartTable<vector<u8>, TaskState>,
        task_count: u64,
    }

    struct FeePool has key {
        token_store: Object<FungibleStore>
    }
    struct TaskCreatorBalanceStore has key, store {
        balances: SmartTable<Object<Metadata>, u64>,
    }

    struct ServiceManagerConfigs has key {
        signer_cap: SignerCapability,
    }

    #[event]
    struct TaskCreatorBalanceUpdated has drop, store {
        creator: address,
        token: Object<Metadata>,
        amount: u64,
    }

    #[event]
    struct TaskCreated has drop, store {
        creator: address,
        timestamp: u64,
        data_request: String,
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
        creator: &signer,
        token: Object<Metadata>,
        amount: u64,
        data_request: String,
        respond_fee_limit: u64
    ) acquires ServiceManagerConfigs, ServiceManagerStore, TaskCreatorBalanceStore {
        ensure_service_manager_store();
        let creator_address = signer::address_of(creator);
        let store = service_manager_store();
        let task_id = store.task_count + 1;
        let hash_data = vector<u8>[];
        vector::append(&mut hash_data, u64_to_vector_u8(task_id));
        vector::append(&mut hash_data, task_creator_store_seeds(creator_address));

        let task_identifier = aptos_hash::keccak256(hash_data);
        assert!(!smart_table::contains(&store.tasks_state, task_identifier), ETASK_ALREADY_SUBMITTED);

        if (amount > 0) {
            let store = primary_fungible_store::ensure_primary_store_exists(creator_address, token);
            let fa = fungible_asset::withdraw(creator, store, amount);
        
            let pool = fee_pool::ensure_fee_pool(token);
            fee_pool::deposit(pool, fa);

            // Ensure creator balance store is created
            ensure_task_creator_balance_store(creator_address);
            let store_mut = task_creator_balance_store_mut(creator_address);
            let current_balance = smart_table::borrow_mut_with_default(&mut store_mut.balances, token, 0);

            *current_balance = *current_balance + amount;

            event::emit(TaskCreatorBalanceUpdated {
                creator: creator_address, 
                token,
                amount: *current_balance
            });
        };

        let current_balance = smart_table::borrow_with_default(&task_creator_balance_store(creator_address).balances, token, &0);

        assert!(*current_balance >= respond_fee_limit, EINSUFFICIENT_FUNDS);

        let now = timestamp::now_seconds();
        let store_mut = service_manager_store_mut();
        store_mut.task_count = task_id;
        smart_table::upsert(&mut store_mut.tasks_state, task_identifier, TaskState{
            task_created_timestamp: now,
            responded: false,
            respond_fee_token: token, 
            respond_fee_limit,
        });
        
        event::emit(TaskCreated{
            creator: creator_address, 
            timestamp: now,
            data_request,
            respond_fee_token: token,
            respond_fee_limit
        })
    }

    public entry fun respond_to_task(
        aggregator: &signer,
        task_id: u64,
        sender: address,
        responses: vector<u128>,
        signer_pubkeys: vector<vector<u8>>,
        signer_sigs: vector<vector<u8>>,
        quorum_aggr_pks: vector<vector<u8>>,
        quorum_apk_indices: vector<u64>,
        total_stake_indices: vector<u64>,
        signer_stake_indices: vector<vector<u64>>,
    ) acquires ServiceManagerStore {
        let task_count = service_manager_store().task_count;
        assert!(task_id <= task_count, EINVALID_TASK_ID);
        let hash_data = vector<u8>[];
        vector::append(&mut hash_data, u64_to_vector_u8(task_id));
        vector::append(&mut hash_data, task_creator_store_seeds(sender));

        let signer_amount = vector::length(&signer_pubkeys);
        assert!(signer_amount == vector::length(&signer_sigs), ESIGNER_SIGNATURE_AMOUNT_NOT_MATCH);
        assert!(signer_amount == vector::length(&responses), ESIGNER_RESPONSE_AMOUNT_NOT_MATCH);
        let task_identifier = aptos_hash::keccak256(hash_data);
        
        let msg_hashes: vector<vector<u8>> = vector::empty();
        for (i in 0..signer_amount) {
            let response = *vector::borrow(&responses, i);
            let msg_hash = hash_data;
            vector::append(&mut msg_hash, u128_to_vector_u8(response));
            let msg_hash = aptos_hash::keccak256(msg_hash);
            vector::push_back(&mut msg_hashes, msg_hash);
        };

        assert!(smart_table::contains(&service_manager_store().tasks_state, task_identifier), ETASK_DOES_NOT_EXIST);
        assert!(smart_table::borrow(&service_manager_store().tasks_state, task_identifier).responded, ETASK_ALREADY_RESPONDED);
        assert!(smart_table::borrow(&service_manager_store().tasks_state, task_identifier).respond_fee_limit > 0, ETASK_HAS_NO_BALANCE);

        let store_mut = service_manager_store_mut();
        let task_state = smart_table::borrow_mut(&mut store_mut.tasks_state, task_identifier);
        task_state.responded = true;
        
        let (signed_stake_for_quorum, total_stake_for_quorum) = bls_sig_checker::check_signatures(
            vector::singleton(0),
            smart_table::borrow(&service_manager_store().tasks_state, task_identifier).task_created_timestamp,
            msg_hashes,
            signer_pubkeys,
            signer_sigs,
            quorum_aggr_pks,
            quorum_apk_indices,
            total_stake_indices,
            signer_stake_indices
        );

        let signed_stake = *vector::borrow(&signed_stake_for_quorum, 0);
        let total_stake = *vector::borrow(&total_stake_for_quorum, 0);
        assert!((signed_stake * THRESHOLD_DENOMINATOR) >= (total_stake * QUORUM_THRESHOLD_PERCENTAGE), ETHRESHOLD_NOT_MEET);
    }

    #[view]
    public fun is_initialized(): bool{
        oracle_manager::address_exists(string::utf8(SERVICE_MANAGER_NAME))
    }

    #[view]
    /// Return the address of the resource account that stores pool manager configs.
    public fun service_manager_address(): address {
        oracle_manager::get_address(string::utf8(SERVICE_MANAGER_NAME))
    }

    #[view]
    public fun task_count(): u64  acquires ServiceManagerStore{
        service_manager_store().task_count
    }

    public fun create_service_manager_store() acquires ServiceManagerConfigs{
        let service_manager_signer = service_manager_signer();
        move_to(service_manager_signer, ServiceManagerStore{
            tasks_state: smart_table::new(),
            task_count: 0
        })
    }

    fun ensure_service_manager_store() acquires ServiceManagerConfigs{
        if(!exists<ServiceManagerStore>(service_manager_address())){
            create_service_manager_store();
        }
    }

    fun ensure_task_creator_balance_store(creator_address: address) acquires ServiceManagerConfigs{
        if(!exists<TaskCreatorBalanceStore>(creator_address)){
            create_task_creator_balance_store(creator_address);
        }
    }

    public fun create_task_creator_balance_store(creator_address: address) acquires ServiceManagerConfigs{
        let service_manager_signer = service_manager_signer();
        let ctor = &object::create_named_object(service_manager_signer, task_creator_store_seeds(creator_address));
        let task_creator_balance_store_signer = object::generate_signer(ctor);
    
        move_to(&task_creator_balance_store_signer, TaskCreatorBalanceStore{
            balances: smart_table::new(),
        })
    }

    inline fun u64_to_vector_u8(value: u64): vector<u8> {
        let result = vector::empty<u8>();
        
        // Extract each byte from the u64 value
        let shift = 56; // Start with the highest byte
        while (shift >= 0) {
            let byte = (((value >> shift) & 0xFF) as u8);
            vector::push_back(&mut result, byte);
            if (shift == 0) {
                break
            };
            shift = shift - 8;
        };
        result
    }

    inline fun u128_to_vector_u8(value: u128): vector<u8> {
        let result = vector::empty<u8>();
        
        // Extract each byte from the u64 value
        let shift = 120; // Start with the highest byte
        while (shift >= 0) {
            let byte = (((value >> shift) & 0xFF) as u8);
            vector::push_back(&mut result, byte);
            if (shift == 0) {
                break
            };
            shift = shift - 8;
        };
        result
    }

    inline fun service_manager_store(): &ServiceManagerStore  acquires ServiceManagerStore {
        borrow_global<ServiceManagerStore>(service_manager_address())
    }

    inline fun service_manager_store_mut(): &mut ServiceManagerStore  acquires ServiceManagerStore {
        borrow_global_mut<ServiceManagerStore>(service_manager_address())
    }

    inline fun task_creator_balance_store(creator_address: address): &TaskCreatorBalanceStore acquires TaskCreatorBalanceStore {
        borrow_global<TaskCreatorBalanceStore>(task_creator_balance_store_address(creator_address))
    }

    inline fun task_creator_balance_store_mut(creator_address: address): &mut TaskCreatorBalanceStore acquires TaskCreatorBalanceStore {
        borrow_global_mut<TaskCreatorBalanceStore>(task_creator_balance_store_address(creator_address))
    }

    inline fun task_creator_balance_store_address(creator_address: address): address {
        object::create_object_address(&service_manager_address(), task_creator_store_seeds(creator_address))
    }

    inline fun service_manager_signer(): &signer acquires ServiceManagerConfigs{
        &account::create_signer_with_capability(&borrow_global<ServiceManagerConfigs>(service_manager_address()).signer_cap)
    }

    inline fun task_creator_store_seeds(creator_address: address): vector<u8>{
        let seeds = vector<u8>[];
        vector::append(&mut seeds, SERVICE_PREFIX);
        vector::append(&mut seeds, bcs::to_bytes(&creator_address));
        seeds
    }
}
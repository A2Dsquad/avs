script {
  fun initialize_avs_modules() {
    oracle::service_manager::initialize();
    oracle::service_manager_base::initialize();
    oracle::bls_apk_registry::initialize();
    oracle::bls_sig_checker::initialize();
    oracle::index_registry::initialize();
    oracle::stake_registry::initialize();
    oracle::registry_coordinator::initialize();
    
  }
}
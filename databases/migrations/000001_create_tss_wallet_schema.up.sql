CREATE TABLE IF NOT EXISTS schema_migrations (
    version VARCHAR(64) NOT NULL PRIMARY KEY,
    applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS wallets (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    wallet_code VARCHAR(96) NOT NULL UNIQUE,
    public_key TEXT NOT NULL,
    address VARCHAR(160) NOT NULL,
    network VARCHAR(32) NOT NULL,
    status VARCHAR(24) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    KEY idx_wallets_network_status (network, status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS wallet_share_refs (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    wallet_id BIGINT UNSIGNED NOT NULL,
    node_code VARCHAR(96) NOT NULL,
    share_ref VARCHAR(255) NOT NULL,
    fingerprint VARCHAR(128) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uq_wallet_node_share (wallet_id, node_code),
    KEY idx_wallet_share_refs_wallet (wallet_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS participants (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    node_code VARCHAR(96) NOT NULL UNIQUE,
    endpoint VARCHAR(255) NOT NULL,
    certificate_fingerprint VARCHAR(128) NOT NULL,
    status VARCHAR(24) NOT NULL,
    key_epoch INT NOT NULL DEFAULT 1,
    version VARCHAR(64) NOT NULL DEFAULT '',
    last_heartbeat TIMESTAMP NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    KEY idx_participants_status_heartbeat (status, last_heartbeat)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS key_epochs (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    wallet_id BIGINT UNSIGNED NOT NULL,
    epoch_no INT NOT NULL,
    operation VARCHAR(32) NOT NULL,
    public_key_unchanged TINYINT(1) NOT NULL DEFAULT 0,
    status VARCHAR(24) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uq_key_epochs_wallet_epoch (wallet_id, epoch_no),
    KEY idx_key_epochs_wallet_status (wallet_id, status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS transactions (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    transaction_code VARCHAR(96) NOT NULL UNIQUE,
    wallet_id BIGINT UNSIGNED NOT NULL,
    from_address VARCHAR(160) NOT NULL,
    to_address VARCHAR(160) NOT NULL,
    amount DECIMAL(36,18) NOT NULL,
    fee DECIMAL(36,18) NOT NULL DEFAULT 0,
    digest VARCHAR(160) NOT NULL,
    tx_hash VARCHAR(128) NULL,
    status VARCHAR(32) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    KEY idx_transactions_wallet_status_created (wallet_id, status, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS sign_sessions (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    session_code VARCHAR(96) NOT NULL UNIQUE,
    wallet_id BIGINT UNSIGNED NOT NULL,
    transaction_id BIGINT UNSIGNED NULL,
    digest VARCHAR(160) NOT NULL,
    threshold_value INT NOT NULL,
    participant_set JSON NOT NULL,
    status VARCHAR(32) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    finished_at TIMESTAMP NULL,
    KEY idx_sign_sessions_wallet_status_created (wallet_id, status, created_at),
    KEY idx_sign_sessions_transaction (transaction_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS approvals (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    session_id BIGINT UNSIGNED NOT NULL,
    node_id BIGINT UNSIGNED NOT NULL,
    decision VARCHAR(16) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE KEY uq_approval_session_node (session_id, node_id),
    KEY idx_approvals_session_created (session_id, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS audit_logs (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    request_id VARCHAR(96) NOT NULL,
    actor VARCHAR(96) NOT NULL,
    action VARCHAR(96) NOT NULL,
    result VARCHAR(24) NOT NULL,
    session_code VARCHAR(96) NULL,
    detail TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    KEY idx_audit_logs_action_created (action, created_at),
    KEY idx_audit_logs_session_created (session_code, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS node_heartbeats (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    node_id BIGINT UNSIGNED NOT NULL,
    version VARCHAR(64) NOT NULL,
    heartbeat_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    KEY idx_node_heartbeats_node_time (node_id, heartbeat_time)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

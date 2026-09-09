CREATE TABLE IF NOT EXISTS mm_branch_service (
    branch_id BIGINT UNSIGNED NOT NULL,
    service_id BIGINT UNSIGNED NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (branch_id, service_id),
    CONSTRAINT fk_branch_service_branch FOREIGN KEY (branch_id) REFERENCES branch(id) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT fk_branch_service_service FOREIGN KEY (service_id) REFERENCES services(id) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
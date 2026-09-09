CREATE TABLE IF NOT EXISTS mm_service_payment (
    service_id BIGINT UNSIGNED NOT NULL,
    payment_type_id BIGINT UNSIGNED NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (service_id, payment_type_id),
    CONSTRAINT fk_service_payment_service FOREIGN KEY (service_id) REFERENCES services(id) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT fk_service_payment_payment_type FOREIGN KEY (payment_type_id) REFERENCES payment_type(id) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

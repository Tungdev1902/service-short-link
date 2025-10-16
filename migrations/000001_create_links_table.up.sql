-- Create links table with all required fields
CREATE TABLE IF NOT EXISTS `links` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `short_code` VARCHAR(10) NOT NULL,
    `original_url` TEXT NOT NULL,
    `title` VARCHAR(255) DEFAULT NULL,
    `description` TEXT DEFAULT NULL,
    `is_active` BOOLEAN NOT NULL DEFAULT TRUE,
    `expires_at` TIMESTAMP NULL DEFAULT NULL,
    `platform` VARCHAR(50) DEFAULT NULL COMMENT 'Platform information from JWT payload',
    `role` VARCHAR(50) DEFAULT NULL COMMENT 'User role from JWT payload',
    `channel_code` VARCHAR(50) DEFAULT NULL COMMENT 'Channel code from JWT payload',
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `click_count` BIGINT UNSIGNED NOT NULL DEFAULT 0,
    `last_accessed_at` TIMESTAMP NULL DEFAULT NULL,
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_short_code` (`short_code`),
    KEY `idx_is_active` (`is_active`),
    KEY `idx_expires_at` (`expires_at`),
    KEY `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

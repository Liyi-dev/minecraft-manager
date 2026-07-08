CREATE DATABASE IF NOT EXISTS minecraft_manager CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE minecraft_manager;

CREATE TABLE IF NOT EXISTS users (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(64) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    role VARCHAR(32) DEFAULT 'user',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS command_logs (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT NOT NULL,
    command VARCHAR(1024) NOT NULL,
    result TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS ban_records (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    uuid VARCHAR(64) NOT NULL,
    player_name VARCHAR(64),
    reason VARCHAR(512),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_uuid (uuid)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- Default admin user (password: admin123, bcrypt hash)
INSERT INTO users (username, password, role) VALUES
('admin', '$2a$10$Jobfa/jxmQbqyduasaRxNOQQUJ213GYwS/9OPeee8/JIUgbuMtf8C', 'admin')
ON DUPLICATE KEY UPDATE username=username;

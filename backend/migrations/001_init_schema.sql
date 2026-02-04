-- 迁移: 001_init_schema
-- 说明: 创建初始表结构

-- 用户表
CREATE TABLE IF NOT EXISTS users (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(20) DEFAULT 'user',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 系统配置表 (管理员配置的 Provider)
CREATE TABLE IF NOT EXISTS system_providers (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    provider_id VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    api_style VARCHAR(20) NOT NULL,
    api_host VARCHAR(255),
    api_key VARCHAR(500),
    enabled TINYINT(1) DEFAULT 1,
    allow_custom_key TINYINT(1) DEFAULT 0,
    models JSON,
    is_default TINYINT(1) DEFAULT 0,
    sort_order INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 创建索引
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_providers_enabled ON system_providers(enabled);
CREATE INDEX idx_providers_sort_order ON system_providers(sort_order);

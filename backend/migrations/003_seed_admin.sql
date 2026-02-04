-- 迁移: 003_seed_admin
-- 说明: 创建默认管理员账户
-- 默认用户名: admin
-- 默认密码: admin123 (首次登录需修改)
-- 密码哈希使用 bcrypt 生成: $2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy

INSERT INTO users (username, password_hash, role, password_changed)
SELECT 'admin', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'admin', 0
WHERE NOT EXISTS (SELECT 1 FROM users WHERE username = 'admin');

-- 迁移: 002_add_password_changed
-- 说明: 添加 password_changed 字段

ALTER TABLE users ADD COLUMN password_changed TINYINT(1) DEFAULT 0 AFTER role;

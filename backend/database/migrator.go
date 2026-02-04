package database

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type Migrator struct {
	db             *sql.DB
	migrationsPath string
}

type Migration struct {
	Version  int
	Filename string
	Checksum string
	Content  string
}

func NewMigrator(db *sql.DB, migrationsPath string) *Migrator {
	return &Migrator{
		db:             db,
		migrationsPath: migrationsPath,
	}
}

func (m *Migrator) Run() error {
	// 创建迁移记录表
	if err := m.createMigrationsTable(); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// 获取所有迁移文件
	migrations, err := m.loadMigrations()
	if err != nil {
		return fmt.Errorf("failed to load migrations: %w", err)
	}

	// 获取已执行的迁移
	executed, err := m.getExecutedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get executed migrations: %w", err)
	}

	// 检查已执行迁移的 checksum
	for version, checksum := range executed {
		for _, mig := range migrations {
			if mig.Version == version {
				if mig.Checksum != checksum {
					return fmt.Errorf("migration %s has been modified after execution (checksum mismatch)", mig.Filename)
				}
				break
			}
		}
	}

	// 执行新的迁移
	for _, mig := range migrations {
		if _, ok := executed[mig.Version]; ok {
			log.Printf("Migration %s already executed, skipping", mig.Filename)
			continue
		}

		log.Printf("Running migration: %s", mig.Filename)
		if err := m.executeMigration(mig); err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", mig.Filename, err)
		}
		log.Printf("Migration %s completed", mig.Filename)
	}

	return nil
}

func (m *Migrator) createMigrationsTable() error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INT PRIMARY KEY,
			filename VARCHAR(255) NOT NULL,
			checksum VARCHAR(64) NOT NULL,
			executed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4
	`
	_, err := m.db.Exec(query)
	return err
}

func (m *Migrator) loadMigrations() ([]Migration, error) {
	files, err := filepath.Glob(filepath.Join(m.migrationsPath, "*.sql"))
	if err != nil {
		return nil, err
	}

	var migrations []Migration
	for _, file := range files {
		filename := filepath.Base(file)

		// 解析版本号 (格式: 001_xxx.sql)
		parts := strings.SplitN(filename, "_", 2)
		if len(parts) < 2 {
			log.Printf("Skipping invalid migration filename: %s", filename)
			continue
		}

		version, err := strconv.Atoi(parts[0])
		if err != nil {
			log.Printf("Skipping invalid migration version: %s", filename)
			continue
		}

		// 读取文件内容
		content, err := os.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("failed to read migration file %s: %w", filename, err)
		}

		// 计算 checksum
		checksum := m.calculateChecksum(content)

		migrations = append(migrations, Migration{
			Version:  version,
			Filename: filename,
			Checksum: checksum,
			Content:  string(content),
		})
	}

	// 按版本号排序
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

func (m *Migrator) getExecutedMigrations() (map[int]string, error) {
	rows, err := m.db.Query("SELECT version, checksum FROM schema_migrations")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	executed := make(map[int]string)
	for rows.Next() {
		var version int
		var checksum string
		if err := rows.Scan(&version, &checksum); err != nil {
			return nil, err
		}
		executed[version] = checksum
	}

	return executed, rows.Err()
}

func (m *Migrator) executeMigration(mig Migration) error {
	// MySQL 不支持在事务中执行 DDL，分开执行每条语句
	statements := splitStatements(mig.Content)

	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" || strings.HasPrefix(stmt, "--") {
			continue
		}
		if _, err := m.db.Exec(stmt); err != nil {
			return fmt.Errorf("failed to execute SQL: %s, error: %w", stmt[:min(50, len(stmt))], err)
		}
	}

	// 记录迁移
	if _, err := m.db.Exec(
		"INSERT INTO schema_migrations (version, filename, checksum) VALUES (?, ?, ?)",
		mig.Version, mig.Filename, mig.Checksum,
	); err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	return nil
}

func (m *Migrator) calculateChecksum(content []byte) string {
	h := md5.New()
	io.WriteString(h, string(content))
	return hex.EncodeToString(h.Sum(nil))
}

// splitStatements 分割 SQL 语句
func splitStatements(content string) []string {
	var statements []string
	var current strings.Builder

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		// 跳过注释行
		if strings.HasPrefix(trimmed, "--") {
			continue
		}

		current.WriteString(line)
		current.WriteString("\n")

		// 以分号结尾表示语句结束
		if strings.HasSuffix(trimmed, ";") {
			statements = append(statements, current.String())
			current.Reset()
		}
	}

	// 处理最后没有分号的语句
	if current.Len() > 0 {
		stmt := strings.TrimSpace(current.String())
		if stmt != "" {
			statements = append(statements, stmt)
		}
	}

	return statements
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

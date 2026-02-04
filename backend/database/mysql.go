package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

func Init(cfg DBConfig) error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DBName)

	var err error
	// 重试连接，等待 MySQL 启动
	for i := 0; i < 30; i++ {
		DB, err = sql.Open("mysql", dsn)
		if err != nil {
			log.Printf("Failed to open database: %v, retrying...", err)
			time.Sleep(2 * time.Second)
			continue
		}

		err = DB.Ping()
		if err != nil {
			log.Printf("Failed to ping database: %v, retrying...", err)
			time.Sleep(2 * time.Second)
			continue
		}

		break
	}

	if err != nil {
		return fmt.Errorf("failed to connect to database after retries: %w", err)
	}

	// 设置连接池参数
	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(5)
	DB.SetConnMaxLifetime(5 * time.Minute)

	log.Printf("Database connected: %s:%s/%s", cfg.Host, cfg.Port, cfg.DBName)
	return nil
}

func Close() {
	if DB != nil {
		DB.Close()
	}
}

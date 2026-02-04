package main

import (
	"log"
	"net/http"

	"chatbox-backend/config"
	"chatbox-backend/database"
	"chatbox-backend/handlers"
	"chatbox-backend/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	cfg := config.Load()

	// 初始化数据库
	if err := database.Init(database.DBConfig{
		Host:     cfg.DBHost,
		Port:     cfg.DBPort,
		User:     cfg.DBUser,
		Password: cfg.DBPassword,
		DBName:   cfg.DBName,
	}); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// 运行数据库迁移
	migrator := database.NewMigrator(database.DB, "migrations")
	if err := migrator.Run(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// 创建默认管理员（如果不存在）
	if err := database.SeedDefaultAdmin(); err != nil {
		log.Fatalf("Failed to seed default admin: %v", err)
	}

	// 设置 Gin
	r := gin.Default()

	// CORS 配置
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// API 路由
	api := r.Group("/api")
	{
		// 认证相关
		auth := api.Group("/auth")
		{
			auth.POST("/login", handlers.Login)
			auth.GET("/me", middleware.AuthRequired(cfg.JWTSecret), handlers.GetCurrentUser)
			auth.POST("/change-password", middleware.AuthRequired(cfg.JWTSecret), handlers.ChangePassword)
		}

		// 配置相关 (公开)
		configGroup := api.Group("/config")
		{
			configGroup.GET("/providers", handlers.GetPublicProviders)
		}

		// 代理相关 (公开，用于非管理员使用系统配置的 EnterAI)
		proxy := api.Group("/proxy")
		{
			proxy.POST("/v1/chat/completions", handlers.ProxyChatCompletion)
			proxy.POST("/v1/images/generations", handlers.ProxyImageGeneration)
		}

		// 管理员相关 (需要管理员权限)
		admin := api.Group("/admin")
		admin.Use(middleware.AuthRequired(cfg.JWTSecret), middleware.AdminRequired())
		{
			admin.GET("/providers", handlers.AdminGetProviders)
			admin.POST("/providers", handlers.AdminCreateProvider)
			admin.PUT("/providers/:id", handlers.AdminUpdateProvider)
			admin.DELETE("/providers/:id", handlers.AdminDeleteProvider)
			admin.GET("/users", handlers.AdminGetUsers)
		}
	}

	// 启动服务器
	log.Printf("Server starting on port %s", cfg.ServerPort)
	if err := r.Run(":" + cfg.ServerPort); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

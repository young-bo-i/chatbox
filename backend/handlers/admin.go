package handlers

import (
	"net/http"
	"strconv"

	"chatbox-backend/models"

	"github.com/gin-gonic/gin"
)

type CreateProviderRequest struct {
	ProviderID     string                 `json:"providerId" binding:"required"`
	Name           string                 `json:"name" binding:"required"`
	APIStyle       string                 `json:"apiStyle" binding:"required"` // openai | google | anthropic
	APIHost        string                 `json:"apiHost"`
	APIKey         string                 `json:"apiKey"`
	Enabled        bool                   `json:"enabled"`
	AllowCustomKey bool                   `json:"allowCustomKey"`
	Models         []models.ProviderModel `json:"models"`
	IsDefault      bool                   `json:"isDefault"`
	SortOrder      int                    `json:"sortOrder"`
}

type UpdateProviderRequest struct {
	ProviderID     string                 `json:"providerId"`
	Name           string                 `json:"name"`
	APIStyle       string                 `json:"apiStyle"`
	APIHost        string                 `json:"apiHost"`
	APIKey         string                 `json:"apiKey"`
	Enabled        *bool                  `json:"enabled"`
	AllowCustomKey *bool                  `json:"allowCustomKey"`
	Models         []models.ProviderModel `json:"models"`
	IsDefault      *bool                  `json:"isDefault"`
	SortOrder      *int                   `json:"sortOrder"`
}

// GetPublicProviders 获取公开的 Provider 列表 (普通用户)
func GetPublicProviders(c *gin.Context) {
	providers, err := models.GetEnabledProviders()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get providers"})
		return
	}

	// 转换为公开格式 (隐藏 API Key)
	var publicProviders []models.PublicProvider
	for _, p := range providers {
		publicProviders = append(publicProviders, p.ToPublic())
	}

	c.JSON(http.StatusOK, gin.H{"providers": publicProviders})
}

// AdminGetProviders 获取所有 Provider (管理员)
func AdminGetProviders(c *gin.Context) {
	providers, err := models.GetAllProviders()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get providers"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"providers": providers})
}

// AdminCreateProvider 创建 Provider (管理员)
func AdminCreateProvider(c *gin.Context) {
	var req CreateProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	provider := &models.Provider{
		ProviderID:     req.ProviderID,
		Name:           req.Name,
		APIStyle:       req.APIStyle,
		APIHost:        req.APIHost,
		APIKey:         req.APIKey,
		Enabled:        req.Enabled,
		AllowCustomKey: req.AllowCustomKey,
		Models:         req.Models,
		IsDefault:      req.IsDefault,
		SortOrder:      req.SortOrder,
	}

	created, err := models.CreateProvider(provider)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create provider: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, created)
}

// AdminUpdateProvider 更新 Provider (管理员)
func AdminUpdateProvider(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid provider ID"})
		return
	}

	// 获取现有 Provider
	provider, err := models.GetProviderByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Provider not found"})
		return
	}

	var req UpdateProviderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// 更新字段
	if req.ProviderID != "" {
		provider.ProviderID = req.ProviderID
	}
	if req.Name != "" {
		provider.Name = req.Name
	}
	if req.APIStyle != "" {
		provider.APIStyle = req.APIStyle
	}
	if req.APIHost != "" {
		provider.APIHost = req.APIHost
	}
	if req.APIKey != "" {
		provider.APIKey = req.APIKey
	}
	if req.Enabled != nil {
		provider.Enabled = *req.Enabled
	}
	if req.AllowCustomKey != nil {
		provider.AllowCustomKey = *req.AllowCustomKey
	}
	if req.Models != nil {
		provider.Models = req.Models
	}
	if req.IsDefault != nil {
		provider.IsDefault = *req.IsDefault
	}
	if req.SortOrder != nil {
		provider.SortOrder = *req.SortOrder
	}

	if err := models.UpdateProvider(provider); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update provider"})
		return
	}

	// 重新获取更新后的 Provider
	updated, _ := models.GetProviderByID(id)
	c.JSON(http.StatusOK, updated)
}

// AdminDeleteProvider 删除 Provider (管理员)
func AdminDeleteProvider(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid provider ID"})
		return
	}

	// 检查 Provider 是否存在
	if _, err := models.GetProviderByID(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Provider not found"})
		return
	}

	if err := models.DeleteProvider(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete provider"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Provider deleted"})
}

// AdminGetUsers 获取用户列表 (管理员)
func AdminGetUsers(c *gin.Context) {
	users, err := models.GetAllUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get users"})
		return
	}

	// 转换为响应格式
	var responses []models.UserResponse
	for _, u := range users {
		responses = append(responses, u.ToResponse())
	}

	c.JSON(http.StatusOK, gin.H{"users": responses})
}

package models

import (
	"chatbox-backend/database"
	"encoding/json"
	"time"
)

type ProviderModel struct {
	ModelID       string   `json:"modelId"`
	Nickname      string   `json:"nickname,omitempty"`
	Type          string   `json:"type,omitempty"`          // chat | embedding | rerank
	APIStyle      string   `json:"apiStyle,omitempty"`      // openai | google | anthropic
	Labels        []string `json:"labels,omitempty"`
	Capabilities  []string `json:"capabilities,omitempty"`  // vision | reasoning | tool_use | web_search
	ContextWindow int      `json:"contextWindow,omitempty"` // 上下文窗口大小
	MaxOutput     int      `json:"maxOutput,omitempty"`     // 最大输出 tokens
}

type Provider struct {
	ID             int64           `json:"id"`
	ProviderID     string          `json:"providerId"`
	Name           string          `json:"name"`
	APIStyle       string          `json:"apiStyle"`
	APIHost        string          `json:"apiHost,omitempty"`
	APIKey         string          `json:"apiKey,omitempty"`
	Enabled        bool            `json:"enabled"`
	AllowCustomKey bool            `json:"allowCustomKey"`
	Models         []ProviderModel `json:"models,omitempty"`
	IsDefault      bool            `json:"isDefault"`
	SortOrder      int             `json:"sortOrder"`
	CreatedAt      time.Time       `json:"createdAt"`
	UpdatedAt      time.Time       `json:"updatedAt"`
}

// PublicProvider 公开的 Provider 信息 (隐藏 API Key)
type PublicProvider struct {
	ID             int64           `json:"id"`
	ProviderID     string          `json:"providerId"`
	Name           string          `json:"name"`
	APIStyle       string          `json:"apiStyle"`
	APIHost        string          `json:"apiHost,omitempty"`
	HasSystemKey   bool            `json:"hasSystemKey"` // 是否有系统配置的 Key
	AllowCustomKey bool            `json:"allowCustomKey"`
	Models         []ProviderModel `json:"models,omitempty"`
	IsDefault      bool            `json:"isDefault"`
	SortOrder      int             `json:"sortOrder"`
}

func (p *Provider) ToPublic() PublicProvider {
	return PublicProvider{
		ID:             p.ID,
		ProviderID:     p.ProviderID,
		Name:           p.Name,
		APIStyle:       p.APIStyle,
		APIHost:        p.APIHost,
		HasSystemKey:   p.APIKey != "",
		AllowCustomKey: p.AllowCustomKey,
		Models:         p.Models,
		IsDefault:      p.IsDefault,
		SortOrder:      p.SortOrder,
	}
}

// CreateProvider 创建新的 Provider
func CreateProvider(p *Provider) (*Provider, error) {
	modelsJSON, err := json.Marshal(p.Models)
	if err != nil {
		return nil, err
	}

	result, err := database.DB.Exec(`
		INSERT INTO system_providers 
		(provider_id, name, api_style, api_host, api_key, enabled, allow_custom_key, models, is_default, sort_order)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, p.ProviderID, p.Name, p.APIStyle, p.APIHost, p.APIKey,
		boolToInt(p.Enabled), boolToInt(p.AllowCustomKey),
		string(modelsJSON), boolToInt(p.IsDefault), p.SortOrder)
	if err != nil {
		return nil, err
	}

	id, _ := result.LastInsertId()
	return GetProviderByID(id)
}

// UpdateProvider 更新 Provider
func UpdateProvider(p *Provider) error {
	modelsJSON, err := json.Marshal(p.Models)
	if err != nil {
		return err
	}

	_, err = database.DB.Exec(`
		UPDATE system_providers SET
			provider_id = ?, name = ?, api_style = ?, api_host = ?, api_key = ?,
			enabled = ?, allow_custom_key = ?, models = ?, is_default = ?, sort_order = ?,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, p.ProviderID, p.Name, p.APIStyle, p.APIHost, p.APIKey,
		boolToInt(p.Enabled), boolToInt(p.AllowCustomKey),
		string(modelsJSON), boolToInt(p.IsDefault), p.SortOrder, p.ID)
	return err
}

// DeleteProvider 删除 Provider
func DeleteProvider(id int64) error {
	_, err := database.DB.Exec("DELETE FROM system_providers WHERE id = ?", id)
	return err
}

// GetProviderByID 根据 ID 获取 Provider
func GetProviderByID(id int64) (*Provider, error) {
	p := &Provider{}
	var modelsJSON string
	var enabled, allowCustomKey, isDefault int

	err := database.DB.QueryRow(`
		SELECT id, provider_id, name, api_style, api_host, api_key, enabled, 
			   allow_custom_key, models, is_default, sort_order, created_at, updated_at
		FROM system_providers WHERE id = ?
	`, id).Scan(&p.ID, &p.ProviderID, &p.Name, &p.APIStyle, &p.APIHost, &p.APIKey,
		&enabled, &allowCustomKey, &modelsJSON, &isDefault, &p.SortOrder, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}

	p.Enabled = enabled == 1
	p.AllowCustomKey = allowCustomKey == 1
	p.IsDefault = isDefault == 1

	if modelsJSON != "" {
		json.Unmarshal([]byte(modelsJSON), &p.Models)
	}

	return p, nil
}

// GetAllProviders 获取所有 Provider (管理员用)
func GetAllProviders() ([]Provider, error) {
	return queryProviders("SELECT id, provider_id, name, api_style, api_host, api_key, enabled, allow_custom_key, models, is_default, sort_order, created_at, updated_at FROM system_providers ORDER BY sort_order, id")
}

// GetEnabledProviders 获取启用的 Provider (普通用户用)
func GetEnabledProviders() ([]Provider, error) {
	return queryProviders("SELECT id, provider_id, name, api_style, api_host, api_key, enabled, allow_custom_key, models, is_default, sort_order, created_at, updated_at FROM system_providers WHERE enabled = 1 ORDER BY sort_order, id")
}

func queryProviders(query string) ([]Provider, error) {
	rows, err := database.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var providers []Provider
	for rows.Next() {
		var p Provider
		var modelsJSON string
		var enabled, allowCustomKey, isDefault int

		if err := rows.Scan(&p.ID, &p.ProviderID, &p.Name, &p.APIStyle, &p.APIHost, &p.APIKey,
			&enabled, &allowCustomKey, &modelsJSON, &isDefault, &p.SortOrder, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}

		p.Enabled = enabled == 1
		p.AllowCustomKey = allowCustomKey == 1
		p.IsDefault = isDefault == 1

		if modelsJSON != "" {
			json.Unmarshal([]byte(modelsJSON), &p.Models)
		}

		providers = append(providers, p)
	}

	return providers, rows.Err()
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

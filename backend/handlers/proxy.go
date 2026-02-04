package handlers

import (
	"bytes"
	"io"
	"net/http"
	"strings"

	"chatbox-backend/models"

	"github.com/gin-gonic/gin"
)

// ProxyChatCompletion 代理聊天完成请求
// 用于非管理员用户使用系统配置的 EnterAI
func ProxyChatCompletion(c *gin.Context) {
	// 获取 EnterAI 配置
	providers, err := models.GetEnabledProviders()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get provider configuration"})
		return
	}

	// 查找 EnterAI provider
	var enterAI *models.Provider
	for _, p := range providers {
		if p.ProviderID == "enter-ai" {
			enterAI = &p
			break
		}
	}

	if enterAI == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "EnterAI provider not configured"})
		return
	}

	if enterAI.APIKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "EnterAI API key not configured"})
		return
	}

	// 读取请求体
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}

	// 构建目标 URL
	apiHost := enterAI.APIHost
	if apiHost == "" {
		apiHost = "https://api.openai.com"
	}
	// 确保 apiHost 没有尾部斜杠
	apiHost = strings.TrimSuffix(apiHost, "/")
	
	// 构建完整的 API URL
	targetURL := apiHost + "/v1/chat/completions"

	// 创建代理请求
	proxyReq, err := http.NewRequest("POST", targetURL, bytes.NewReader(body))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create proxy request"})
		return
	}

	// 设置请求头
	proxyReq.Header.Set("Content-Type", "application/json")
	proxyReq.Header.Set("Authorization", "Bearer "+enterAI.APIKey)

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(proxyReq)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to connect to AI service: " + err.Error()})
		return
	}
	defer resp.Body.Close()

	// 检查是否是流式响应
	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "text/event-stream") {
		// 流式响应
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		c.Header("Transfer-Encoding", "chunked")
		
		c.Stream(func(w io.Writer) bool {
			buf := make([]byte, 1024)
			n, err := resp.Body.Read(buf)
			if err != nil {
				return false
			}
			w.Write(buf[:n])
			return true
		})
	} else {
		// 非流式响应
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response"})
			return
		}

		// 转发响应
		c.Data(resp.StatusCode, contentType, respBody)
	}
}

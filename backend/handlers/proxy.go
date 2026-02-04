package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"chatbox-backend/models"

	"github.com/gin-gonic/gin"
)

// getEnterAIProvider 获取 EnterAI 配置
func getEnterAIProvider() (*models.Provider, error) {
	providers, err := models.GetEnabledProviders()
	if err != nil {
		return nil, err
	}

	for _, p := range providers {
		if p.ProviderID == "enter-ai" {
			return &p, nil
		}
	}

	return nil, nil
}

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

// ProxyImageGeneration 代理图片生成请求
// 用于非管理员用户使用系统配置的 EnterAI 生成图片
func ProxyImageGeneration(c *gin.Context) {
	enterAI, err := getEnterAIProvider()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get provider configuration"})
		return
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

	// 打印原始请求体用于调试
	log.Printf("[ImageProxy] Original request body: %s", string(body))

	// 解析并转换请求体格式
	// ai-sdk 可能发送的格式与 OpenAI API 不完全兼容
	var requestData map[string]interface{}
	if err := json.Unmarshal(body, &requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// 如果 prompt 是对象类型，提取文本
	if prompt, ok := requestData["prompt"]; ok {
		log.Printf("[ImageProxy] Prompt type: %T, value: %v", prompt, prompt)
		switch p := prompt.(type) {
		case map[string]interface{}:
			// 如果是对象，尝试提取 text 字段
			if text, ok := p["text"].(string); ok {
				requestData["prompt"] = text
			}
		case []interface{}:
			// 如果是数组，提取第一个文本元素
			for _, item := range p {
				if itemMap, ok := item.(map[string]interface{}); ok {
					if text, ok := itemMap["text"].(string); ok {
						requestData["prompt"] = text
						break
					}
				} else if text, ok := item.(string); ok {
					requestData["prompt"] = text
					break
				}
			}
		}
	}

	// 重新序列化请求体
	convertedBody, err := json.Marshal(requestData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process request"})
		return
	}

	log.Printf("[ImageProxy] Converted request body: %s", string(convertedBody))

	// 构建目标 URL
	apiHost := enterAI.APIHost
	if apiHost == "" {
		apiHost = "https://api.openai.com"
	}
	apiHost = strings.TrimSuffix(apiHost, "/")
	targetURL := apiHost + "/v1/images/generations"

	// 创建代理请求
	proxyReq, err := http.NewRequest("POST", targetURL, bytes.NewReader(convertedBody))
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

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response"})
		return
	}

	// 转发响应
	contentType := resp.Header.Get("Content-Type")
	c.Data(resp.StatusCode, contentType, respBody)
}

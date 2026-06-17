package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"ai-content-platform/middleware"
	"ai-content-platform/services"

	"github.com/gin-gonic/gin"
)

// ========== 基础数据结构 ==========

// AIChatMessage 单条对话消息
type AIChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// AIChatRequest 前端发来的对话请求
type AIChatRequest struct {
	Messages    []AIChatMessage `json:"messages"`
	Stream      bool            `json:"stream"`
	ContentType string          `json:"content_type"` // article / poster / video
}

// AIConfigResponse 返回给前端的 LLM 配置摘要（不含 API Key）
type AIConfigResponse struct {
	Configured bool              `json:"configured"`
	Provider   string            `json:"provider"`
	BaseURL    string            `json:"base_url"`
	Model      string            `json:"model"`
	Error      string            `json:"error,omitempty"`
	Hint       map[string]string `json:"hint,omitempty"`
}

// ========== 1. /ai/config - 查询 LLM 配置状态 ==========

// GetAIConfig 返回当前后端的 LLM 配置摘要（敏感信息已脱敏）
func GetAIConfig(c *gin.Context) {
	svc := services.NewAIService()
	info := svc.ConfigInfo()

	var errorMsg string
	var modeText string
	if info["mode"] == "demo" {
		modeText = "演示模式"
		errorMsg = "当前为演示模式，返回预设的模拟回答。如需使用真实大模型，请设置 LLM_API_KEY 环境变量后重启服务。"
	} else {
		modeText = "已就绪"
	}

	resp := AIConfigResponse{
		Configured: svc.IsConfigured(),
		Provider:   info["provider"],
		BaseURL:    info["base_url"],
		Model:      info["model"] + " (" + modeText + ")",
		Error:      errorMsg,
		Hint: map[string]string{
			"env_LLM_PROVIDER": "siliconflow（默认，免费额度）/ deepseek / zhipu / ollama",
			"env_LLM_API_KEY":  "在 https://siliconflow.cn/ 免费注册获取",
			"env_LLM_BASE_URL": "https://api.siliconflow.cn/v1（默认）",
			"env_LLM_MODEL":    "Qwen/Qwen2.5-7B-Instruct（默认，完全免费）",
		},
	}
	c.JSON(http.StatusOK, resp)
}

// ========== 2. /ai/chat - 非流式对话 (一次性返回) ==========

// ChatWithAI 非流式对话接口
func ChatWithAI(c *gin.Context) {
	if _, ok := middleware.GetUserFromContext(c); !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}
	var req AIChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求体格式错误: " + err.Error()})
		return
	}
	if len(req.Messages) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "messages 不能为空"})
		return
	}

	svc := services.NewAIService()
	msgs := convertMessages(req.Messages, svc)

	reply, err := svc.Chat(msgs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"reply":   "",
			"success": false,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"reply":   reply,
		"success": true,
	})
}

// ========== 3. /ai/chat/stream - 流式对话 (SSE，打字机效果) ==========

// ChatWithAIStream 流式对话接口：服务端主动推送增量文本
func ChatWithAIStream(c *gin.Context) {
	if _, ok := middleware.GetUserFromContext(c); !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
		return
	}
	var req AIChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求体格式错误: " + err.Error()})
		return
	}
	if len(req.Messages) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "messages 不能为空"})
		return
	}

	svc := services.NewAIService()
	msgs := convertMessages(req.Messages, svc)

	// 设置 SSE headers
	c.Writer.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")

	// 立即发送 "开始" 事件
	writeSSEEvent(c, "start", map[string]interface{}{"time": time.Now().Unix()})

	// 调用流式 LLM
	fullText, err := svc.StreamChat(msgs, func(delta string) {
		writeSSEEvent(c, "delta", map[string]interface{}{"content": delta})
	})
	if err != nil {
		writeSSEEvent(c, "error", map[string]interface{}{"message": err.Error()})
		return
	}
	// 发送 "结束" 事件
	writeSSEEvent(c, "done", map[string]interface{}{"full": fullText})
}

// ========== 4. /ai/generate-content - 直接基于 prompt 生成内容（非流式，用于内容管理） ==========

// 保持与旧版 content.go 中的 GenerateContentWithAI 类似的行为，但调用真实 LLM
// （GenerateContentWithAI 已在 content.go 中存在，这里不再重复）

// ========== 工具函数 ==========

// convertMessages 把前端消息列表转换为 services.ChatMessage，并在没有 system 消息时注入默认 system prompt
func convertMessages(front []AIChatMessage, svc *services.AIService) []services.ChatMessage {
	out := make([]services.ChatMessage, 0, len(front)+1)
	// 若第一条不是 system，则注入默认 system
	hasSystem := false
	if len(front) > 0 && strings.ToLower(front[0].Role) == "system" {
		hasSystem = true
	}
	if !hasSystem {
		out = append(out, services.ChatMessage{
			Role:    "system",
			Content: svc.SystemPromptOrDefault(),
		})
	}
	for _, m := range front {
		role := strings.ToLower(m.Role)
		if role != "user" && role != "assistant" && role != "system" {
			role = "user"
		}
		out = append(out, services.ChatMessage{
			Role:    role,
			Content: m.Content,
		})
	}
	return out
}

// writeSSEEvent 写一条 SSE 事件：event: xxx\ndata: {...}\n\n
func writeSSEEvent(c *gin.Context, event string, payload map[string]interface{}) {
	b, _ := json.Marshal(payload)
	line := fmt.Sprintf("event: %s\ndata: %s\n\n", event, string(b))
	if _, err := c.Writer.Write([]byte(line)); err != nil {
		return
	}
	if flusher, ok := c.Writer.(http.Flusher); ok {
		flusher.Flush()
	}
}

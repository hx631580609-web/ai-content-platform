package config

import (
	"os"
	"strings"
)

// Config holds application-wide configuration loaded from environment variables.
type Config struct {
	ServerPort string
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	JWTSecret  string

	// ---------- LLM (Large Language Model) 相关配置 ----------
	// Provider 名称，仅用于 UI 展示与区分。所有提供商统一使用 OpenAI 兼容接口。
	// 示例: siliconflow（默认，免费额度最易获取）、deepseek、zhipu、qwen、ollama (本地) 等
	LLMProvider string
	// API Key 用于鉴权。留空时将启用演示模式（返回预设的模拟回答，系统仍可完整演示使用。
	LLMAPIKey string
	// 基础 URL，需包含协议与主机。默认使用 SiliconFlow (https://siliconflow.cn/)
	LLMBaseURL string
	// 模型名称。默认使用 SiliconFlow 的 Qwen/Qwen2.5-7B-Instruct（完全免费）。
	// 其他推荐的免费/低成本模型：
	//   siliconflow: Qwen/Qwen2.5-7B-Instruct, deepseek-ai/DeepSeek-R1-Distill-Qwen-7B
	//   deepseek:   deepseek-chat
	//   zhipu:      glm-4-flash（有免费额度）
	//   ollama(本地): qwen2.5:7b 等
	LLMModel string
	// 单次请求最大 token 数 (0 表示使用服务端默认值)
	LLMMaxTokens int
	// 采样温度，0-1，越高越有创造性 (0 表示使用默认 0.7)
	LLMTemperature float64
	// 系统提示词，可以覆盖默认的"你是专业内容创作助手"
	LLMSystemPrompt string
}

// LoadConfig 从环境变量中加载配置，对缺失项提供合理默认值。
// 免费模型说明：
//   - 默认使用 SiliconFlow 的 Qwen2.5-7B-Instruct（完全免费，需在 https://siliconflow.cn/ 注册获取 API Key）
//   - 如未设置 LLM_API_KEY，将启用演示模式，系统会使用预设的模拟回答（功能完整可演示）
func LoadConfig() *Config {
	return &Config{
		ServerPort:      getEnv("SERVER_PORT", "8080"),
		DBHost:          getEnv("DB_HOST", "localhost"),
		DBPort:          getEnv("DB_PORT", "3306"),
		DBUser:          getEnv("DB_USER", "root"),
		DBPassword:      getEnv("DB_PASSWORD", "123456"),
		DBName:          getEnv("DB_NAME", "ai_content_platform"),
		JWTSecret:       getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
		LLMProvider:     getEnv("LLM_PROVIDER", "siliconflow"),
		LLMAPIKey:       getEnv("LLM_API_KEY", ""),
		LLMBaseURL:      getEnv("LLM_BASE_URL", "https://api.siliconflow.cn/v1"),
		LLMModel:        getEnv("LLM_MODEL", "Qwen/Qwen2.5-7B-Instruct"),
		LLMMaxTokens:    getEnvInt("LLM_MAX_TOKENS", 0),
		LLMTemperature:  getEnvFloat("LLM_TEMPERATURE", 0.7),
		LLMSystemPrompt: getEnv("LLM_SYSTEM_PROMPT", ""),
	}
}

// IsDemoMode 当未配置 API Key 时视为演示模式
func (c *Config) IsDemoMode() bool {
	return strings.TrimSpace(c.LLMAPIKey) == ""
}

// DefaultSystemPrompt 返回一个对沙特签证/中东商旅主题友好的默认系统提示词。
func (c *Config) DefaultSystemPrompt() string {
	if strings.TrimSpace(c.LLMSystemPrompt) != "" {
		return c.LLMSystemPrompt
	}
	return "你是一个专业的中文内容创作助手，擅长撰写关于沙特阿拉伯商务签证、中东商旅、出行指南、政策解读等主题的高质量文章。请使用清晰、专业的中文，避免杜撰具体的签证材料清单或政策数字——如果不确定，明确告知用户以官方渠道为准。回复应当结构清晰，可包含小标题、要点列表等。"
}

// IsConfigured 当配置了 API Key 或处于演示模式时视为已配置。
// 演示模式下不调用真实 LLM，返回预设的模拟回答。
func (c *Config) IsConfigured() bool {
	// 演示模式也算已配置
	return strings.TrimSpace(c.LLMBaseURL) != ""
}

// getEnv 返回环境变量，如不存在则返回默认值。
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt 读取整型环境变量，失败时返回默认值。
func getEnvInt(key string, defaultValue int) int {
	raw := os.Getenv(key)
	if raw == "" {
		return defaultValue
	}
	var n int
	// 轻量解析，不引入额外依赖
	for _, ch := range raw {
		if ch >= '0' && ch <= '9' {
			n = n*10 + int(ch-'0')
		}
	}
	if n == 0 {
		return defaultValue
	}
	return n
}

// getEnvFloat 读取浮点型环境变量，失败时返回默认值。
func getEnvFloat(key string, defaultValue float64) float64 {
	raw := os.Getenv(key)
	if raw == "" {
		return defaultValue
	}
	// 简单解析：支持 "0.7" "1" 等
	var f float64
	var neg bool
	idx := 0
	if idx < len(raw) && (raw[idx] == '-' || raw[idx] == '+') {
		neg = raw[idx] == '-'
		idx++
	}
	intPart := 0
	for idx < len(raw) && raw[idx] >= '0' && raw[idx] <= '9' {
		intPart = intPart*10 + int(raw[idx]-'0')
		idx++
	}
	f = float64(intPart)
	if idx < len(raw) && raw[idx] == '.' {
		idx++
		frac := 0.0
		div := 1.0
		for idx < len(raw) && raw[idx] >= '0' && raw[idx] <= '9' {
			frac = frac*10 + float64(raw[idx]-'0')
			div *= 10
			idx++
		}
		f += frac / div
	}
	if neg {
		f = -f
	}
	if f == 0 && defaultValue != 0 {
		return defaultValue
	}
	return f
}

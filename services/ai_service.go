package services

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"ai-content-platform/config"
	"ai-content-platform/models"
)

// ============================================================
//  ChatMessage / ChatRequest - OpenAI 兼容接口的数据结构
// ============================================================

// ChatMessage 表示单条对话消息
type ChatMessage struct {
	Role    string `json:"role"`    // system / user / assistant
	Content string `json:"content"` // 消息内容
}

// ChatRequest 对应 OpenAI /chat/completions 请求体
type ChatRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Temperature float64       `json:"temperature,omitempty"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Stream      bool          `json:"stream,omitempty"`
}

// ChatResponse 对应 OpenAI 非流式响应
type ChatResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Choices []struct {
		Index        int         `json:"index"`
		Message      ChatMessage `json:"message"`
		FinishReason string      `json:"finish_reason"`
	} `json:"choices"`
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

// ChatStreamChunk 对应 OpenAI 流式响应的一条数据块 (data: {...})
type ChatStreamChunk struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
			Role    string `json:"role"`
		} `json:"delta"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
}

// ============================================================
//  AIService - 对外服务
// ============================================================

type AIService struct {
	APIKey       string
	BaseURL      string
	Model        string
	Provider     string
	MaxTokens    int
	Temperature  float64
	SystemPrompt string
	HTTPClient   *http.Client
	DemoMode     bool // 演示模式：未配置 API Key 时返回模拟回答
}

// NewAIService 从全局配置创建服务实例
func NewAIService() *AIService {
	cfg := config.LoadConfig()
	return &AIService{
		APIKey:       cfg.LLMAPIKey,
		BaseURL:      strings.TrimRight(cfg.LLMBaseURL, "/"),
		Model:        cfg.LLMModel,
		Provider:     cfg.LLMProvider,
		MaxTokens:    cfg.LLMMaxTokens,
		Temperature:  cfg.LLMTemperature,
		SystemPrompt: cfg.DefaultSystemPrompt(),
		HTTPClient: &http.Client{
			Timeout: 5 * time.Minute,
		},
		DemoMode: cfg.IsDemoMode(),
	}
}

// IsConfigured 返回是否已配置好可调用（演示模式也视为已配置）
func (ai *AIService) IsConfigured() bool {
	// 演示模式或真实模式皆可
	return strings.TrimSpace(ai.BaseURL) != ""
}

// ConfigInfo 返回当前配置信息，用于 UI 展示
func (ai *AIService) ConfigInfo() map[string]string {
	info := map[string]string{
		"provider": ai.Provider,
		"base_url": ai.BaseURL,
		"model":    ai.Model,
	}
	if ai.DemoMode {
		info["mode"] = "demo"
	}
	return info
}

// SystemPromptOrDefault 返回 system prompt，若为空则使用默认中文助手提示
func (ai *AIService) SystemPromptOrDefault() string {
	if strings.TrimSpace(ai.SystemPrompt) != "" {
		return ai.SystemPrompt
	}
	return "你是一个专业的中文 AI 助手，擅长撰写沙特签证、中东商旅、出行指南与政策解读类内容。回答清晰有条理，必要时使用小标题和要点。"
}

// ==================== 演示模式核心逻辑 ====================

// demoReply 根据最后一条用户消息的关键词，生成符合上下文的模拟回答
func demoReply(userInput string) string {
	in := strings.ToLower(strings.TrimSpace(userInput))

	// --- 通用问候 ---
	switch {
	case containsAny(in, "你好", "hello", "hi", "您好", "嗨", "在吗"):
		return "您好！我是 AI 创作助手。您可以告诉我一个主题（如沙特商务签证、中东商旅指南等），我会帮您生成专业内容。\n\n💡 当前处于演示模式（未配置 API Key）。如需使用真实大模型，请设置 LLM_API_KEY 环境变量。"

	case containsAny(in, "你是谁", "你是", "介绍你", "who are"):
		return "我是一个集成在内容分发平台中的 AI 创作助手。\n\n我的能力包括：\n- 文章撰写（300-800 字专业正文）\n- 海报文案（标题+要点列表）\n- 短视频脚本（开场+要点+结尾）\n- 自动打标签、分类、生成摘要\n\n💡 演示模式说明：当前使用预设回答。要使用真实大模型，请在启动前设置 LLM_API_KEY（推荐 SiliconFlow，Qwen2.5-7B-Instruct 完全免费）。"

	case containsAny(in, "演示", "demo", "测试", "示例", "怎么用", "使用", "功能"):
		return "以下是系统功能演示：\n\n✅ 内容创作：输入主题即可生成文章/海报/脚本\n✅ 智能分类：自动识别主题归入（沙特签证 / 中东商旅 / 出行指南 / 政策解读）\n✅ 内容分发：一键适配微信公众号 / 小红书 / CMS 等多平台\n\n如何接入真实大模型：\n1. 访问 https://siliconflow.cn/ 注册（免费）\n2. 复制你的 API Key\n3. 设置环境变量：LLM_API_KEY=你的key\n4. 重启服务\n\n默认模型：Qwen/Qwen2.5-7B-Instruct（完全免费）"

	case containsAny(in, "商务签证", "business visa"):
		return "【沙特商务签证概览】\n\n一、签证类型与适用人群\n沙特商务签证主要用于赴沙参加会议、洽谈业务、签订合同等商业活动。常见签证类别包括单次入境、多次入境及电子签证（E-Visa）。\n\n二、申请基本材料\n1. 有效期 6 个月以上的护照\n2. 近期白底彩色证件照\n3. 沙特方邀请函或会议通知\n4. 在职证明与公司营业执照\n*材料清单以官方最新要求为准，建议出行前 4-6 周开始准备*。\n\n三、办理方式与时效\n可通过沙特驻华使领馆或官方指定代办机构申请，电子签证审批通常 3-7 个工作日。部分国家公民可通过 Visit Saudi 平台在线申请。\n\n四、出行建议\n- 女性需遵守着装规范（Abaya）\n- 周五为公共假日，政府与银行休息\n- 商务会晤建议提前预约并准时到达\n\n*本内容由演示模式生成。如需更专业的建议，请在真实大模型中咨询。*"

	case containsAny(in, "旅游签证", "旅游签", "tourist"):
		return "【沙特旅游签证指南】\n\n一、签证概述\n沙特自 2019 年起向全球 49 个国家开放旅游电子签证，中国公民含港澳台在内均可办理。\n\n二、申请方式\n1. 在线申请（visitvisa.visitsaudi.com）\n2. 抵达利雅得/吉达机场后落地签\n3. 通过中国代办机构办理\n\n三、主要旅游城市推荐\n- 利雅得：王国中心、国家博物馆、老城区\n- 吉达：滨海大道、老城 Al-Balad\n- 麦加/麦地那：宗教圣地（穆斯林方可进入）\n- 欧拉：世界文化遗产、沙漠岩雕\n\n四、实用贴士\n- 夏季气温可达 45°C，推荐 11 月-3 月出行\n- 公共交通不发达，建议租车或打车\n- 当地使用里亚尔（SAR），1 SAR ≈ 1.85 RMB\n\n*演示模式内容仅供参考，请以官方信息为准。*"

	case containsAny(in, "中东", "商旅", "商务旅行", "出行指南"):
		return "【中东商务旅行实用指南】\n\n一、行前准备\n- 签证：提前确认目的地签证政策，部分国家支持落地签或电子签\n- 服装：男士长袖衬衫+长裤；女士需备 Abaya 或宽松服装\n- 电源：英标/欧标插座，电压 220V\n\n二、商务礼仪要点\n1. 会面：提前预约，准时到达\n2. 问候：使用右手，握手有力\n3. 交谈：避免讨论宗教、政治敏感话题\n4. 用餐：穆斯林不食猪肉、不饮酒\n\n三、主要商务城市\n- 迪拜（阿联酋）：中东商业枢纽，自由区政策\n- 利雅得（沙特）：新兴科技与金融中心\n- 多哈（卡塔尔）：体育、会展经济\n\n四、常见挑战与应对\n- 语言：阿拉伯语为主，英语在商务圈通用\n- 工作时间：周日-周四，周五周六休息\n- 交通：出租车/Uber 方便，女性建议网约车\n\n*演示模式生成内容，具体业务请以专业咨询为准。*"

	case containsAny(in, "政策", "新规", "签证政策", "更新", "change"):
		return "【2026 沙特签证政策解读】\n\n一、近期政策变化\n- 电子签证范围持续扩大，更多国家公民可在线申请\n- 多次入境签证有效期延长\n- 停留天数上限调整（以申请时官方说明为准）\n\n二、关注要点\n1. Visit Saudi 平台（visitsaudi.com）始终是最权威信息来源\n2. 签证申请建议提前 4 周准备\n3. 往返机票与酒店预订单为常见要求\n\n三、中国公民适用政策\n- 支持旅游电子签证和落地签证\n- 商务签证仍需邀请函或企业担保\n- 停留期：旅游签通常 90 天内\n\n四、建议行动\n1. 出行前 1 周登录官方平台核实最新政策\n2. 确认护照有效期 6 个月以上\n3. 购买涵盖医疗的境外旅行保险\n\n*演示模式内容不构成法律建议，具体以官方公告为准。*"

	case containsAny(in, "海报", "poster", "标题"):
		return "【海报文案示例】\n\n🔥 标题：中东商务新机遇——沙特市场深度解析\n\n✅ 要点：\n• 3 万亿美元经济体，年轻人口红利\n• 投资自由区政策，外资可 100% 控股\n• 数字经济 / 新能源 / 旅游业三大赛道\n• Riyadh Season / Qiddiya 等大型项目持续释放\n\n📞 更多信息：请联系我们获取深度报告\n\n*这是演示模式输出的海报示例。接入真实大模型后可根据您的具体主题生成更精准内容。*"

	case containsAny(in, "脚本", "视频", "短视频", "抖音", "script", "video"):
		return "【短视频脚本示例】\n\n【开场】\n“你知道吗？去沙特出差比你想象的更简单！今天 1 分钟告诉你如何搞定商务签证。”\n\n【要点】\n1. 先确认邀请函——没有邀请，商务签证基本不受理\n2. 在线申请 Visit Saudi 电子签证，3-7 天出签\n3. 女性服装注意：长袖+长裤或备 Abaya\n4. 现金+信用卡并行，里亚尔与人民币约 1:1.85\n\n【结尾】\n“关注我们，中东商务出行不再踩坑。有问题评论区见！”\n\n*这是演示模式生成的脚本示例。接入真实大模型后可针对您的具体主题生成剧本化内容。*"

	case containsAny(in, "模型", "llm", "api", "key", "怎么配置", "设置", "配置"):
		return "【如何配置真实大模型】\n\n步骤一：获取免费 API Key\n推荐使用 SiliconFlow（完全免费，支持 Qwen2.5-7B-Instruct、DeepSeek-R1 等开源大模型）\n网址：https://siliconflow.cn/\n\n步骤二：设置环境变量\n```\nLLM_PROVIDER=siliconflow\nLLM_API_KEY=你的APIKey\nLLM_BASE_URL=https://api.siliconflow.cn/v1\nLLM_MODEL=Qwen/Qwen2.5-7B-Instruct\n```\n\n步骤三：重启服务\n重启后刷新页面，配置状态会显示为「已就绪」。\n\n其他可选服务商：\n- DeepSeek：deepseek-chat（首月有免费额度）\n- 智谱 AI：glm-4-flash（有免费调用额度）\n- Ollama 本地：qwen2.5:7b（完全离线，本地部署）"

	default:
		// 通用文章模式：根据输入生成结构化回答
		topic := strings.TrimSpace(userInput)
		// 按 rune 截断，避免中文乱码
		topicRunes := []rune(topic)
		if len(topicRunes) > 50 {
			topic = string(topicRunes[:50]) + "..."
		}
		return fmt.Sprintf(
			"【主题：%s】\n\n一、背景概述\n该主题涉及多个层面的分析，建议从核心概念入手，理清基本定义与适用场景。\n\n二、主要要点\n1. 明确目标与适用人群\n2. 梳理关键数据与证据\n3. 关注政策与环境的最新变化\n4. 结合实际案例进行分析\n5. 给出可操作的建议\n\n三、实用建议\n- 信息搜集：优先参考官方渠道与权威来源\n- 时间规划：提前 4-6 周准备关键材料\n- 风险管理：准备备选方案应对不可预见因素\n\n四、延伸阅读\n可结合同类主题的专业报告与案例，形成更全面的理解。\n\n*本内容由演示模式自动生成。接入真实大模型后，将输出更具针对性的内容。（原始主题：%s）*",
			topic, topic,
		)
	}
}

// containsAny 判断字符串是否包含任意关键词
func containsAny(s string, keywords ...string) bool {
	for _, k := range keywords {
		if strings.Contains(s, strings.ToLower(k)) {
			return true
		}
	}
	return false
}

// ============================================================
//  底层 HTTP 请求封装
// ============================================================

// chatEndpoint 拼接 chat completions 端点
func (ai *AIService) chatEndpoint() string {
	return ai.BaseURL + "/chat/completions"
}

// doJSON 发送一次 JSON 请求并返回响应体
func (ai *AIService) doJSON(reqBody []byte) (*http.Response, error) {
	httpReq, err := http.NewRequest("POST", ai.chatEndpoint(), bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+ai.APIKey)
	// 部分提供商 (如 Azure OpenAI) 使用 api-key 头；同时兼容
	httpReq.Header.Set("api-key", ai.APIKey)

	resp, err := ai.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("调用 LLM 失败: %w", err)
	}
	return resp, nil
}

// ============================================================
//  非流式对话（一次性返回）
// ============================================================

// Chat 发送一次对话：基于多轮对话消息列表生成文本
// 在演示模式下不调用真实 LLM，返回模拟回答
func (ai *AIService) Chat(messages []ChatMessage) (string, error) {
	if !ai.IsConfigured() {
		return "", fmt.Errorf("LLM 尚未配置：请设置环境变量 LLM_API_KEY 与 LLM_BASE_URL")
	}

	// 演示模式：直接返回模拟回答
	if ai.DemoMode {
		// 取最后一条 user 消息作为输入
		userInput := ""
		for i := len(messages) - 1; i >= 0; i-- {
			if messages[i].Role == "user" {
				userInput = messages[i].Content
				break
			}
		}
		return demoReply(userInput), nil
	}

	body := ChatRequest{
		Model:       ai.Model,
		Messages:    messages,
		Temperature: ai.Temperature,
		MaxTokens:   ai.MaxTokens,
		Stream:      false,
	}
	reqBytes, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	resp, err := ai.doJSON(reqBytes)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("LLM 返回错误 (HTTP %d): %s", resp.StatusCode, string(b))
	}

	var chatResp ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return "", fmt.Errorf("解析 LLM 响应失败: %w", err)
	}
	if chatResp.Error.Message != "" {
		return "", fmt.Errorf("LLM 接口错误: %s", chatResp.Error.Message)
	}
	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("LLM 未返回任何内容")
	}
	return chatResp.Choices[0].Message.Content, nil
}

// ============================================================
//  流式对话（Server-Sent Events) —— 适合前端打字机效果
// ============================================================

// StreamChat 以流式方式进行对话，onChunk 每次收到增量文本就被调用一次。
// 返回完整文本作为函数返回值（累计）
// 在演示模式下，会将模拟回答逐字流式推送，模拟打字机效果
func (ai *AIService) StreamChat(messages []ChatMessage, onChunk func(delta string)) (string, error) {
	if !ai.IsConfigured() {
		return "", fmt.Errorf("LLM 尚未配置：请设置环境变量 LLM_API_KEY 与 LLM_BASE_URL")
	}

	// 演示模式：生成模拟回答并逐字流式推送
	if ai.DemoMode {
		userInput := ""
		for i := len(messages) - 1; i >= 0; i-- {
			if messages[i].Role == "user" {
				userInput = messages[i].Content
				break
			}
		}
		fullText := demoReply(userInput)
		runes := []rune(fullText)
		// 以 2-4 个字符为一块模拟流式输出（按 rune 切分避免中文乱码）
		for i := 0; i < len(runes); {
			chunkSize := 2 + randInt(0, 2)
			if i+chunkSize > len(runes) {
				chunkSize = len(runes) - i
			}
			chunk := string(runes[i : i+chunkSize])
			if onChunk != nil {
				onChunk(chunk)
			}
			time.Sleep(20 * time.Millisecond)
			i += chunkSize
		}
		return fullText, nil
	}

	body := ChatRequest{
		Model:       ai.Model,
		Messages:    messages,
		Temperature: ai.Temperature,
		MaxTokens:   ai.MaxTokens,
		Stream:      true,
	}
	reqBytes, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	resp, err := ai.doJSON(reqBytes)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("LLM 返回错误 (HTTP %d): %s", resp.StatusCode, string(b))
	}

	var fullBuilder strings.Builder
	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			break
		}
		line = strings.TrimSpace(line)
		if line == "" {
			if err == io.EOF {
				break
			}
			continue
		}
		// SSE 格式: "data: {...}"
		if !strings.HasPrefix(line, "data:") {
			if err == io.EOF {
				break
			}
			continue
		}
		payload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if payload == "[DONE]" {
			break
		}
		var chunk ChatStreamChunk
		if jerr := json.Unmarshal([]byte(payload), &chunk); jerr != nil {
			// 忽略无法解析的块
			if err == io.EOF {
				break
			}
			continue
		}
		if len(chunk.Choices) > 0 {
			delta := chunk.Choices[0].Delta.Content
			if delta != "" {
				fullBuilder.WriteString(delta)
				if onChunk != nil {
					onChunk(delta)
				}
			}
		}
		if err == io.EOF {
			break
		}
	}
	return fullBuilder.String(), nil
}

// randInt 生成指定范围的随机整数（简易实现，不引入额外依赖）
func randInt(min, max int) int {
	return min + int(time.Now().UnixNano())%(max-min+1)
}

// ============================================================
//  原有高层 API (保留签名，内部改为真实 LLM)
// ============================================================

// normalizeTopic 从 prompt 做简单关键词识别，仅用于兜底显示
func normalizeTopic(prompt string) string {
	topic := strings.TrimSpace(prompt)
	if topic == "" {
		return "通用内容"
	}
	switch {
	case strings.Contains(prompt, "商务签证"), strings.Contains(prompt, "商务签"):
		return "沙特商务签证"
	case strings.Contains(prompt, "旅游签证"), strings.Contains(prompt, "旅游签"):
		return "沙特旅游签证"
	case strings.Contains(prompt, "工作签证"):
		return "沙特工作签证"
	case strings.Contains(prompt, "探亲"):
		return "沙特探亲签证"
	case strings.Contains(prompt, "过境"):
		return "沙特过境签证"
	case strings.Contains(prompt, "签证"):
		return "沙特签证"
	case strings.Contains(prompt, "迪拜"), strings.Contains(prompt, "阿联酋"):
		return "阿联酋签证"
	case strings.Contains(prompt, "商旅"), strings.Contains(prompt, "商务出行"), strings.Contains(prompt, "中东"):
		return "中东商旅"
	case strings.Contains(prompt, "出行"), strings.Contains(prompt, "攻略"), strings.Contains(prompt, "指南"):
		return "出行指南"
	case strings.Contains(prompt, "政策"), strings.Contains(prompt, "新规"):
		return "政策解读"
	default:
		if len(topic) > 40 {
			return topic[:40]
		}
		return topic
	}
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// GenerateTextContent 使用真实 LLM 生成正文，根据 contentType 决定输出格式
func (ai *AIService) GenerateTextContent(prompt string, contentType models.ContentType) (string, error) {
	systemMsg := ChatMessage{Role: "system", Content: ai.SystemPrompt}

	var userPrompt string
	switch contentType {
	case models.Poster:
		userPrompt = fmt.Sprintf(
			"请为以下主题生成一段适合做海报内容的简短文案（标题+要点列表）。主题：%s\n\n输出格式：\n【标题】\n(一行标题)\n\n【要点】\n- 要点1\n- 要点2\n- 要点3\n\n内容简洁有力，不超过 200 字。",
			prompt,
		)
	case models.Video:
		userPrompt = fmt.Sprintf(
			"请为以下主题生成一个短视频脚本（开场白+要点+结尾）。主题：%s。格式：\n\n【开场】\n...\n\n【要点】\n1. ...\n2. ...\n3. ...\n\n【结尾】\n...",
			prompt,
		)
	default: // Article
		userPrompt = fmt.Sprintf(
			"请围绕以下用户输入或主题生成一篇 300-800 字的中文正文。使用小标题+要点，结构清晰，语言专业流畅。主题：%s",
			prompt,
		)
	}

	return ai.Chat([]ChatMessage{systemMsg, {Role: "user", Content: userPrompt}})
}

// GenerateTitleFromPrompt 让 LLM 写标题
func (ai *AIService) GenerateTitleFromPrompt(prompt string) (string, error) {
	systemMsg := ChatMessage{Role: "system", Content: ai.SystemPrompt}
	userPrompt := fmt.Sprintf(
		"请为以下内容生成一个不超过 30 字的中文标题，不要输出其他文字，只输出标题本身：\n\n%s",
		prompt,
	)
	title, err := ai.Chat([]ChatMessage{systemMsg, {Role: "user", Content: userPrompt}})
	if err != nil {
		return "", err
	}
	title = strings.TrimSpace(title)
	title = strings.Trim(title, "“”\"'")
	if title == "" {
		title = normalizeTopic(prompt)
	}
	return title, nil
}

// GenerateSummary 让 LLM 写摘要
func (ai *AIService) GenerateSummary(prompt string, content string) (string, error) {
	systemMsg := ChatMessage{Role: "system", Content: ai.SystemPrompt}
	src := content
	if src == "" {
		src = prompt
	}
	userPrompt := fmt.Sprintf(
		"请为以下内容写一段 80-120 字的中文摘要：\n\n%s",
		truncate(src, 2000),
	)
	summary, err := ai.Chat([]ChatMessage{systemMsg, {Role: "user", Content: userPrompt}})
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(summary), nil
}

// GenerateMetaTags 让 LLM 生成 3-5 个以顿号分隔的中文关键词
func (ai *AIService) GenerateMetaTags(prompt string) (string, error) {
	systemMsg := ChatMessage{Role: "system", Content: ai.SystemPrompt}
	userPrompt := fmt.Sprintf(
		"请为以下内容生成 3-5 个中文关键词/标签，使用顿号分隔，不要输出其他文字：\n\n%s",
		truncate(prompt, 1500),
	)
	tags, err := ai.Chat([]ChatMessage{systemMsg, {Role: "user", Content: userPrompt}})
	if err != nil {
		return "", err
	}
	tags = strings.TrimSpace(tags)
	tags = strings.Trim(tags, "。.，,、")
	return tags, nil
}

// DetectCategory 让 LLM 分类：从候选类别 沙特签证 / 中东商旅 / 出行指南 / 政策解读
func (ai *AIService) DetectCategory(prompt string) (string, error) {
	systemMsg := ChatMessage{Role: "system", Content: ai.SystemPrompt}
	userPrompt := fmt.Sprintf(
		"请从以下候选类别中为给定文本选择最合适的一个分类，只输出类别名，不要输出其他文字或标点。\n\n候选类别：沙特签证、中东商旅、出行指南、政策解读\n\n文本：\n%s",
		truncate(prompt, 1500),
	)
	cat, err := ai.Chat([]ChatMessage{systemMsg, {Role: "user", Content: userPrompt}})
	if err != nil {
		return "", err
	}
	cat = strings.TrimSpace(cat)
	candidates := []string{"沙特签证", "中东商旅", "出行指南", "政策解读"}
	for _, c := range candidates {
		if strings.Contains(cat, c) {
			return c, nil
		}
	}
	// 兜底
	return "沙特签证", nil
}

// ============================================================
//  ExtractContentFromURL / ExtractContentFromFile
// ============================================================

// ExtractContentFromURL 对 URL 做内容摘要（真实调用 LLM）
func (ai *AIService) ExtractContentFromURL(url string) (string, error) {
	systemMsg := ChatMessage{Role: "system", Content: ai.SystemPrompt}
	userPrompt := fmt.Sprintf(
		"用户提供了一个链接：%s\n\n请基于这个 URL 给出一个 200-400 字的内容摘要与要点。",
		url,
	)
	return ai.Chat([]ChatMessage{systemMsg, {Role: "user", Content: userPrompt}})
}

// ExtractContentFromFile 从文件内容生成结构化摘要
func (ai *AIService) ExtractContentFromFile(fileData []byte, fileName string) (string, error) {
	content := string(fileData)
	if content == "" {
		return fmt.Sprintf("(空文件：%s)", fileName), nil
	}
	return ai.GenerateTextContent(
		fmt.Sprintf("基于文件 '%s' 的内容：\n\n%s", fileName, truncate(content, 3000)),
		models.Article,
	)
}

// ============================================================
//  GenerateStructuredContent — 生成结构化 Content
// ============================================================

func (ai *AIService) GenerateStructuredContent(prompt string, contentType models.ContentType) (*models.Content, error) {
	body, err := ai.GenerateTextContent(prompt, contentType)
	if err != nil {
		return nil, err
	}
	title, _ := ai.GenerateTitleFromPrompt(prompt + "\n" + truncate(body, 500))
	summary, _ := ai.GenerateSummary(prompt, body)
	metaTags, _ := ai.GenerateMetaTags(prompt + "\n" + truncate(body, 500))
	category, _ := ai.DetectCategory(prompt + "\n" + truncate(body, 500))

	blocks := buildBlocks(title, summary, body, metaTags, category)

	dataJSON, _ := json.Marshal(map[string]interface{}{
		"title":    title,
		"summary":  summary,
		"category": category,
		"tags":     metaTags,
		"body":     body,
	})

	return &models.Content{
		Title:            title,
		Summary:          summary,
		Type:             contentType,
		InputType:        models.PromptInput,
		ContentData:      string(dataJSON),
		GeneratedContent: body,
		Tags:             metaTags,
		Status:           "draft",
		Category:         category,
		MetaTags:         metaTags,
		Blocks:           blocks,
	}, nil
}

// ============================================================
//  分发格式化 ProcessContentForDistribution
// ============================================================

func (ai *AIService) ProcessContentForDistribution(content *models.Content, platform string) (string, error) {
	systemMsg := ChatMessage{Role: "system", Content: ai.SystemPrompt}
	var userPrompt string
	switch platform {
	case "wechat":
		userPrompt = fmt.Sprintf(
			"请将以下内容改写为微信公众号风格，标题+正文，段落+表情，不超过 2000 字：\n\n标题：%s\n\n内容：%s",
			content.Title, truncate(content.GeneratedContent, 3000),
		)
	case "douyin", "xiaohongshu", "video_channel":
		userPrompt = fmt.Sprintf(
			"请将以下内容改写为短文案，适合短视频/小红书风格，带话题标签：\n\n%s",
			truncate(content.GeneratedContent, 3000),
		)
	default:
		// CMS 默认直接原样输出 JSON
		return fmt.Sprintf(
			"{\"title\":\"%s\",\"content\":\"%s\",\"description\":\"%s\",\"tags\":\"%s\"}",
			content.Title,
			escapeJSON(content.GeneratedContent),
			escapeJSON(content.Summary),
			escapeJSON(content.MetaTags),
		), nil
	}
	return ai.Chat([]ChatMessage{systemMsg, {Role: "user", Content: userPrompt}})
}

// ============================================================
//  工具函数
// ============================================================

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

func escapeJSON(s string) string {
	b, err := json.Marshal(s)
	if err != nil {
		return ""
	}
	out := string(b)
	if len(out) >= 2 {
		return out[1 : len(out)-1]
	}
	return out
}

// buildBlocks 将 title/summary/body 拆分为结构化内容块
func buildBlocks(title, summary, body, metaTags, category string) []models.ContentBlock {
	order := 1
	blocks := []models.ContentBlock{}
	blocks = append(blocks, models.ContentBlock{
		BlockType: models.TitleBlock,
		Content:   title,
		Order:     order,
	})
	order++
	if strings.TrimSpace(summary) != "" {
		blocks = append(blocks, models.ContentBlock{
			BlockType: models.ParagraphBlock,
			Content:   summary,
			Order:     order,
		})
		order++
	}
	for _, para := range splitParagraphs(body) {
		blocks = append(blocks, models.ContentBlock{
			BlockType: models.ParagraphBlock,
			Content:   para,
			Order:     order,
		})
		order++
	}
	if strings.TrimSpace(metaTags) != "" {
		blocks = append(blocks, models.ContentBlock{
			BlockType: models.ListBlock,
			Content:   "标签：" + metaTags + "；分类：" + category,
			Order:     order,
		})
	}
	return blocks
}

// splitParagraphs 把文本按空行切分
func splitParagraphs(text string) []string {
	raw := strings.ReplaceAll(text, "\r\n", "\n")
	parts := strings.Split(raw, "\n\n")
	var out []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// ============================================================
//  兼容保留 CallAIEndpoint (向后兼容)
// ============================================================

// CallAIEndpoint 通用 HTTP 调用端点的低级请求，保留向后兼容
func (ai *AIService) CallAIEndpoint(endpoint string, requestData map[string]interface{}) ([]byte, error) {
	reqBytes, err := json.Marshal(requestData)
	if err != nil {
		return nil, err
	}
	fullURL := ai.BaseURL + endpoint
	httpReq, err := http.NewRequest("POST", fullURL, bytes.NewReader(reqBytes))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+ai.APIKey)
	httpReq.Header.Set("api-key", ai.APIKey)
	resp, err := ai.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("LLM 返回错误 (HTTP %d): %s", resp.StatusCode, string(b))
	}
	return io.ReadAll(resp.Body)
}

// 避免未使用警告
var _ = minInt

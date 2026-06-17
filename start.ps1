# ============================================================
#  AI Content Platform - PowerShell 启动脚本
#  配置 LLM 提供商、API Key 等后执行: ./start.ps1
# ============================================================

# ============ 配置项（请修改）============

# LLM 提供商标识 (仅用于 UI 显示)
$env:LLM_PROVIDER = "deepseek"

# 你的 API Key（必填）
$env:LLM_API_KEY = "sk-your-api-key-here"

# API Base URL（OpenAI 兼容格式）
# 常见示例：
#   OpenAI:      https://api.openai.com/v1
#   DeepSeek:    https://api.deepseek.com/v1
#   智谱:        https://open.bigmodel.cn/api/paas/v4
#   阿里通义:    https://dashscope.aliyuncs.com/compatible-mode/v1
#   SiliconFlow: https://api.siliconflow.cn/v1
#   Ollama本地:  http://localhost:11434/v1
$env:LLM_BASE_URL = "https://api.deepseek.com/v1"

# 模型名称
#   DeepSeek: deepseek-chat / deepseek-reasoner
#   智谱:    glm-4 / glm-4-flash
#   OpenAI:  gpt-4o-mini / gpt-4o / gpt-3.5-turbo
#   通义千问: qwen-turbo / qwen-plus
#   Ollama:  qwen2.5:7b / deepseek-r1:7b 等
$env:LLM_MODEL = "deepseek-chat"

# （可选）系统提示词 - 覆盖默认系统提示
# $env:LLM_SYSTEM_PROMPT = "你是一个专业的中文写作助手，擅长撰写沙特签证和中东商旅类内容。"

# ============ 启动服务 ============

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location $scriptDir

Write-Host ""
Write-Host "============================================" -ForegroundColor Cyan
Write-Host "  AI Content Platform 启动中..." -ForegroundColor Cyan
Write-Host "  Provider : $env:LLM_PROVIDER" -ForegroundColor Yellow
Write-Host "  Base URL : $env:LLM_BASE_URL" -ForegroundColor Yellow
Write-Host "  Model    : $env:LLM_MODEL" -ForegroundColor Yellow
Write-Host "============================================" -ForegroundColor Cyan
Write-Host ""

# 构建
Write-Host "构建项目..." -ForegroundColor Green
go build -o ai-content-platform.exe .
if ($LASTEXITCODE -ne 0) {
    Write-Host "[错误] 构建失败，请先安装 Go 1.21+" -ForegroundColor Red
    Read-Host "按 Enter 退出"
    exit 1
}

Write-Host ""
Write-Host "服务地址: http://localhost:8080" -ForegroundColor Green
Write-Host "管理后台: http://localhost:8080/admin" -ForegroundColor Green
Write-Host "登录账号: admin / admin123" -ForegroundColor Green
Write-Host "按 Ctrl+C 停止服务" -ForegroundColor Gray
Write-Host ""

./ai-content-platform.exe

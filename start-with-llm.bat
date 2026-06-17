@echo off
REM ============================================================
REM  AI Content Platform - 启动脚本 (Windows)
REM  配置 LLM 提供商、API Key、Base URL 和模型名称后，双击即可启动
REM ============================================================

REM ============ 请修改以下配置 ============

REM LLM 提供商标识 (仅用于 UI 显示): openai / deepseek / zhipu / anthropic / ollama 等
set LLM_PROVIDER=deepseek

REM 你的 API Key（必填）- 到对应提供商官网申请
set LLM_API_KEY=sk-your-api-key-here

REM API Base URL（OpenAI 兼容格式，需包含 /v1）
REM 常见示例：
REM   OpenAI:      https://api.openai.com/v1
REM   DeepSeek:    https://api.deepseek.com/v1
REM   智谱:        https://open.bigmodel.cn/api/paas/v4
REM   阿里通义:    https://dashscope.aliyuncs.com/compatible-mode/v1
REM   Ollama本地:  http://localhost:11434/v1
set LLM_BASE_URL=https://api.deepseek.com/v1

REM 模型名称
REM   DeepSeek: deepseek-chat / deepseek-reasoner
REM   智谱:    glm-4 / glm-4-flash
REM   OpenAI:  gpt-4o-mini / gpt-4o / gpt-3.5-turbo
REM   通义千问: qwen-turbo / qwen-plus
REM   Ollama:  qwen2.5:7b / deepseek-r1:7b 等
set LLM_MODEL=deepseek-chat

REM （可选）系统提示词 - 覆盖默认系统提示
REM set LLM_SYSTEM_PROMPT=你是一个专业的中文写作助手，擅长撰写沙特签证和中东商旅类内容。

REM ================ 启动服务 ================

echo.
echo ============================================
echo   AI Content Platform 启动中...
echo   Provider : %LLM_PROVIDER%
echo   Base URL : %LLM_BASE_URL%
echo   Model    : %LLM_MODEL%
echo ============================================
echo.

cd /d "%~dp0"
go build -o ai-content-platform.exe .
if errorlevel 1 (
    echo.
    echo [错误] 构建失败，请先安装 Go 1.21+
    pause
    exit /b 1
)

echo.
echo 启动服务: http://localhost:8080
echo 管理后台: http://localhost:8080/admin
echo 登录账号: admin / admin123
echo 按 Ctrl+C 停止服务
echo.

ai-content-platform.exe

pause

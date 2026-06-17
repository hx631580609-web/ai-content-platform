# AI内容生成分发系统与中盛启瀚官网

## 项目概述

本项目包含两个核心系统的建设：

### 1. AI内容生成分发系统
一个基于AI技术的内容创作与分发平台，支持文章和海报的智能化生成、编辑和多渠道分发。系统通过AI对话式交互，帮助用户快速生成高质量的文章内容和视觉海报，并支持一键发布到CMS官网、公众号、视频号、抖音、小红书等主流平台。

### 2. 中盛启瀚官网
一个面向沙特商务签证及中东商旅服务的官方网站，提供签证办理、保险、定制游、企业考察等业务的在线展示与咨询入口。网站采用响应式设计，支持中英双语，配备完善的内容管理后台，便于运营人员自主更新内容。

## 技术架构

- **后端框架**: Go + Gin
- **数据库**: PostgreSQL/MySQL (支持SQLite开发模式)
- **ORM**: GORM
- **认证**: JWT + Bcrypt
- **前端**: (待实现) Next.js 14+, Tailwind CSS, shadcn/ui

## 功能模块

### AI内容生成分发系统
- 用户认证与权限管理
- AI驱动的内容生成（文章、海报、视频）
- 内容编辑器（结构化内容块）
- 多平台内容分发
- 内容池管理

### 中盛启瀚官网
- 响应式前台页面
- 多语言支持（中英双语）
- 内容管理系统
- 博客文章管理
- 服务页面管理
- 关于我们页面
- 联系信息管理

## 安全特性

- JWT身份验证
- 输入验证与清理
- SQL注入防护
- XSS防护
- 速率限制
- 安全日志记录

## API文档

完整的API文档请参见 [docs/api_documentation.md](./docs/api_documentation.md)

## 项目结构

```
ai-content-platform/
├── api/                 # API相关代码
├── config/             # 配置文件
├── database/           # 数据库连接和迁移
├── docs/               # 文档
├── handlers/           # HTTP请求处理器
├── middleware/         # 中间件
├── models/             # 数据模型
├── routes/             # 路由定义
├── services/           # 业务逻辑服务
├── utils/              # 工具函数
├── main.go             # 主程序入口
├── go.mod
├── go.sum
└── README.md
```

## 快速开始

1. 克隆项目
```bash
git clone <repository-url>
cd ai-content-platform
```

2. 安装依赖
```bash
go mod tidy
```

3. 设置环境变量
```bash
export SERVER_PORT=8080
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=your_db_user
export DB_PASSWORD=your_db_password
export DB_NAME=ai_content_platform
export JWT_SECRET=your-jwt-secret
```

4. 运行应用
```bash
go run main.go
```

## 环境要求

- Go 1.18+
- PostgreSQL 或 MySQL 数据库
- Node.js (前端部分)

## 部署

应用启动时会自动创建数据库表结构和默认管理员账户(admin/admin123)。

## API端点

- `POST /register` - 用户注册
- `POST /login` - 用户登录
- `GET /profile` - 获取用户资料
- `POST /contents` - 创建内容
- `GET /contents` - 获取内容列表
- 管理员端点需要管理员权限

## 默认账户

- 用户名: `admin`
- 密码: `admin123`

## 许可证

MIT License
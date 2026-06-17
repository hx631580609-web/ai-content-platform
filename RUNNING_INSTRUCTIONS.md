# 运行说明

## 当前状态
项目已在8080端口上运行，但由于没有可用的数据库服务，系统以无数据库模式运行。API端点已加载，但数据不会被持久化。

## 数据库配置
要启用完整的数据库功能，您需要：

### 1. 安装并启动数据库服务
#### MySQL
- 安装MySQL 5.7或更高版本
- 启动MySQL服务
- 创建数据库: `CREATE DATABASE ai_content_platform;`
- 创建用户并授权: `CREATE USER 'app_user'@'%' IDENTIFIED BY 'password'; GRANT ALL PRIVILEGES ON ai_content_platform.* TO 'app_user'@'%';`

#### PostgreSQL
- 安装PostgreSQL 9.6或更高版本
- 启动PostgreSQL服务
- 创建数据库: `CREATE DATABASE ai_content_platform;`
- 创建用户: `CREATE USER app_user WITH PASSWORD 'password';`
- 授权: `GRANT ALL PRIVILEGES ON DATABASE ai_content_platform TO app_user;`

### 2. 设置环境变量
```bash
# MySQL
export DB_HOST=localhost
export DB_PORT=3306
export DB_USER=app_user
export DB_PASSWORD=password
export DB_NAME=ai_content_platform

# 或者 PostgreSQL
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=app_user
export DB_PASSWORD=password
export DB_NAME=ai_content_platform
```

### 3. 重启应用
设置好数据库服务和环境变量后，重启应用即可使用完整的数据库功能。

## API端点
即使在无数据库模式下，以下API端点也可用（但数据不会持久化）：

### 公共端点
- `POST /register` - 用户注册
- `POST /login` - 用户登录
- `GET /website/modules` - 获取网站模块
- `GET /blog-posts` - 获取博客文章
- `GET /service-pages` - 获取服务页面

### 认证端点
- `GET /profile` - 获取用户资料
- `PUT /profile` - 更新用户资料
- `POST /contents` - 创建内容
- `GET /contents` - 获取内容列表

### 管理员端点
- `GET /users` - 获取所有用户
- `PUT /users/:id/role` - 更新用户角色

## 默认凭据
- 用户名: `admin`
- 密码: `admin123`

## 生产部署
在生产环境中，请确保：
1. 使用安全的数据库凭证
2. 设置强密钥用于JWT: `JWT_SECRET=your-very-secure-secret-key`
3. 使用HTTPS
4. 配置适当的防火墙规则
5. 设置数据库连接池参数
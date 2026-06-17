# AI内容生成分发系统 API 文档

## 概述

本文档提供了AI内容生成分发系统和中盛启瀚官网的所有API端点的详细信息。

## 基础信息

- **Base URL**: `http://localhost:8080`
- **认证**: 使用JWT Bearer Token进行身份验证
- **内容类型**: `application/json`

## 认证

### 注册用户
```
POST /register
```

**请求体**:
```json
{
  "username": "string (required)",
  "email": "string (required, valid email)",
  "password": "string (required, min 6 chars)"
}
```

**响应**:
```json
{
  "message": "User registered successfully",
  "user": {
    "id": 1,
    "username": "string",
    "email": "string",
    "role": "employee|admin"
  },
  "token": "jwt_token_string"
}
```

### 用户登录
```
POST /login
```

**请求体**:
```json
{
  "username": "string (required)",
  "password": "string (required)"
}
```

**响应**:
```json
{
  "message": "Login successful",
  "user": {
    "id": 1,
    "username": "string",
    "email": "string",
    "role": "employee|admin"
  },
  "token": "jwt_token_string"
}
```

## 用户管理

### 获取用户资料
```
GET /profile
Authorization: Bearer {token}
```

**响应**:
```json
{
  "user": {
    "id": 1,
    "username": "string",
    "email": "string",
    "role": "employee|admin",
    "created_at": "timestamp",
    "updated_at": "timestamp"
  }
}
```

### 更新用户资料
```
PUT /profile
Authorization: Bearer {token}
```

**请求体**:
```json
{
  "username": "string (optional)",
  "email": "string (optional, valid email)",
  "password": "string (optional, min 6 chars)"
}
```

## 内容管理

### 创建内容
```
POST /contents
Authorization: Bearer {token}
```

**请求体**:
```json
{
  "title": "string (required)",
  "summary": "string (optional)",
  "type": "article|poster|video (required)",
  "input_type": "prompt|link|file|paste (required)",
  "content_data": "string (optional)",
  "cover_image": "string (optional)",
  "tags": "string (optional)",
  "source_url": "string (optional)",
  "status": "draft|published|archived (optional, defaults to draft)",
  "blocks": [
    {
      "block_type": "title|paragraph|image|list|table",
      "content": "string",
      "media_url": "string (optional)",
      "order": "number (optional, defaults to 0)"
    }
  ]
}
```

### 获取内容列表
```
GET /contents?page=1&limit=10&title=search&status=draft&type=article&tags=tag1,tag2
Authorization: Bearer {token}
```

**响应**:
```json
{
  "contents": [
    {
      "id": 1,
      "title": "string",
      "summary": "string",
      "type": "article|poster|video",
      "input_type": "prompt|link|file|paste",
      "status": "draft|published|archived",
      "cover_image": "string",
      "tags": "string",
      "source_url": "string",
      "user": {
        "id": 1,
        "username": "string",
        "email": "string"
      },
      "blocks": [...],
      "distributors": [...]
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 10,
    "total": 100
  }
}
```

### 获取单个内容
```
GET /contents/{id}
Authorization: Bearer {token}
```

### 更新内容
```
PUT /contents/{id}
Authorization: Bearer {token}
```

### 删除内容
```
DELETE /contents/{id}
Authorization: Bearer {token}
```

### 发布内容
```
POST /contents/{id}/publish
Authorization: Bearer {token}
```

### 归档内容
```
POST /contents/{id}/archive
Authorization: Bearer {token}
```

## 网站内容管理

### 获取网站模块
```
GET /website/modules
```

### 获取特定网站模块
```
GET /website/modules/{name}
```

### 更新网站模块
```
PUT /website/modules/{name}
Authorization: Bearer {token}
```

**请求体**:
```json
{
  "enabled": true|false,
  "position": number,
  "config": "json_string"
}
```

### 获取博客文章列表
```
GET /blog-posts?page=1&limit=10&category=saudi_visa&status=published&search=keywords
```

### 获取单个博客文章
```
GET /blog-posts/{slug}
```

### 创建博客文章
```
POST /blog-posts
Authorization: Bearer {token}
```

**请求体**:
```json
{
  "title": "string (required)",
  "slug": "string (required)",
  "content": "string (required)",
  "summary": "string (optional)",
  "category": "string (optional)",
  "cover_image": "string (optional)",
  "status": "string (optional)"
}
```

### 更新博客文章
```
PUT /blog-posts/{slug}
Authorization: Bearer {token}
```

### 删除博客文章
```
DELETE /blog-posts/{slug}
Authorization: Bearer {token}
```

### 获取服务页面列表
```
GET /service-pages?page=1&limit=10&service_type=saudi_business_visa&status=active
```

### 获取单个服务页面
```
GET /service-pages/{slug}
```

### 创建服务页面
```
POST /service-pages
Authorization: Bearer {token}
```

**请求体**:
```json
{
  "title": "string (required)",
  "slug": "string (required)",
  "description": "string (optional)",
  "content": "string (required)",
  "service_type": "string (required)",
  "cover_image": "string (optional)",
  "status": "string (optional)",
  "contact_info": "json_string (optional)",
  "faq": "json_string (optional)"
}
```

### 更新服务页面
```
PUT /service-pages/{slug}
Authorization: Bearer {token}
```

### 删除服务页面
```
DELETE /service-pages/{slug}
Authorization: Bearer {token}
```

### 获取关于我们页面
```
GET /about-us
```

### 更新关于我们页面
```
PUT /about-us
Authorization: Bearer {token}
```

**请求体**:
```json
{
  "company_intro": "string (optional)",
  "qualifications": "json_string (optional)",
  "team_photos": "json_string (optional)",
  "contact_info": "json_string (optional)"
}
```

### 获取页脚联系信息
```
GET /footer-contact
```

### 更新页脚联系信息
```
PUT /footer-contact
Authorization: Bearer {token}
```

**请求体**:
```json
{
  "we_chat_qr": "string (optional)",
  "we_chat_id": "string (optional)",
  "phone": "string (optional)",
  "email": "string (optional)",
  "address": "string (optional)"
}
```

## 管理员专用端点

### 获取所有用户（仅限管理员）
```
GET /users
Authorization: Bearer {token}
```

### 获取特定用户（仅限管理员）
```
GET /users/{id}
Authorization: Bearer {token}
```

### 更新用户角色（仅限管理员）
```
PUT /users/{id}/role
Authorization: Bearer {token}
```

**请求体**:
```json
{
  "role": 0|1 (0=admin, 1=employee)
}
```

### 删除用户（仅限管理员）
```
DELETE /users/{id}
Authorization: Bearer {token}
```

### 获取系统日志（仅限管理员）
```
GET /system-logs?page=1&limit=10&user_id=1&action=login&object_type=user
Authorization: Bearer {token}
```

## 错误响应

所有错误响应都遵循以下格式：

```json
{
  "error": "错误消息"
}
```

## 状态码

- `200`: 成功
- `201`: 创建成功
- `400`: 请求错误
- `401`: 未授权
- `403`: 禁止访问
- `404`: 资源未找到
- `409`: 冲突（例如用户名已存在）
- `429`: 请求过多
- `500`: 服务器内部错误
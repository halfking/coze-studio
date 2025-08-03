# CozeRights API 参考文档

## 📋 API 概述

CozeRights 提供完整的 RESTful API，支持权限管理、审计日志、高级功能等企业级特性。

## 🔐 认证方式

### JWT Token认证
```bash
# 登录获取Token
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "admin123"
  }'

# 响应
{
  "code": 200,
  "message": "Login successful",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": 1,
      "username": "admin",
      "email": "admin@example.com"
    }
  }
}
```

### API调用示例
```bash
# 使用Token访问API
curl -X GET http://localhost:8080/api/v1/users \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

## 👥 用户管理 API

### 获取用户列表
```http
GET /api/v1/users
```

**参数:**
- `page` (int): 页码，默认1
- `page_size` (int): 每页数量，默认10
- `search` (string): 搜索关键词

**响应:**
```json
{
  "code": 200,
  "message": "Success",
  "data": {
    "total": 100,
    "items": [
      {
        "id": 1,
        "username": "john_doe",
        "email": "john@example.com",
        "system_role": "user",
        "status": "active",
        "created_at": "2025-08-03T10:00:00Z"
      }
    ]
  }
}
```

### 创建用户
```http
POST /api/v1/users
```

**请求体:**
```json
{
  "username": "new_user",
  "email": "user@example.com",
  "password": "secure_password",
  "system_role": "user"
}
```

## 🏢 工作空间管理 API

### 获取工作空间列表
```http
GET /api/v1/workspaces
```

### 创建工作空间
```http
POST /api/v1/workspaces
```

**请求体:**
```json
{
  "name": "开发团队",
  "description": "产品开发工作空间",
  "settings": {
    "max_members": 50,
    "features": ["agent", "workflow", "knowledge"]
  }
}
```

## 🔑 权限管理 API

### 检查权限
```http
POST /api/v1/permissions/check
```

**请求体:**
```json
{
  "resource": "agent",
  "action": "create",
  "workspace_id": 1
}
```

**响应:**
```json
{
  "code": 200,
  "data": {
    "has_permission": true,
    "reason": "User has workspace admin role"
  }
}
```

### 批量权限检查
```http
POST /api/v1/permissions/batch-check
```

**请求体:**
```json
{
  "checks": [
    {
      "resource": "agent",
      "action": "create",
      "workspace_id": 1
    },
    {
      "resource": "workflow",
      "action": "execute",
      "workspace_id": 1
    }
  ]
}
```

## 📊 审计日志 API

### 获取审计日志
```http
GET /api/v1/audit/logs
```

**参数:**
- `start_time` (string): 开始时间 (ISO 8601)
- `end_time` (string): 结束时间 (ISO 8601)
- `user_id` (int): 用户ID
- `action` (string): 操作类型
- `resource` (string): 资源类型
- `status` (string): 状态 (success/failed)

**响应:**
```json
{
  "code": 200,
  "data": {
    "total": 1000,
    "items": [
      {
        "id": 1,
        "user_id": 1,
        "action": "create",
        "resource": "agent",
        "status": "success",
        "ip_address": "192.168.1.100",
        "created_at": "2025-08-03T10:00:00Z"
      }
    ]
  }
}
```

### 导出审计日志
```http
GET /api/v1/audit/logs/export?format=csv
```

## 🔧 Coze集成 API

### Coze权限检查
```http
POST /api/v1/coze/permissions/check
```

**请求体:**
```json
{
  "resource": "agent",
  "action": "create",
  "workspace_id": 1,
  "resource_id": "agent_123"
}
```

### 记录使用量
```http
POST /api/v1/coze/usage/record
```

**请求体:**
```json
{
  "resource": "model",
  "action": "api_call",
  "workspace_id": 1,
  "quantity": 1000,
  "unit": "tokens",
  "metadata": {
    "model_name": "gpt-4",
    "endpoint": "/v1/chat/completions"
  }
}
```

### 获取用户配额
```http
GET /api/v1/coze/quota?workspace_id=1
```

**响应:**
```json
{
  "code": 200,
  "data": {
    "workspace_id": 1,
    "quotas": {
      "api_calls": {
        "limit": 10000,
        "used": 2500,
        "remaining": 7500
      },
      "storage": {
        "limit": "10GB",
        "used": "2.5GB",
        "remaining": "7.5GB"
      }
    }
  }
}
```

## 🚀 高级功能 API

### 策略评估
```http
POST /api/v1/advanced/policy/evaluate
```

**请求体:**
```json
{
  "user_id": 1,
  "resource": "agent",
  "action": "create",
  "context": {
    "time": "2025-08-03T10:00:00Z",
    "ip": "192.168.1.100",
    "workspace_id": 1
  }
}
```

### 计费使用量记录
```http
POST /api/v1/advanced/billing/usage
```

### 生成账单
```http
POST /api/v1/advanced/billing/invoice
```

### 获取监控指标
```http
GET /api/v1/advanced/monitoring/metrics
```

**响应:**
```json
{
  "code": 200,
  "data": {
    "api_requests": 10000,
    "active_users": 150,
    "error_rate": 0.01,
    "avg_response_time": 120,
    "system_health": "healthy"
  }
}
```

## 📈 错误处理

### 错误响应格式
```json
{
  "code": 400,
  "message": "Validation failed",
  "error": "Invalid input parameters",
  "details": {
    "field": "email",
    "reason": "Invalid email format"
  },
  "trace_id": "abc123def456"
}
```

### 常见错误码
- `200`: 成功
- `400`: 请求参数错误
- `401`: 未认证
- `403`: 权限不足
- `404`: 资源不存在
- `429`: 请求频率限制
- `500`: 服务器内部错误

## 🔄 API版本控制

当前API版本：`v1`

所有API端点都包含版本前缀：`/api/v1/`

## 📞 技术支持

- **文档**: https://github.com/coze-dev/coze-studio/wiki
- **Issues**: https://github.com/coze-dev/coze-studio/issues
- **讨论**: https://github.com/coze-dev/coze-studio/discussions

---

**完整的API让您轻松集成CozeRights到任何系统！** 🚀

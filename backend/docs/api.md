# CozeRights API 文档

## 概述

CozeRights 是一个多租户权限管理系统，提供完整的 RBAC（基于角色的访问控制）功能。

## 认证

所有API请求都需要在请求头中包含认证信息：

```
Authorization: Bearer <token>
```

## 响应格式

所有API响应都遵循统一的格式：

```json
{
  "code": 200,
  "message": "Success",
  "data": {},
  "error": ""
}
```

分页响应格式：

```json
{
  "code": 200,
  "message": "Success",
  "data": {
    "items": [],
    "pagination": {
      "page": 1,
      "page_size": 20,
      "total": 100,
      "total_pages": 5
    }
  }
}
```

## 租户管理 API

### 创建租户

**POST** `/api/v1/tenants`

**权限要求**: 超级管理员

**请求体**:
```json
{
  "name": "租户名称",
  "code": "tenant_code",
  "description": "租户描述",
  "max_users": 100,
  "max_spaces": 10
}
```

**响应**:
```json
{
  "code": 200,
  "message": "Tenant created successfully",
  "data": {
    "id": 1,
    "name": "租户名称",
    "code": "tenant_code",
    "description": "租户描述",
    "is_active": true,
    "max_users": 100,
    "max_spaces": 10,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

### 获取租户列表

**GET** `/api/v1/tenants`

**权限要求**: 超级管理员

**查询参数**:
- `page`: 页码（默认: 1）
- `page_size`: 每页大小（默认: 20，最大: 100）
- `name`: 租户名称搜索
- `code`: 租户代码搜索
- `is_active`: 是否激活（true/false）

### 获取单个租户

**GET** `/api/v1/tenants/{id}`

**权限要求**: 超级管理员或租户管理员（只能查看自己的租户）

### 更新租户

**PUT** `/api/v1/tenants/{id}`

**权限要求**: 超级管理员

**请求体**:
```json
{
  "name": "新租户名称",
  "description": "新描述",
  "max_users": 200,
  "max_spaces": 20,
  "is_active": true
}
```

### 删除租户

**DELETE** `/api/v1/tenants/{id}`

**权限要求**: 超级管理员

## 用户管理 API

### 创建用户

**POST** `/api/v1/users`

**权限要求**: 租户管理员或超级管理员

**请求体**:
```json
{
  "username": "用户名",
  "email": "user@example.com",
  "password": "password123",
  "first_name": "名",
  "last_name": "姓",
  "phone": "手机号",
  "system_role": "user",
  "department_id": 1
}
```

### 获取用户列表

**GET** `/api/v1/users`

**权限要求**: 租户内用户

**查询参数**:
- `page`: 页码
- `page_size`: 每页大小
- `username`: 用户名搜索
- `email`: 邮箱搜索
- `system_role`: 系统角色过滤
- `department_id`: 部门ID过滤
- `is_active`: 是否激活

### 获取单个用户

**GET** `/api/v1/users/{id}`

**权限要求**: 用户本人或管理员

### 更新用户

**PUT** `/api/v1/users/{id}`

**权限要求**: 用户本人或管理员

**请求体**:
```json
{
  "first_name": "新名字",
  "last_name": "新姓氏",
  "phone": "新手机号",
  "avatar": "头像URL",
  "is_active": true,
  "department_id": 2
}
```

## 角色管理 API

### 创建角色

**POST** `/api/v1/roles`

**权限要求**: 租户管理员或超级管理员

**请求体**:
```json
{
  "name": "角色名称",
  "code": "role_code",
  "description": "角色描述",
  "parent_id": 1
}
```

### 获取角色列表

**GET** `/api/v1/roles`

**权限要求**: 租户内用户

**查询参数**:
- `page`: 页码
- `page_size`: 每页大小
- `name`: 角色名称搜索
- `code`: 角色代码搜索
- `is_system`: 是否系统角色

### 获取单个角色

**GET** `/api/v1/roles/{id}`

**权限要求**: 租户内用户

### 更新角色

**PUT** `/api/v1/roles/{id}`

**权限要求**: 租户管理员或超级管理员

### 删除角色

**DELETE** `/api/v1/roles/{id}`

**权限要求**: 租户管理员或超级管理员

## 权限管理 API

### 检查权限

**POST** `/api/v1/permissions/check`

**权限要求**: 已认证用户

**请求体**:
```json
{
  "resource": "user",
  "action": "read"
}
```

**响应**:
```json
{
  "code": 200,
  "message": "Permission check completed",
  "data": {
    "allowed": true,
    "resource": "user",
    "action": "read"
  }
}
```

### 批量检查权限

**POST** `/api/v1/permissions/batch-check`

**权限要求**: 已认证用户

**请求体**:
```json
{
  "checks": [
    {
      "resource": "user",
      "action": "read"
    },
    {
      "resource": "user",
      "action": "create"
    }
  ]
}
```

### 分配角色

**POST** `/api/v1/permissions/assign-role`

**权限要求**: 租户管理员或超级管理员

**请求体**:
```json
{
  "user_id": 1,
  "role_id": 2
}
```

### 撤销角色

**POST** `/api/v1/permissions/revoke-role`

**权限要求**: 租户管理员或超级管理员

**请求体**:
```json
{
  "user_id": 1,
  "role_id": 2
}
```

### 授予权限

**POST** `/api/v1/permissions/grant`

**权限要求**: 租户管理员或超级管理员

**请求体**:
```json
{
  "role_id": 1,
  "permission_codes": ["user:read", "user:create"]
}
```

### 获取用户权限

**GET** `/api/v1/permissions/users/{id}`

**权限要求**: 用户本人或管理员

**响应**:
```json
{
  "code": 200,
  "message": "User permissions retrieved successfully",
  "data": {
    "user_id": 1,
    "permissions": ["user:read", "user:create"],
    "roles": [
      {
        "id": 1,
        "name": "普通用户",
        "code": "user"
      }
    ]
  }
}
```

## 错误码

| 错误码 | 说明 |
|--------|------|
| 400 | 请求参数错误 |
| 401 | 未认证 |
| 403 | 权限不足 |
| 404 | 资源不存在 |
| 409 | 资源冲突 |
| 500 | 服务器内部错误 |

## 权限代码

### 系统权限
- `tenant:read` - 查看租户信息
- `tenant:create` - 创建租户
- `tenant:update` - 更新租户
- `tenant:delete` - 删除租户

### 用户权限
- `user:read` - 查看用户
- `user:create` - 创建用户
- `user:update` - 更新用户
- `user:delete` - 删除用户

### 角色权限
- `role:read` - 查看角色
- `role:create` - 创建角色
- `role:update` - 更新角色
- `role:delete` - 删除角色

### 权限管理
- `permission:assign` - 分配权限
- `permission:revoke` - 撤销权限

### 工作空间权限
- `workspace:read` - 查看工作空间
- `workspace:create` - 创建工作空间
- `workspace:update` - 更新工作空间
- `workspace:delete` - 删除工作空间
- `workspace:member:read` - 查看成员
- `workspace:member:create` - 添加成员
- `workspace:member:update` - 更新成员
- `workspace:member:delete` - 删除成员

### 审计权限
- `audit:read` - 查看审计日志

## 工作空间管理 API

### 创建工作空间

**POST** `/api/v1/workspaces`

**权限要求**: 租户管理员或有创建权限的用户

**请求体**:
```json
{
  "name": "工作空间名称",
  "code": "workspace_code",
  "description": "工作空间描述",
  "type": "team",
  "max_members": 50,
  "max_agents": 100,
  "max_workflows": 200
}
```

**响应示例**:
```json
{
  "success": true,
  "message": "Workspace created successfully",
  "data": {
    "id": 1,
    "tenant_id": 1,
    "name": "工作空间名称",
    "code": "workspace_code",
    "description": "工作空间描述",
    "type": "team",
    "is_active": true,
    "max_members": 50,
    "max_agents": 100,
    "max_workflows": 200,
    "created_by": 1,
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

### 获取工作空间列表

**GET** `/api/v1/workspaces`

**权限要求**: 已认证用户

**查询参数**:
- `page`: 页码 (默认: 1)
- `page_size`: 每页数量 (默认: 20)
- `name`: 工作空间名称过滤
- `type`: 工作空间类型过滤 (personal/team/enterprise)
- `is_active`: 是否激活过滤

### 获取单个工作空间

**GET** `/api/v1/workspaces/{id}`

**权限要求**: 工作空间成员

### 更新工作空间

**PUT** `/api/v1/workspaces/{id}`

**权限要求**: 工作空间管理员或所有者

**请求体**:
```json
{
  "name": "新的工作空间名称",
  "description": "新的描述",
  "max_members": 100,
  "max_agents": 200,
  "max_workflows": 500,
  "is_active": true
}
```

### 删除工作空间

**DELETE** `/api/v1/workspaces/{id}`

**权限要求**: 工作空间所有者

## 工作空间成员管理 API

### 获取工作空间成员列表

**GET** `/api/v1/workspaces/{id}/members`

**权限要求**: 工作空间成员

**查询参数**:
- `page`: 页码 (默认: 1)
- `page_size`: 每页数量 (默认: 20)
- `role`: 角色过滤 (owner/admin/member/guest)

### 添加工作空间成员

**POST** `/api/v1/workspaces/{id}/members`

**权限要求**: 工作空间管理员或所有者

**请求体**:
```json
{
  "user_id": 123,
  "role": "member"
}
```

### 更新成员角色

**PUT** `/api/v1/workspaces/{id}/members/{user_id}`

**权限要求**: 工作空间管理员或所有者

**请求体**:
```json
{
  "role": "admin"
}
```

### 移除工作空间成员

**DELETE** `/api/v1/workspaces/{id}/members/{user_id}`

**权限要求**: 工作空间管理员或所有者

## 工作空间资源管理 API

### Agent管理

#### 创建Agent

**POST** `/api/v1/workspaces/{workspace_id}/agents`

**权限要求**: 工作空间成员（创建权限）

**请求体**:
```json
{
  "name": "我的聊天Agent",
  "description": "智能客服Agent",
  "type": "chat",
  "coze_agent_id": "coze_agent_123",
  "config": "{\"model\":\"gpt-4\",\"temperature\":0.7}",
  "is_public": false
}
```

#### 获取Agent列表

**GET** `/api/v1/workspaces/{workspace_id}/agents`

**权限要求**: 工作空间成员

**查询参数**:
- `page`: 页码 (默认: 1)
- `page_size`: 每页数量 (默认: 20)
- `name`: Agent名称过滤
- `type`: Agent类型过滤 (chat/workflow/api)
- `status`: 状态过滤 (active/inactive/archived)
- `is_public`: 是否公共过滤
- `creator`: 创建者ID过滤

#### 获取单个Agent

**GET** `/api/v1/workspaces/{workspace_id}/agents/{id}`

**权限要求**: 工作空间成员

#### 更新Agent

**PUT** `/api/v1/workspaces/{workspace_id}/agents/{id}`

**权限要求**: Agent创建者或工作空间管理员

**请求体**:
```json
{
  "name": "更新的Agent名称",
  "description": "更新的描述",
  "config": "{\"model\":\"gpt-4\",\"temperature\":0.8}",
  "status": "active",
  "is_public": true
}
```

#### 删除Agent

**DELETE** `/api/v1/workspaces/{workspace_id}/agents/{id}`

**权限要求**: Agent创建者或工作空间管理员

### Workflow管理

#### 创建Workflow

**POST** `/api/v1/workspaces/{workspace_id}/workflows`

**权限要求**: 工作空间成员（创建权限）

**请求体**:
```json
{
  "name": "数据处理工作流",
  "description": "自动化数据处理流程",
  "version": "1.0.0",
  "coze_workflow_id": "coze_workflow_456",
  "definition": "{\"nodes\":[],\"edges\":[]}",
  "is_public": false
}
```

#### 获取Workflow列表

**GET** `/api/v1/workspaces/{workspace_id}/workflows`

**权限要求**: 工作空间成员

**查询参数**:
- `page`: 页码 (默认: 1)
- `page_size`: 每页数量 (默认: 20)
- `name`: Workflow名称过滤
- `version`: 版本过滤
- `status`: 状态过滤 (draft/published/archived)
- `is_public`: 是否公共过滤
- `creator`: 创建者ID过滤

#### 获取单个Workflow

**GET** `/api/v1/workspaces/{workspace_id}/workflows/{id}`

**权限要求**: 工作空间成员

#### 更新Workflow

**PUT** `/api/v1/workspaces/{workspace_id}/workflows/{id}`

**权限要求**: Workflow创建者或工作空间管理员

**请求体**:
```json
{
  "name": "更新的工作流名称",
  "description": "更新的描述",
  "version": "1.1.0",
  "definition": "{\"nodes\":[],\"edges\":[]}",
  "status": "published",
  "is_public": true
}
```

#### 删除Workflow

**DELETE** `/api/v1/workspaces/{workspace_id}/workflows/{id}`

**权限要求**: Workflow创建者或工作空间管理员

#### 执行Workflow

**POST** `/api/v1/workspaces/{workspace_id}/workflows/{id}/execute`

**权限要求**: 工作空间成员（执行权限）

**请求体**:
```json
{
  "input_data": {
    "text": "要处理的文本",
    "parameters": {}
  },
  "variables": {
    "env": "production",
    "timeout": 30
  }
}
```

**响应示例**:
```json
{
  "code": 200,
  "message": "Workflow execution started successfully",
  "data": {
    "execution_id": "exec_123_1672531200",
    "workflow_id": 123,
    "status": "running",
    "started_at": "2023-01-01T00:00:00Z",
    "coze_execution_id": "coze_workflow_456_exec_1672531200"
  }
}
```

#### 发布Workflow

**POST** `/api/v1/workspaces/{workspace_id}/workflows/{id}/publish`

**权限要求**: Workflow创建者或工作空间管理员

### Plugin管理

#### 安装Plugin

**POST** `/api/v1/workspaces/{workspace_id}/plugins`

**权限要求**: 工作空间成员（安装权限）

**请求体**:
```json
{
  "name": "数据分析插件",
  "description": "强大的数据分析工具",
  "type": "third_party",
  "version": "2.1.0",
  "coze_plugin_id": "coze_plugin_789",
  "config": {
    "api_key": "your_api_key",
    "endpoint": "https://api.example.com"
  },
  "is_shared": true
}
```

#### 获取Plugin列表

**GET** `/api/v1/workspaces/{workspace_id}/plugins`

**权限要求**: 工作空间成员

**查询参数**:
- `page`: 页码 (默认: 1)
- `page_size`: 每页数量 (默认: 20)
- `name`: Plugin名称过滤
- `is_enabled`: 是否启用过滤
- `category`: 插件类型过滤 (builtin/custom/third_party)

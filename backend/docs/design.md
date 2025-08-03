# CozeRights 系统设计文档

## 1. 系统架构设计

### 1.1 整体架构
```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   前端应用      │    │   API网关       │    │   负载均衡      │
│   (React/Vue)   │◄──►│   (Nginx)       │◄──►│   (HAProxy)     │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                        CozeRights 后端服务                      │
├─────────────────┬─────────────────┬─────────────────┬───────────┤
│   认证服务      │   权限服务      │   用户服务      │  审计服务 │
│   (Auth)        │   (RBAC)        │   (User)        │  (Audit)  │
└─────────────────┴─────────────────┴─────────────────┴───────────┘
                                │
                ┌───────────────┼───────────────┐
                ▼               ▼               ▼
        ┌─────────────┐ ┌─────────────┐ ┌─────────────┐
        │ PostgreSQL  │ │    Redis    │ │   日志系统  │
        │   数据库    │ │    缓存     │ │ (ELK Stack) │
        └─────────────┘ └─────────────┘ └─────────────┘
```

### 1.2 模块设计
- **认证模块**：JWT token生成和验证
- **权限模块**：RBAC权限检查和管理
- **用户模块**：用户和角色管理
- **工作空间模块**：工作空间和成员管理
- **审计模块**：操作日志记录和查询
- **缓存模块**：Redis缓存管理

### 1.3 数据流设计
```
用户请求 → 认证中间件 → 权限中间件 → 业务处理 → 审计记录 → 响应返回
    ↓           ↓           ↓           ↓           ↓
  JWT验证   → 权限检查   → 数据操作   → 日志记录   → 缓存更新
```

## 2. 数据库设计

### 2.1 核心表结构

#### 租户表 (tenants)
```sql
CREATE TABLE tenants (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    code VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    is_active BOOLEAN DEFAULT true,
    max_users INTEGER DEFAULT 100,
    max_spaces INTEGER DEFAULT 10,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);
```

#### 用户表 (users)
```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    tenant_id INTEGER REFERENCES tenants(id),
    username VARCHAR(50) NOT NULL,
    email VARCHAR(100) NOT NULL,
    password VARCHAR(255) NOT NULL,
    first_name VARCHAR(50),
    last_name VARCHAR(50),
    avatar VARCHAR(255),
    phone VARCHAR(20),
    is_active BOOLEAN DEFAULT true,
    system_role VARCHAR(20) DEFAULT 'user',
    department_id INTEGER REFERENCES departments(id),
    last_login TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    UNIQUE(tenant_id, username),
    UNIQUE(tenant_id, email)
);
```

#### 工作空间表 (workspaces)
```sql
CREATE TABLE workspaces (
    id SERIAL PRIMARY KEY,
    tenant_id INTEGER REFERENCES tenants(id),
    name VARCHAR(100) NOT NULL,
    code VARCHAR(50) NOT NULL,
    description TEXT,
    type VARCHAR(20) DEFAULT 'team',
    is_active BOOLEAN DEFAULT true,
    max_members INTEGER DEFAULT 50,
    max_agents INTEGER DEFAULT 100,
    max_workflows INTEGER DEFAULT 200,
    created_by INTEGER REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    UNIQUE(tenant_id, code)
);
```

#### 工作空间成员表 (workspace_members)
```sql
CREATE TABLE workspace_members (
    id SERIAL PRIMARY KEY,
    workspace_id INTEGER REFERENCES workspaces(id),
    user_id INTEGER REFERENCES users(id),
    role VARCHAR(20) DEFAULT 'member',
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(workspace_id, user_id)
);
```

### 2.2 数据隔离策略

#### 租户级隔离
- 所有业务数据都包含 `tenant_id` 字段
- 查询时必须带上租户条件
- 通过中间件自动注入租户过滤条件

#### 工作空间级隔离
- 工作空间资源包含 `workspace_id` 字段
- 成员权限通过 `workspace_members` 表控制
- 资源访问需要验证工作空间成员身份

#### 数据访问模式
```go
// 租户级查询
db.Where("tenant_id = ?", tenantID).Find(&users)

// 工作空间级查询
db.Where("workspace_id = ? AND user_id IN (?)", workspaceID, memberIDs).Find(&agents)
```

## 3. 权限设计

### 3.1 权限模型
```
租户 (Tenant)
├── 用户 (Users)
├── 角色 (Roles)
├── 权限 (Permissions)
└── 工作空间 (Workspaces)
    ├── 成员 (Members)
    ├── 角色 (Workspace Roles)
    └── 资源 (Resources)
        ├── Agents
        ├── Workflows
        └── Plugins
```

### 3.2 权限层级
1. **系统级权限**：超级管理员权限
2. **租户级权限**：租户管理员权限
3. **工作空间级权限**：工作空间内的权限
4. **资源级权限**：具体资源的权限

### 3.3 权限检查流程
```
1. 验证用户身份 (JWT)
2. 获取用户租户信息
3. 检查系统级权限
4. 检查租户级权限
5. 检查工作空间权限
6. 检查资源级权限
7. 返回权限结果
```

## 4. 工作空间设计

### 4.1 工作空间类型
- **个人空间**：个人使用，只有创建者
- **团队空间**：小团队使用，成员数量限制
- **企业空间**：大型组织使用，功能完整

### 4.2 成员角色权限
```
Owner (所有者):
  - workspace:*
  - member:*
  - agent:*
  - workflow:*
  - plugin:*

Admin (管理员):
  - workspace:read,update
  - member:read,create,update,delete
  - agent:*
  - workflow:*
  - plugin:*

Member (成员):
  - workspace:read
  - member:read
  - agent:read,create,update
  - workflow:read,create,update
  - plugin:read,use

Guest (访客):
  - workspace:read
  - member:read
  - agent:read
  - workflow:read
  - plugin:read
```

### 4.3 资源隔离设计
```go
// Agent 表结构
type WorkspaceAgent struct {
    ID          uint   `gorm:"primaryKey"`
    WorkspaceID uint   `gorm:"not null;index"`
    TenantID    uint   `gorm:"not null;index"`
    Name        string `gorm:"not null"`
    Config      string `gorm:"type:text"`
    CreatedBy   uint   `gorm:"not null"`
    // ... 其他字段
}

// 查询时的隔离
func GetWorkspaceAgents(workspaceID, userID uint) ([]WorkspaceAgent, error) {
    // 1. 验证用户是否为工作空间成员
    if !IsWorkspaceMember(workspaceID, userID) {
        return nil, ErrAccessDenied
    }
    
    // 2. 查询工作空间内的 Agents
    var agents []WorkspaceAgent
    return agents, db.Where("workspace_id = ?", workspaceID).Find(&agents).Error
}
```

## 5. 缓存设计

### 5.1 缓存策略
- **用户权限缓存**：5分钟TTL
- **角色权限缓存**：30分钟TTL
- **工作空间成员缓存**：10分钟TTL
- **工作空间配置缓存**：1小时TTL

### 5.2 缓存键设计
```
user_permissions:{user_id}
role_permissions:{role_id}
workspace_members:{workspace_id}
workspace_config:{workspace_id}
tenant_config:{tenant_id}
```

### 5.3 缓存失效策略
- **用户权限变更**：立即失效用户相关缓存
- **角色权限变更**：失效角色和相关用户缓存
- **工作空间变更**：失效工作空间相关缓存
- **定时刷新**：每小时刷新长期缓存

## 6. API设计

### 6.1 RESTful API规范
```
GET    /api/v1/workspaces              # 获取工作空间列表
POST   /api/v1/workspaces              # 创建工作空间
GET    /api/v1/workspaces/{id}         # 获取工作空间详情
PUT    /api/v1/workspaces/{id}         # 更新工作空间
DELETE /api/v1/workspaces/{id}         # 删除工作空间

GET    /api/v1/workspaces/{id}/members # 获取成员列表
POST   /api/v1/workspaces/{id}/members # 添加成员
PUT    /api/v1/workspaces/{id}/members/{user_id} # 更新成员角色
DELETE /api/v1/workspaces/{id}/members/{user_id} # 移除成员
```

### 6.2 响应格式
```json
{
  "success": true,
  "message": "操作成功",
  "data": {
    "id": 1,
    "name": "工作空间名称",
    "type": "team",
    "members_count": 5
  },
  "meta": {
    "page": 1,
    "page_size": 20,
    "total": 100
  }
}
```

### 6.3 错误处理
```json
{
  "success": false,
  "message": "操作失败",
  "error": "具体错误信息",
  "code": "ERROR_CODE"
}
```

## 7. 安全设计

### 7.1 认证安全
- **JWT Token**：使用RS256算法签名
- **Token刷新**：支持refresh token机制
- **Token黑名单**：支持token撤销

### 7.2 授权安全
- **最小权限原则**：用户只获得必要的权限
- **权限检查**：每个API都进行权限验证
- **敏感操作**：重要操作需要二次验证

### 7.3 数据安全
- **SQL注入防护**：使用参数化查询
- **XSS防护**：输入输出过滤
- **CSRF防护**：使用CSRF token

## 8. 性能优化

### 8.1 数据库优化
- **索引优化**：为常用查询字段建立索引
- **查询优化**：避免N+1查询问题
- **连接池**：使用数据库连接池

### 8.2 缓存优化
- **多级缓存**：内存缓存 + Redis缓存
- **缓存预热**：系统启动时预加载热点数据
- **缓存穿透防护**：使用布隆过滤器

### 8.3 API优化
- **批量操作**：支持批量权限检查
- **分页查询**：大数据量分页返回
- **异步处理**：耗时操作异步执行

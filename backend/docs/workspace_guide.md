# 工作空间管理功能使用指南

## 概述

工作空间管理功能提供了完整的多租户工作空间管理能力，包括工作空间创建、成员管理、权限控制和数据隔离。

## 核心特性

### 1. 数据隔离
- **租户级隔离**：不同租户的工作空间完全隔离
- **工作空间级隔离**：同一租户内不同工作空间的数据隔离
- **成员权限隔离**：只有工作空间成员才能访问工作空间资源

### 2. 角色权限
- **Owner（所有者）**：拥有工作空间的完全控制权，包括删除工作空间
- **Admin（管理员）**：可以管理成员和资源，但不能删除工作空间
- **Member（成员）**：可以使用工作空间资源，有限的管理权限
- **Guest（访客）**：只读访问权限

### 3. 工作空间类型
- **Personal（个人）**：个人使用，默认最多1个成员
- **Team（团队）**：小团队使用，默认最多50个成员
- **Enterprise（企业）**：大型组织使用，默认最多500个成员

## API使用示例

### 创建工作空间

```bash
curl -X POST http://localhost:8080/api/v1/workspaces \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "name": "我的团队工作空间",
    "code": "myteamworkspace",
    "description": "团队协作工作空间",
    "type": "team",
    "max_members": 20,
    "max_agents": 50,
    "max_workflows": 100
  }'
```

### 获取工作空间列表

```bash
curl -X GET "http://localhost:8080/api/v1/workspaces?page=1&page_size=10&type=team" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### 添加工作空间成员

```bash
curl -X POST http://localhost:8080/api/v1/workspaces/1/members \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "user_id": 123,
    "role": "member"
  }'
```

### 更新成员角色

```bash
curl -X PUT http://localhost:8080/api/v1/workspaces/1/members/123 \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "role": "admin"
  }'
```

## 权限控制

### 工作空间操作权限

| 操作 | Owner | Admin | Member | Guest |
|------|-------|-------|--------|-------|
| 查看工作空间 | ✅ | ✅ | ✅ | ✅ |
| 更新工作空间 | ✅ | ✅ | ❌ | ❌ |
| 删除工作空间 | ✅ | ❌ | ❌ | ❌ |
| 添加成员 | ✅ | ✅ | ❌ | ❌ |
| 移除成员 | ✅ | ✅ | ❌ | ❌ |
| 更新成员角色 | ✅ | ✅ | ❌ | ❌ |
| 转让所有权 | ✅ | ❌ | ❌ | ❌ |

### 资源操作权限

| 操作 | Owner | Admin | Member | Guest |
|------|-------|-------|--------|-------|
| 创建Agent | ✅ | ✅ | ✅ | ❌ |
| 编辑Agent | ✅ | ✅ | ✅ | ❌ |
| 删除Agent | ✅ | ✅ | ❌ | ❌ |
| 查看Agent | ✅ | ✅ | ✅ | ✅ |
| 创建Workflow | ✅ | ✅ | ✅ | ❌ |
| 编辑Workflow | ✅ | ✅ | ✅ | ❌ |
| 删除Workflow | ✅ | ✅ | ❌ | ❌ |
| 查看Workflow | ✅ | ✅ | ✅ | ✅ |

## 数据模型

### 工作空间表结构

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

### 工作空间成员表结构

```sql
CREATE TABLE workspace_members (
    id SERIAL PRIMARY KEY,
    workspace_id INTEGER REFERENCES workspaces(id),
    user_id INTEGER REFERENCES users(id),
    role VARCHAR(20) DEFAULT 'member',
    permissions TEXT[], -- 额外权限
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(workspace_id, user_id)
);
```

## 最佳实践

### 1. 工作空间设计
- 根据团队规模选择合适的工作空间类型
- 设置合理的成员和资源限制
- 使用有意义的工作空间代码（英文字母和数字）

### 2. 成员管理
- 遵循最小权限原则，只给予必要的权限
- 定期审查成员权限，移除不活跃的成员
- 重要操作前进行权限确认

### 3. 安全考虑
- 所有工作空间操作都会记录审计日志
- 敏感操作需要相应的权限验证
- 定期备份工作空间数据

### 4. 性能优化
- 工作空间成员信息会被缓存10分钟
- 大量成员的工作空间建议分页查询
- 避免频繁的权限检查操作

## 故障排除

### 常见错误

1. **工作空间代码已存在**
   - 错误：`Workspace code already exists`
   - 解决：使用不同的工作空间代码

2. **成员数量超限**
   - 错误：`Member limit exceeded`
   - 解决：增加工作空间成员限制或移除不活跃成员

3. **权限不足**
   - 错误：`Access denied`
   - 解决：确认用户有相应的工作空间权限

4. **用户不存在**
   - 错误：`User not found`
   - 解决：确认用户ID正确且属于当前租户

### 调试技巧

1. 检查用户的工作空间成员身份
2. 验证工作空间是否属于正确的租户
3. 查看审计日志了解操作历史
4. 确认工作空间状态是否为激活状态

## 扩展功能

工作空间管理功能为后续的资源管理提供了基础：

- **Agent管理**：在工作空间内创建和管理AI Agent
- **Workflow管理**：设计和执行工作流
- **Plugin管理**：安装和配置插件
- **资源共享**：在工作空间内共享资源
- **协作功能**：团队协作和沟通功能

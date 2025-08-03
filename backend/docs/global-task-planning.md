# CozeRights 与 Coze-Studio 全局任务规划

## 🎯 规划概述

基于对Coze-Studio源码的深度分析和整合策略制定，本文档提供了详细的分步实施计划，确保CozeRights权限管理系统与Coze-Studio的无缝整合，最终实现企业级Coze工作平台。

## 📅 总体时间规划

```
阶段1: 基础整合 (1-2周)     ████████████████████ 
阶段2: 核心功能 (1-2月)     ████████████████████████████████████████
阶段3: 企业平台 (3-6月)     ████████████████████████████████████████████████████████████████
```

## 🚀 阶段1：基础整合与紧急需求 (1-2周)

### 第1周：环境搭建与基础适配

#### 任务1.1：开发环境整合 (2天)
**目标**：建立统一的开发环境

**具体任务**：
- [ ] **代码仓库整合**
  ```bash
  # 创建统一仓库结构
  coze-enterprise/
  ├── backend/
  │   ├── cozerights/     # CozeRights服务
  │   ├── coze-studio/    # Coze-Studio服务
  │   └── gateway/        # API网关
  └── frontend/
      ├── enterprise-ui/  # 企业管理界面
      └── coze-studio/    # Coze-Studio前端
  ```

- [ ] **构建系统统一**
  ```yaml
  # docker-compose.yml
  version: '3.8'
  services:
    cozerights:
      build: ./backend/cozerights
      ports: ["8080:8080"]
    coze-studio:
      build: ./backend/coze-studio
      ports: ["8081:8081"]
    api-gateway:
      build: ./backend/gateway
      ports: ["80:80"]
    postgres:
      image: postgres:14
      environment:
        POSTGRES_DB: coze_enterprise
  ```

- [ ] **配置管理统一**
  ```go
  // 统一配置结构
  type Config struct {
      CozeRights  CozeRightsConfig  `yaml:"cozerights"`
      CozeStudio  CozeStudioConfig  `yaml:"coze_studio"`
      Database    DatabaseConfig    `yaml:"database"`
      Gateway     GatewayConfig     `yaml:"gateway"`
  }
  ```

**交付物**：
- 统一的开发环境
- Docker容器化配置
- 统一的配置管理

#### 任务1.2：数据库整合设计 (2天)
**目标**：设计统一的数据库架构

**具体任务**：
- [ ] **数据库架构设计**
  ```sql
  -- 统一数据库架构
  -- CozeRights核心表 (保持不变)
  CREATE SCHEMA cozerights;
  
  -- Coze-Studio表 (增强版)
  CREATE SCHEMA coze_studio;
  
  -- 共享表
  CREATE SCHEMA shared;
  
  -- ID映射表
  CREATE TABLE shared.id_mappings (
      id SERIAL PRIMARY KEY,
      source_system VARCHAR(50),
      source_id VARCHAR(100),
      target_system VARCHAR(50),
      target_id VARCHAR(100),
      resource_type VARCHAR(50),
      created_at TIMESTAMP DEFAULT NOW()
  );
  ```

- [ ] **数据迁移脚本**
  ```go
  // 数据迁移服务
  type MigrationService struct {
      sourceDB *gorm.DB
      targetDB *gorm.DB
  }
  
  func (s *MigrationService) MigrateAll() error {
      // 1. 迁移用户数据
      if err := s.MigrateUsers(); err != nil {
          return err
      }
      
      // 2. 迁移空间数据
      if err := s.MigrateSpaces(); err != nil {
          return err
      }
      
      // 3. 迁移资源数据
      return s.MigrateResources()
  }
  ```

**交付物**：
- 统一数据库架构设计
- 数据迁移脚本
- 数据兼容性测试

#### 任务1.3：API网关搭建 (2天)
**目标**：建立统一的API入口

**具体任务**：
- [ ] **API网关实现**
  ```go
  // API网关路由配置
  type Gateway struct {
      cozeRightsURL string
      cozeStudioURL string
      router        *gin.Engine
  }
  
  func (g *Gateway) SetupRoutes() {
      // CozeRights API
      g.router.Any("/api/v1/*path", g.proxyToCozeRights)
      
      // Coze-Studio API (兼容)
      g.router.Any("/api/legacy/*path", g.proxyToCozeStudio)
      
      // 统一API (新)
      g.router.Any("/api/v2/*path", g.handleUnifiedAPI)
  }
  ```

- [ ] **认证中间件**
  ```go
  func AuthMiddleware() gin.HandlerFunc {
      return func(c *gin.Context) {
          token := c.GetHeader("Authorization")
          
          // 验证JWT Token
          claims, err := validateJWT(token)
          if err != nil {
              c.JSON(401, gin.H{"error": "Unauthorized"})
              c.Abort()
              return
          }
          
          // 设置用户上下文
          c.Set("user_id", claims.UserID)
          c.Set("tenant_id", claims.TenantID)
          c.Next()
      }
  }
  ```

**交付物**：
- API网关服务
- 统一认证中间件
- 路由配置

### 第2周：基础功能对接

#### 任务1.4：用户认证整合 (3天)
**目标**：实现统一的用户认证

**具体任务**：
- [ ] **统一登录接口**
  ```go
  // 统一登录服务
  type UnifiedAuthService struct {
      cozeRightsAuth *CozeRightsAuthService
      cozeStudioAuth *CozeStudioAuthService
      userMapping    *UserMappingService
  }
  
  func (s *UnifiedAuthService) Login(email, password string) (*LoginResponse, error) {
      // 1. 使用CozeRights验证
      user, err := s.cozeRightsAuth.ValidateUser(email, password)
      if err != nil {
          return nil, err
      }
      
      // 2. 生成统一Token
      token, err := s.generateUnifiedToken(user)
      if err != nil {
          return nil, err
      }
      
      // 3. 同步到Coze-Studio
      s.cozeStudioAuth.SyncUserSession(user.ID, token)
      
      return &LoginResponse{
          Token: token,
          User:  user,
      }, nil
  }
  ```

- [ ] **会话管理**
  ```go
  // 会话管理服务
  type SessionManager struct {
      store SessionStore
      ttl   time.Duration
  }
  
  func (m *SessionManager) CreateSession(userID uint, token string) error {
      session := &Session{
          UserID:    userID,
          Token:     token,
          ExpiresAt: time.Now().Add(m.ttl),
      }
      return m.store.Save(session)
  }
  ```

**交付物**：
- 统一登录接口
- 会话管理系统
- Token验证机制

#### 任务1.5：权限检查适配 (2天)
**目标**：实现基础的权限检查适配

**具体任务**：
- [ ] **权限适配器**
  ```go
  // 权限适配器
  type PermissionAdapter struct {
      cozeRightsRBAC *rbac.RBACService
      mapping        map[string]string
  }
  
  func (a *PermissionAdapter) CheckCozeStudioPermission(
      userID string, 
      spaceID string, 
      action string) (bool, error) {
      
      // 1. 转换ID
      workspaceID := a.convertSpaceToWorkspace(spaceID)
      
      // 2. 转换权限
      permission := a.mapping[action]
      
      // 3. 检查权限
      return a.cozeRightsRBAC.CheckWorkspacePermission(
          context.Background(),
          parseUint(userID),
          workspaceID,
          "workspace",
          permission,
      )
  }
  ```

**交付物**：
- 权限适配器
- 权限映射配置
- 基础权限检查

## 🏗️ 阶段2：核心功能实现与深度整合 (1-2月)

### 第3-4周：数据模型统一

#### 任务2.1：用户模型整合 (1周)
**目标**：统一用户数据模型

**具体任务**：
- [ ] **统一用户模型**
  ```go
  // 统一用户模型
  type UnifiedUser struct {
      // CozeRights字段
      ID         uint      `json:"id" gorm:"primaryKey"`
      TenantID   uint      `json:"tenant_id" gorm:"not null;index"`
      Username   string    `json:"username" gorm:"uniqueIndex"`
      Email      string    `json:"email" gorm:"uniqueIndex"`
      SystemRole string    `json:"system_role"`
      
      // Coze-Studio兼容字段
      UserUniqueName string `json:"user_unique_name" gorm:"index"`
      Avatar         string `json:"avatar"`
      DisplayName    string `json:"display_name"`
      
      // 扩展字段
      Profile    UserProfile `json:"profile" gorm:"type:jsonb"`
      Settings   UserSettings `json:"settings" gorm:"type:jsonb"`
      
      // 审计字段
      IsActive   bool      `json:"is_active" gorm:"default:true"`
      CreatedAt  time.Time `json:"created_at"`
      UpdatedAt  time.Time `json:"updated_at"`
      DeletedAt  *time.Time `json:"deleted_at" gorm:"index"`
  }
  ```

- [ ] **用户服务整合**
  ```go
  // 统一用户服务
  type UnifiedUserService struct {
      userRepo     UserRepository
      tenantRepo   TenantRepository
      auditService AuditService
  }
  
  func (s *UnifiedUserService) CreateUser(req *CreateUserRequest) (*UnifiedUser, error) {
      // 1. 验证租户权限
      if !s.validateTenantAccess(req.TenantID) {
          return nil, ErrTenantAccessDenied
      }
      
      // 2. 创建用户
      user := &UnifiedUser{
          TenantID:       req.TenantID,
          Username:       req.Username,
          Email:          req.Email,
          UserUniqueName: req.Username, // 兼容字段
          SystemRole:     "user",
          IsActive:       true,
      }
      
      // 3. 保存用户
      if err := s.userRepo.Create(user); err != nil {
          return nil, err
      }
      
      // 4. 记录审计日志
      s.auditService.LogUserOperation("create", user.ID, req.OperatorID)
      
      return user, nil
  }
  ```

**交付物**：
- 统一用户模型
- 用户服务整合
- 数据迁移完成

#### 任务2.2：工作空间模型整合 (1周)
**目标**：统一工作空间数据模型

**具体任务**：
- [ ] **统一工作空间模型**
  ```go
  // 统一工作空间模型
  type UnifiedWorkspace struct {
      // CozeRights字段
      ID           uint   `json:"id" gorm:"primaryKey"`
      TenantID     uint   `json:"tenant_id" gorm:"not null;index"`
      Name         string `json:"name" gorm:"not null"`
      Code         string `json:"code" gorm:"uniqueIndex"`
      Type         string `json:"type" gorm:"default:'team'"`
      MaxAgents    int    `json:"max_agents" gorm:"default:100"`
      MaxWorkflows int    `json:"max_workflows" gorm:"default:200"`
      MaxPlugins   int    `json:"max_plugins" gorm:"default:50"`
      
      // Coze-Studio兼容字段
      SpaceID     string `json:"space_id" gorm:"index"`
      Description string `json:"description"`
      Avatar      string `json:"avatar"`
      
      // 扩展字段
      Settings    WorkspaceSettings `json:"settings" gorm:"type:jsonb"`
      Metadata    WorkspaceMetadata `json:"metadata" gorm:"type:jsonb"`
      
      // 关联关系
      Members     []WorkspaceMember `json:"members" gorm:"foreignKey:WorkspaceID"`
      Agents      []WorkspaceAgent  `json:"agents" gorm:"foreignKey:WorkspaceID"`
      Workflows   []WorkspaceWorkflow `json:"workflows" gorm:"foreignKey:WorkspaceID"`
      
      // 审计字段
      CreatedBy   uint      `json:"created_by"`
      IsActive    bool      `json:"is_active" gorm:"default:true"`
      CreatedAt   time.Time `json:"created_at"`
      UpdatedAt   time.Time `json:"updated_at"`
      DeletedAt   *time.Time `json:"deleted_at" gorm:"index"`
  }
  ```

**交付物**：
- 统一工作空间模型
- 工作空间服务整合
- 成员管理功能

### 第5-6周：资源管理整合

#### 任务2.3：Agent管理整合 (1周)
**目标**：整合Agent管理功能

**具体任务**：
- [ ] **Agent模型统一**
- [ ] **Agent API整合**
- [ ] **权限控制增强**

#### 任务2.4：Workflow管理整合 (1周)
**目标**：整合Workflow管理功能

**具体任务**：
- [ ] **Workflow模型统一**
- [ ] **执行管理整合**
- [ ] **版本控制增强**

### 第7-8周：前端界面整合

#### 任务2.5：企业管理界面 (2周)
**目标**：开发企业级管理界面

**具体任务**：
- [ ] **租户管理界面**
- [ ] **用户管理界面**
- [ ] **工作空间管理界面**
- [ ] **权限管理界面**

## 🏢 阶段3：企业级平台完善 (3-6月)

### 第9-12周：高级功能实现

#### 任务3.1：高级权限管理 (2周)
- [ ] **自定义角色**
- [ ] **权限模板**
- [ ] **批量权限操作**

#### 任务3.2：审计与监控 (2周)
- [ ] **完整审计日志**
- [ ] **实时监控**
- [ ] **安全告警**

### 第13-16周：性能优化与扩展

#### 任务3.3：性能优化 (2周)
- [ ] **数据库优化**
- [ ] **缓存机制**
- [ ] **API性能优化**

#### 任务3.4：扩展功能 (2周)
- [ ] **多语言支持**
- [ ] **主题定制**
- [ ] **插件系统**

### 第17-24周：企业级部署

#### 任务3.5：部署架构 (4周)
- [ ] **容器化部署**
- [ ] **Kubernetes配置**
- [ ] **CI/CD流水线**

#### 任务3.6：运维监控 (4周)
- [ ] **监控系统**
- [ ] **日志聚合**
- [ ] **告警系统**

## 📊 里程碑与交付物

### 里程碑1：基础整合完成 (第2周末)
- ✅ 开发环境统一
- ✅ 基础认证整合
- ✅ API网关搭建

### 里程碑2：核心功能整合 (第8周末)
- ✅ 数据模型统一
- ✅ 资源管理整合
- ✅ 企业管理界面

### 里程碑3：企业平台发布 (第24周末)
- ✅ 完整企业级功能
- ✅ 生产环境部署
- ✅ 运维监控体系

## 🎯 成功标准

### 功能标准
- [ ] 100% API兼容性
- [ ] 完整的多租户支持
- [ ] 细粒度权限控制
- [ ] 完整的审计日志

### 性能标准
- [ ] API响应时间 < 200ms
- [ ] 支持1000+并发用户
- [ ] 99.9%系统可用性
- [ ] 数据一致性保证

### 安全标准
- [ ] 完整的数据隔离
- [ ] 安全的认证机制
- [ ] 完整的权限验证
- [ ] 审计日志完整性

---

**规划版本**：v1.0.0  
**制定时间**：2025-01-02  
**制定者**：CozeRights开发团队

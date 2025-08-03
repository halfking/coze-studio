package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"cozerights-backend/internal/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// AuditService 审计日志服务接口
type AuditService interface {
	// 权限审计
	LogPermissionCheck(ctx context.Context, req *PermissionCheckLog) error
	LogRoleAssignment(ctx context.Context, req *RoleAssignmentLog) error
	LogWorkspaceAccess(ctx context.Context, req *WorkspaceAccessLog) error
	
	// 资源操作审计
	LogResourceOperation(ctx context.Context, req *ResourceOperationLog) error
	
	// 查询审计日志
	GetAuditLogs(ctx context.Context, filter *AuditLogFilter) ([]models.AuditLog, error)
	GetUserAuditLogs(ctx context.Context, userID uint, filter *AuditLogFilter) ([]models.AuditLog, error)
	GetTenantAuditLogs(ctx context.Context, tenantID uint, filter *AuditLogFilter) ([]models.AuditLog, error)
	GetWorkspaceAuditLogs(ctx context.Context, workspaceID uint, filter *AuditLogFilter) ([]models.AuditLog, error)
}

// AuditServiceImpl 审计日志服务实现
type AuditServiceImpl struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewAuditService 创建审计日志服务
func NewAuditService(db *gorm.DB, logger *zap.Logger) AuditService {
	return &AuditServiceImpl{
		db:     db,
		logger: logger,
	}
}

// PermissionCheckLog 权限检查日志
type PermissionCheckLog struct {
	UserID      uint   `json:"user_id"`
	TenantID    uint   `json:"tenant_id"`
	WorkspaceID *uint  `json:"workspace_id,omitempty"`
	Resource    string `json:"resource"`
	Action      string `json:"action"`
	Allowed     bool   `json:"allowed"`
	Reason      string `json:"reason,omitempty"`
	IP          string `json:"ip,omitempty"`
	UserAgent   string `json:"user_agent,omitempty"`
	Method      string `json:"method,omitempty"`
	Path        string `json:"path,omitempty"`
}

// RoleAssignmentLog 角色分配日志
type RoleAssignmentLog struct {
	UserID      uint   `json:"user_id"`
	TenantID    uint   `json:"tenant_id"`
	RoleID      uint   `json:"role_id"`
	Operation   string `json:"operation"` // assign, revoke
	OperatorID  uint   `json:"operator_id"`
	IP          string `json:"ip,omitempty"`
	UserAgent   string `json:"user_agent,omitempty"`
}

// WorkspaceAccessLog 工作空间访问日志
type WorkspaceAccessLog struct {
	UserID      uint   `json:"user_id"`
	TenantID    uint   `json:"tenant_id"`
	WorkspaceID uint   `json:"workspace_id"`
	Action      string `json:"action"` // join, leave, access
	Role        string `json:"role,omitempty"`
	IP          string `json:"ip,omitempty"`
	UserAgent   string `json:"user_agent,omitempty"`
}

// ResourceOperationLog 资源操作日志
type ResourceOperationLog struct {
	UserID      uint   `json:"user_id"`
	TenantID    uint   `json:"tenant_id"`
	WorkspaceID *uint  `json:"workspace_id,omitempty"`
	Resource    string `json:"resource"`
	ResourceID  string `json:"resource_id"`
	Action      string `json:"action"`
	Status      string `json:"status"` // success, failed
	Message     string `json:"message,omitempty"`
	IP          string `json:"ip,omitempty"`
	UserAgent   string `json:"user_agent,omitempty"`
	Method      string `json:"method,omitempty"`
	Path        string `json:"path,omitempty"`
	Extra       map[string]interface{} `json:"extra,omitempty"`
}

// AuditLogFilter 审计日志过滤器
type AuditLogFilter struct {
	StartTime   *time.Time `json:"start_time,omitempty"`
	EndTime     *time.Time `json:"end_time,omitempty"`
	Action      string     `json:"action,omitempty"`
	Resource    string     `json:"resource,omitempty"`
	Status      string     `json:"status,omitempty"`
	Limit       int        `json:"limit,omitempty"`
	Offset      int        `json:"offset,omitempty"`
}

// LogPermissionCheck 记录权限检查日志
func (a *AuditServiceImpl) LogPermissionCheck(ctx context.Context, req *PermissionCheckLog) error {
	extra := map[string]interface{}{
		"resource": req.Resource,
		"action":   req.Action,
		"allowed":  req.Allowed,
		"reason":   req.Reason,
	}

	extraJSON, err := json.Marshal(extra)
	if err != nil {
		a.logger.Error("Failed to marshal permission check extra data", zap.Error(err))
		extraJSON = []byte("{}")
	}

	auditLog := &models.AuditLog{
		TenantID:    req.TenantID,
		UserID:      req.UserID,
		WorkspaceID: req.WorkspaceID,
		Action:      "permission_check",
		Resource:    req.Resource,
		ResourceID:  fmt.Sprintf("%s:%s", req.Resource, req.Action),
		Method:      req.Method,
		Path:        req.Path,
		Status:      getStatusFromAllowed(req.Allowed),
		Message:     req.Reason,
		IP:          req.IP,
		UserAgent:   req.UserAgent,
		Extra:       string(extraJSON),
	}

	if err := a.db.WithContext(ctx).Create(auditLog).Error; err != nil {
		a.logger.Error("Failed to create permission check audit log", 
			zap.Error(err),
			zap.Uint("user_id", req.UserID),
			zap.String("resource", req.Resource),
			zap.String("action", req.Action))
		return fmt.Errorf("failed to create permission check audit log: %w", err)
	}

	return nil
}

// LogRoleAssignment 记录角色分配日志
func (a *AuditServiceImpl) LogRoleAssignment(ctx context.Context, req *RoleAssignmentLog) error {
	extra := map[string]interface{}{
		"role_id":     req.RoleID,
		"operation":   req.Operation,
		"operator_id": req.OperatorID,
	}

	extraJSON, err := json.Marshal(extra)
	if err != nil {
		a.logger.Error("Failed to marshal role assignment extra data", zap.Error(err))
		extraJSON = []byte("{}")
	}

	auditLog := &models.AuditLog{
		TenantID:   req.TenantID,
		UserID:     req.UserID,
		Action:     "role_assignment",
		Resource:   "role",
		ResourceID: fmt.Sprintf("%d", req.RoleID),
		Status:     "success",
		Message:    fmt.Sprintf("Role %s by user %d", req.Operation, req.OperatorID),
		IP:         req.IP,
		UserAgent:  req.UserAgent,
		Extra:      string(extraJSON),
	}

	if err := a.db.WithContext(ctx).Create(auditLog).Error; err != nil {
		a.logger.Error("Failed to create role assignment audit log", 
			zap.Error(err),
			zap.Uint("user_id", req.UserID),
			zap.Uint("role_id", req.RoleID))
		return fmt.Errorf("failed to create role assignment audit log: %w", err)
	}

	return nil
}

// LogWorkspaceAccess 记录工作空间访问日志
func (a *AuditServiceImpl) LogWorkspaceAccess(ctx context.Context, req *WorkspaceAccessLog) error {
	extra := map[string]interface{}{
		"action": req.Action,
		"role":   req.Role,
	}

	extraJSON, err := json.Marshal(extra)
	if err != nil {
		a.logger.Error("Failed to marshal workspace access extra data", zap.Error(err))
		extraJSON = []byte("{}")
	}

	auditLog := &models.AuditLog{
		TenantID:    req.TenantID,
		UserID:      req.UserID,
		WorkspaceID: &req.WorkspaceID,
		Action:      "workspace_access",
		Resource:    "workspace",
		ResourceID:  fmt.Sprintf("%d", req.WorkspaceID),
		Status:      "success",
		Message:     fmt.Sprintf("Workspace %s with role %s", req.Action, req.Role),
		IP:          req.IP,
		UserAgent:   req.UserAgent,
		Extra:       string(extraJSON),
	}

	if err := a.db.WithContext(ctx).Create(auditLog).Error; err != nil {
		a.logger.Error("Failed to create workspace access audit log", 
			zap.Error(err),
			zap.Uint("user_id", req.UserID),
			zap.Uint("workspace_id", req.WorkspaceID))
		return fmt.Errorf("failed to create workspace access audit log: %w", err)
	}

	return nil
}

// LogResourceOperation 记录资源操作日志
func (a *AuditServiceImpl) LogResourceOperation(ctx context.Context, req *ResourceOperationLog) error {
	extraJSON, err := json.Marshal(req.Extra)
	if err != nil {
		a.logger.Error("Failed to marshal resource operation extra data", zap.Error(err))
		extraJSON = []byte("{}")
	}

	auditLog := &models.AuditLog{
		TenantID:    req.TenantID,
		UserID:      req.UserID,
		WorkspaceID: req.WorkspaceID,
		Action:      req.Action,
		Resource:    req.Resource,
		ResourceID:  req.ResourceID,
		Method:      req.Method,
		Path:        req.Path,
		Status:      req.Status,
		Message:     req.Message,
		IP:          req.IP,
		UserAgent:   req.UserAgent,
		Extra:       string(extraJSON),
	}

	if err := a.db.WithContext(ctx).Create(auditLog).Error; err != nil {
		a.logger.Error("Failed to create resource operation audit log", 
			zap.Error(err),
			zap.Uint("user_id", req.UserID),
			zap.String("resource", req.Resource),
			zap.String("action", req.Action))
		return fmt.Errorf("failed to create resource operation audit log: %w", err)
	}

	return nil
}

// GetAuditLogs 获取审计日志
func (a *AuditServiceImpl) GetAuditLogs(ctx context.Context, filter *AuditLogFilter) ([]models.AuditLog, error) {
	query := a.db.WithContext(ctx).Model(&models.AuditLog{})

	query = a.applyFilter(query, filter)

	var logs []models.AuditLog
	if err := query.Order("created_at DESC").Find(&logs).Error; err != nil {
		return nil, fmt.Errorf("failed to get audit logs: %w", err)
	}

	return logs, nil
}

// GetUserAuditLogs 获取用户审计日志
func (a *AuditServiceImpl) GetUserAuditLogs(ctx context.Context, userID uint, filter *AuditLogFilter) ([]models.AuditLog, error) {
	query := a.db.WithContext(ctx).Model(&models.AuditLog{}).Where("user_id = ?", userID)

	query = a.applyFilter(query, filter)

	var logs []models.AuditLog
	if err := query.Order("created_at DESC").Find(&logs).Error; err != nil {
		return nil, fmt.Errorf("failed to get user audit logs: %w", err)
	}

	return logs, nil
}

// GetTenantAuditLogs 获取租户审计日志
func (a *AuditServiceImpl) GetTenantAuditLogs(ctx context.Context, tenantID uint, filter *AuditLogFilter) ([]models.AuditLog, error) {
	query := a.db.WithContext(ctx).Model(&models.AuditLog{}).Where("tenant_id = ?", tenantID)

	query = a.applyFilter(query, filter)

	var logs []models.AuditLog
	if err := query.Order("created_at DESC").Find(&logs).Error; err != nil {
		return nil, fmt.Errorf("failed to get tenant audit logs: %w", err)
	}

	return logs, nil
}

// GetWorkspaceAuditLogs 获取工作空间审计日志
func (a *AuditServiceImpl) GetWorkspaceAuditLogs(ctx context.Context, workspaceID uint, filter *AuditLogFilter) ([]models.AuditLog, error) {
	query := a.db.WithContext(ctx).Model(&models.AuditLog{}).Where("workspace_id = ?", workspaceID)

	query = a.applyFilter(query, filter)

	var logs []models.AuditLog
	if err := query.Order("created_at DESC").Find(&logs).Error; err != nil {
		return nil, fmt.Errorf("failed to get workspace audit logs: %w", err)
	}

	return logs, nil
}

// applyFilter 应用过滤条件
func (a *AuditServiceImpl) applyFilter(query *gorm.DB, filter *AuditLogFilter) *gorm.DB {
	if filter == nil {
		return query.Limit(100) // 默认限制
	}

	if filter.StartTime != nil {
		query = query.Where("created_at >= ?", *filter.StartTime)
	}

	if filter.EndTime != nil {
		query = query.Where("created_at <= ?", *filter.EndTime)
	}

	if filter.Action != "" {
		query = query.Where("action = ?", filter.Action)
	}

	if filter.Resource != "" {
		query = query.Where("resource = ?", filter.Resource)
	}

	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}

	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	} else {
		query = query.Limit(100) // 默认限制
	}

	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	return query
}

// getStatusFromAllowed 从权限检查结果获取状态
func getStatusFromAllowed(allowed bool) string {
	if allowed {
		return "success"
	}
	return "denied"
}

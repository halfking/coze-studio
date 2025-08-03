package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// CacheService 缓存服务接口
type CacheService interface {
	// 用户权限缓存
	GetUserPermissions(ctx context.Context, userID uint) (*UserPermissionCache, error)
	SetUserPermissions(ctx context.Context, userID uint, cache *UserPermissionCache, ttl time.Duration) error
	InvalidateUserPermissions(ctx context.Context, userID uint) error

	// 角色权限缓存
	GetRolePermissions(ctx context.Context, roleID uint) (*RolePermissionCache, error)
	SetRolePermissions(ctx context.Context, roleID uint, cache *RolePermissionCache, ttl time.Duration) error
	InvalidateRolePermissions(ctx context.Context, roleID uint) error

	// 工作空间成员缓存
	GetWorkspaceMember(ctx context.Context, workspaceID, userID uint) (*WorkspaceMemberCache, error)
	SetWorkspaceMember(ctx context.Context, workspaceID, userID uint, cache *WorkspaceMemberCache, ttl time.Duration) error
	InvalidateWorkspaceMember(ctx context.Context, workspaceID, userID uint) error

	// 批量操作
	InvalidateUserCache(ctx context.Context, userID uint) error
	InvalidateTenantCache(ctx context.Context, tenantID uint) error
	InvalidateWorkspaceCache(ctx context.Context, workspaceID uint) error
}

// RedisCacheService Redis缓存服务实现
type RedisCacheService struct {
	client *redis.Client
}

// NewRedisCacheService 创建Redis缓存服务
func NewRedisCacheService(client *redis.Client) CacheService {
	return &RedisCacheService{client: client}
}

// UserPermissionCache 用户权限缓存结构
type UserPermissionCache struct {
	UserID      uint     `json:"user_id"`
	TenantID    uint     `json:"tenant_id"`
	SystemRole  string   `json:"system_role"`
	Permissions []string `json:"permissions"`
	Roles       []uint   `json:"roles"`
	CachedAt    int64    `json:"cached_at"`
}

// RolePermissionCache 角色权限缓存结构
type RolePermissionCache struct {
	RoleID      uint     `json:"role_id"`
	TenantID    uint     `json:"tenant_id"`
	Permissions []string `json:"permissions"`
	CachedAt    int64    `json:"cached_at"`
}

// WorkspaceMemberCache 工作空间成员缓存结构
type WorkspaceMemberCache struct {
	WorkspaceID uint     `json:"workspace_id"`
	UserID      uint     `json:"user_id"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
	CachedAt    int64    `json:"cached_at"`
}

// 缓存键生成函数
func (r *RedisCacheService) userPermissionKey(userID uint) string {
	return fmt.Sprintf("rbac:user_permissions:%d", userID)
}

func (r *RedisCacheService) rolePermissionKey(roleID uint) string {
	return fmt.Sprintf("rbac:role_permissions:%d", roleID)
}

func (r *RedisCacheService) workspaceMemberKey(workspaceID, userID uint) string {
	return fmt.Sprintf("rbac:workspace_member:%d:%d", workspaceID, userID)
}

func (r *RedisCacheService) userCachePattern(userID uint) string {
	return fmt.Sprintf("rbac:*:%d", userID)
}

func (r *RedisCacheService) tenantCachePattern(tenantID uint) string {
	return fmt.Sprintf("rbac:tenant:%d:*", tenantID)
}

func (r *RedisCacheService) workspaceCachePattern(workspaceID uint) string {
	return fmt.Sprintf("rbac:workspace:%d:*", workspaceID)
}

// GetUserPermissions 获取用户权限缓存
func (r *RedisCacheService) GetUserPermissions(ctx context.Context, userID uint) (*UserPermissionCache, error) {
	key := r.userPermissionKey(userID)
	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // 缓存未命中
		}
		return nil, fmt.Errorf("failed to get user permissions from cache: %w", err)
	}

	var cache UserPermissionCache
	if err := json.Unmarshal([]byte(data), &cache); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user permissions cache: %w", err)
	}

	return &cache, nil
}

// SetUserPermissions 设置用户权限缓存
func (r *RedisCacheService) SetUserPermissions(ctx context.Context, userID uint, cache *UserPermissionCache, ttl time.Duration) error {
	key := r.userPermissionKey(userID)
	cache.CachedAt = time.Now().Unix()

	data, err := json.Marshal(cache)
	if err != nil {
		return fmt.Errorf("failed to marshal user permissions cache: %w", err)
	}

	if err := r.client.Set(ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("failed to set user permissions cache: %w", err)
	}

	return nil
}

// InvalidateUserPermissions 失效用户权限缓存
func (r *RedisCacheService) InvalidateUserPermissions(ctx context.Context, userID uint) error {
	key := r.userPermissionKey(userID)
	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to invalidate user permissions cache: %w", err)
	}
	return nil
}

// GetRolePermissions 获取角色权限缓存
func (r *RedisCacheService) GetRolePermissions(ctx context.Context, roleID uint) (*RolePermissionCache, error) {
	key := r.rolePermissionKey(roleID)
	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // 缓存未命中
		}
		return nil, fmt.Errorf("failed to get role permissions from cache: %w", err)
	}

	var cache RolePermissionCache
	if err := json.Unmarshal([]byte(data), &cache); err != nil {
		return nil, fmt.Errorf("failed to unmarshal role permissions cache: %w", err)
	}

	return &cache, nil
}

// SetRolePermissions 设置角色权限缓存
func (r *RedisCacheService) SetRolePermissions(ctx context.Context, roleID uint, cache *RolePermissionCache, ttl time.Duration) error {
	key := r.rolePermissionKey(roleID)
	cache.CachedAt = time.Now().Unix()

	data, err := json.Marshal(cache)
	if err != nil {
		return fmt.Errorf("failed to marshal role permissions cache: %w", err)
	}

	if err := r.client.Set(ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("failed to set role permissions cache: %w", err)
	}

	return nil
}

// InvalidateRolePermissions 失效角色权限缓存
func (r *RedisCacheService) InvalidateRolePermissions(ctx context.Context, roleID uint) error {
	key := r.rolePermissionKey(roleID)
	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to invalidate role permissions cache: %w", err)
	}
	return nil
}

// GetWorkspaceMember 获取工作空间成员缓存
func (r *RedisCacheService) GetWorkspaceMember(ctx context.Context, workspaceID, userID uint) (*WorkspaceMemberCache, error) {
	key := r.workspaceMemberKey(workspaceID, userID)
	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // 缓存未命中
		}
		return nil, fmt.Errorf("failed to get workspace member from cache: %w", err)
	}

	var cache WorkspaceMemberCache
	if err := json.Unmarshal([]byte(data), &cache); err != nil {
		return nil, fmt.Errorf("failed to unmarshal workspace member cache: %w", err)
	}

	return &cache, nil
}

// SetWorkspaceMember 设置工作空间成员缓存
func (r *RedisCacheService) SetWorkspaceMember(ctx context.Context, workspaceID, userID uint, cache *WorkspaceMemberCache, ttl time.Duration) error {
	key := r.workspaceMemberKey(workspaceID, userID)
	cache.CachedAt = time.Now().Unix()

	data, err := json.Marshal(cache)
	if err != nil {
		return fmt.Errorf("failed to marshal workspace member cache: %w", err)
	}

	if err := r.client.Set(ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("failed to set workspace member cache: %w", err)
	}

	return nil
}

// InvalidateWorkspaceMember 失效工作空间成员缓存
func (r *RedisCacheService) InvalidateWorkspaceMember(ctx context.Context, workspaceID, userID uint) error {
	key := r.workspaceMemberKey(workspaceID, userID)
	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to invalidate workspace member cache: %w", err)
	}
	return nil
}

// InvalidateUserCache 失效用户相关的所有缓存
func (r *RedisCacheService) InvalidateUserCache(ctx context.Context, userID uint) error {
	// 删除用户权限缓存
	if err := r.InvalidateUserPermissions(ctx, userID); err != nil {
		return err
	}

	// 删除用户相关的工作空间成员缓存
	pattern := fmt.Sprintf("rbac:workspace_member:*:%d", userID)
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to get user workspace member cache keys: %w", err)
	}

	if len(keys) > 0 {
		if err := r.client.Del(ctx, keys...).Err(); err != nil {
			return fmt.Errorf("failed to delete user workspace member caches: %w", err)
		}
	}

	return nil
}

// InvalidateTenantCache 失效租户相关的所有缓存
func (r *RedisCacheService) InvalidateTenantCache(ctx context.Context, tenantID uint) error {
	pattern := r.tenantCachePattern(tenantID)
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to get tenant cache keys: %w", err)
	}

	if len(keys) > 0 {
		if err := r.client.Del(ctx, keys...).Err(); err != nil {
			return fmt.Errorf("failed to delete tenant caches: %w", err)
		}
	}

	return nil
}

// InvalidateWorkspaceCache 失效工作空间相关的所有缓存
func (r *RedisCacheService) InvalidateWorkspaceCache(ctx context.Context, workspaceID uint) error {
	pattern := r.workspaceCachePattern(workspaceID)
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to get workspace cache keys: %w", err)
	}

	if len(keys) > 0 {
		if err := r.client.Del(ctx, keys...).Err(); err != nil {
			return fmt.Errorf("failed to delete workspace caches: %w", err)
		}
	}

	return nil
}

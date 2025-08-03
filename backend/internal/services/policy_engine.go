package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"cozerights-backend/internal/models"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// PolicyEngine 动态权限策略引擎
type PolicyEngine struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewPolicyEngine 创建策略引擎
func NewPolicyEngine(db *gorm.DB, logger *zap.Logger) *PolicyEngine {
	return &PolicyEngine{
		db:     db,
		logger: logger,
	}
}

// PolicyContext 策略上下文
type PolicyContext struct {
	UserID      uint                   `json:"user_id"`
	TenantID    uint                   `json:"tenant_id"`
	WorkspaceID *uint                  `json:"workspace_id"`
	Resource    string                 `json:"resource"`
	Action      string                 `json:"action"`
	ResourceID  string                 `json:"resource_id"`
	Time        time.Time              `json:"time"`
	IP          string                 `json:"ip"`
	UserAgent   string                 `json:"user_agent"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// PolicyResult 策略执行结果
type PolicyResult struct {
	Allow      bool                   `json:"allow"`
	Reason     string                 `json:"reason"`
	Conditions []string               `json:"conditions,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// Policy 权限策略定义
type Policy struct {
	ID          uint                   `json:"id" gorm:"primaryKey"`
	TenantID    uint                   `json:"tenant_id" gorm:"index"`
	Name        string                 `json:"name" gorm:"size:100;not null"`
	Description string                 `json:"description" gorm:"size:500"`
	Type        string                 `json:"type" gorm:"size:50;not null"` // allow, deny, conditional
	Priority    int                    `json:"priority" gorm:"default:0"`    // 优先级，数字越大优先级越高
	Conditions  string                 `json:"conditions" gorm:"type:text"`  // JSON格式的条件
	Effect      string                 `json:"effect" gorm:"type:text"`      // JSON格式的效果
	Enabled     bool                   `json:"enabled" gorm:"default:true"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// PolicyCondition 策略条件
type PolicyCondition struct {
	Field    string      `json:"field"`    // 字段名
	Operator string      `json:"operator"` // 操作符: eq, ne, gt, lt, gte, lte, in, not_in, contains, regex
	Value    interface{} `json:"value"`    // 比较值
	Logic    string      `json:"logic"`    // 逻辑操作: and, or
}

// PolicyEffect 策略效果
type PolicyEffect struct {
	Allow       bool                   `json:"allow"`
	Reason      string                 `json:"reason"`
	Quota       *QuotaLimit            `json:"quota,omitempty"`
	RateLimit   *RateLimit             `json:"rate_limit,omitempty"`
	TimeWindow  *TimeWindow            `json:"time_window,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// QuotaLimit 配额限制
type QuotaLimit struct {
	Resource string `json:"resource"`
	Limit    int64  `json:"limit"`
	Period   string `json:"period"` // daily, monthly, yearly
	Unit     string `json:"unit"`   // tokens, requests, bytes
}

// RateLimit 速率限制
type RateLimit struct {
	Requests int           `json:"requests"`
	Window   time.Duration `json:"window"`
}

// TimeWindow 时间窗口限制
type TimeWindow struct {
	StartTime string `json:"start_time"` // HH:MM
	EndTime   string `json:"end_time"`   // HH:MM
	Timezone  string `json:"timezone"`
	Days      []int  `json:"days"` // 0=Sunday, 1=Monday, ...
}

// EvaluatePolicy 评估策略
func (pe *PolicyEngine) EvaluatePolicy(ctx context.Context, policyCtx PolicyContext) (*PolicyResult, error) {
	// 获取适用的策略
	policies, err := pe.getApplicablePolicies(policyCtx.TenantID, policyCtx.Resource, policyCtx.Action)
	if err != nil {
		return nil, fmt.Errorf("failed to get policies: %w", err)
	}

	// 默认结果：允许
	result := &PolicyResult{
		Allow:  true,
		Reason: "No applicable policies found",
	}

	// 按优先级排序并评估策略
	for _, policy := range policies {
		if !policy.Enabled {
			continue
		}

		// 评估策略条件
		matches, err := pe.evaluateConditions(policy.Conditions, policyCtx)
		if err != nil {
			pe.logger.Error("Failed to evaluate policy conditions",
				zap.Error(err),
				zap.Uint("policy_id", policy.ID),
				zap.String("policy_name", policy.Name))
			continue
		}

		if matches {
			// 应用策略效果
			effect, err := pe.parseEffect(policy.Effect)
			if err != nil {
				pe.logger.Error("Failed to parse policy effect",
					zap.Error(err),
					zap.Uint("policy_id", policy.ID))
				continue
			}

			// 根据策略类型处理
			switch policy.Type {
			case "deny":
				return &PolicyResult{
					Allow:  false,
					Reason: fmt.Sprintf("Denied by policy: %s", policy.Name),
				}, nil
			case "allow":
				result.Allow = true
				result.Reason = fmt.Sprintf("Allowed by policy: %s", policy.Name)
			case "conditional":
				// 应用条件性策略
				if err := pe.applyConditionalPolicy(result, effect, policyCtx); err != nil {
					pe.logger.Error("Failed to apply conditional policy",
						zap.Error(err),
						zap.Uint("policy_id", policy.ID))
				}
			}

			// 记录策略应用
			pe.logger.Info("Policy applied",
				zap.Uint("policy_id", policy.ID),
				zap.String("policy_name", policy.Name),
				zap.String("policy_type", policy.Type),
				zap.Bool("allow", result.Allow))
		}
	}

	return result, nil
}

// getApplicablePolicies 获取适用的策略
func (pe *PolicyEngine) getApplicablePolicies(tenantID uint, resource, action string) ([]Policy, error) {
	var policies []Policy
	
	// 查询适用的策略，按优先级降序排列
	err := pe.db.Where("tenant_id = ? AND enabled = ?", tenantID, true).
		Order("priority DESC, created_at ASC").
		Find(&policies).Error
	
	if err != nil {
		return nil, err
	}

	// 过滤适用于当前资源和操作的策略
	var applicablePolicies []Policy
	for _, policy := range policies {
		if pe.isPolicyApplicable(policy, resource, action) {
			applicablePolicies = append(applicablePolicies, policy)
		}
	}

	return applicablePolicies, nil
}

// isPolicyApplicable 检查策略是否适用
func (pe *PolicyEngine) isPolicyApplicable(policy Policy, resource, action string) bool {
	// 解析策略条件，检查是否包含资源和操作匹配
	var conditions []PolicyCondition
	if err := json.Unmarshal([]byte(policy.Conditions), &conditions); err != nil {
		return false
	}

	resourceMatch := false
	actionMatch := false

	for _, condition := range conditions {
		if condition.Field == "resource" {
			if pe.evaluateCondition(condition, resource) {
				resourceMatch = true
			}
		}
		if condition.Field == "action" {
			if pe.evaluateCondition(condition, action) {
				actionMatch = true
			}
		}
	}

	// 如果没有指定资源或操作条件，则认为适用于所有
	if !pe.hasResourceCondition(conditions) {
		resourceMatch = true
	}
	if !pe.hasActionCondition(conditions) {
		actionMatch = true
	}

	return resourceMatch && actionMatch
}

// evaluateConditions 评估策略条件
func (pe *PolicyEngine) evaluateConditions(conditionsJSON string, ctx PolicyContext) (bool, error) {
	if conditionsJSON == "" {
		return true, nil
	}

	var conditions []PolicyCondition
	if err := json.Unmarshal([]byte(conditionsJSON), &conditions); err != nil {
		return false, fmt.Errorf("failed to parse conditions: %w", err)
	}

	if len(conditions) == 0 {
		return true, nil
	}

	// 评估所有条件
	results := make([]bool, len(conditions))
	for i, condition := range conditions {
		value := pe.getContextValue(condition.Field, ctx)
		results[i] = pe.evaluateCondition(condition, value)
	}

	// 应用逻辑操作
	return pe.applyLogic(conditions, results), nil
}

// getContextValue 从上下文中获取字段值
func (pe *PolicyEngine) getContextValue(field string, ctx PolicyContext) interface{} {
	switch field {
	case "user_id":
		return ctx.UserID
	case "tenant_id":
		return ctx.TenantID
	case "workspace_id":
		return ctx.WorkspaceID
	case "resource":
		return ctx.Resource
	case "action":
		return ctx.Action
	case "resource_id":
		return ctx.ResourceID
	case "time":
		return ctx.Time
	case "hour":
		return ctx.Time.Hour()
	case "day_of_week":
		return int(ctx.Time.Weekday())
	case "ip":
		return ctx.IP
	case "user_agent":
		return ctx.UserAgent
	default:
		// 从元数据中查找
		if ctx.Metadata != nil {
			return ctx.Metadata[field]
		}
		return nil
	}
}

// evaluateCondition 评估单个条件
func (pe *PolicyEngine) evaluateCondition(condition PolicyCondition, value interface{}) bool {
	switch condition.Operator {
	case "eq":
		return pe.compareValues(value, condition.Value) == 0
	case "ne":
		return pe.compareValues(value, condition.Value) != 0
	case "gt":
		return pe.compareValues(value, condition.Value) > 0
	case "lt":
		return pe.compareValues(value, condition.Value) < 0
	case "gte":
		return pe.compareValues(value, condition.Value) >= 0
	case "lte":
		return pe.compareValues(value, condition.Value) <= 0
	case "in":
		return pe.valueInList(value, condition.Value)
	case "not_in":
		return !pe.valueInList(value, condition.Value)
	case "contains":
		return pe.stringContains(value, condition.Value)
	case "regex":
		return pe.regexMatch(value, condition.Value)
	default:
		return false
	}
}

// compareValues 比较两个值
func (pe *PolicyEngine) compareValues(a, b interface{}) int {
	// 简化的比较实现，实际应该更完善
	aStr := fmt.Sprintf("%v", a)
	bStr := fmt.Sprintf("%v", b)
	if aStr < bStr {
		return -1
	} else if aStr > bStr {
		return 1
	}
	return 0
}

// valueInList 检查值是否在列表中
func (pe *PolicyEngine) valueInList(value, list interface{}) bool {
	listSlice, ok := list.([]interface{})
	if !ok {
		return false
	}
	
	for _, item := range listSlice {
		if pe.compareValues(value, item) == 0 {
			return true
		}
	}
	return false
}

// stringContains 检查字符串包含
func (pe *PolicyEngine) stringContains(value, substr interface{}) bool {
	valueStr := fmt.Sprintf("%v", value)
	substrStr := fmt.Sprintf("%v", substr)
	return strings.Contains(valueStr, substrStr)
}

// regexMatch 正则表达式匹配
func (pe *PolicyEngine) regexMatch(value, pattern interface{}) bool {
	// TODO: 实现正则表达式匹配
	return false
}

// applyLogic 应用逻辑操作
func (pe *PolicyEngine) applyLogic(conditions []PolicyCondition, results []bool) bool {
	if len(results) == 0 {
		return true
	}
	
	if len(results) == 1 {
		return results[0]
	}

	// 默认使用AND逻辑
	result := results[0]
	for i := 1; i < len(results); i++ {
		logic := "and"
		if i-1 < len(conditions) && conditions[i-1].Logic != "" {
			logic = conditions[i-1].Logic
		}
		
		if logic == "or" {
			result = result || results[i]
		} else {
			result = result && results[i]
		}
	}
	
	return result
}

// parseEffect 解析策略效果
func (pe *PolicyEngine) parseEffect(effectJSON string) (*PolicyEffect, error) {
	if effectJSON == "" {
		return &PolicyEffect{Allow: true}, nil
	}

	var effect PolicyEffect
	if err := json.Unmarshal([]byte(effectJSON), &effect); err != nil {
		return nil, fmt.Errorf("failed to parse effect: %w", err)
	}

	return &effect, nil
}

// applyConditionalPolicy 应用条件性策略
func (pe *PolicyEngine) applyConditionalPolicy(result *PolicyResult, effect *PolicyEffect, ctx PolicyContext) error {
	// 应用配额限制
	if effect.Quota != nil {
		// TODO: 检查配额使用情况
	}

	// 应用速率限制
	if effect.RateLimit != nil {
		// TODO: 检查速率限制
	}

	// 应用时间窗口限制
	if effect.TimeWindow != nil {
		if !pe.isInTimeWindow(effect.TimeWindow, ctx.Time) {
			result.Allow = false
			result.Reason = "Outside allowed time window"
			return nil
		}
	}

	// 合并元数据
	if effect.Metadata != nil {
		if result.Metadata == nil {
			result.Metadata = make(map[string]interface{})
		}
		for k, v := range effect.Metadata {
			result.Metadata[k] = v
		}
	}

	return nil
}

// isInTimeWindow 检查是否在时间窗口内
func (pe *PolicyEngine) isInTimeWindow(window *TimeWindow, t time.Time) bool {
	// TODO: 实现时间窗口检查
	return true
}

// hasResourceCondition 检查是否有资源条件
func (pe *PolicyEngine) hasResourceCondition(conditions []PolicyCondition) bool {
	for _, condition := range conditions {
		if condition.Field == "resource" {
			return true
		}
	}
	return false
}

// hasActionCondition 检查是否有操作条件
func (pe *PolicyEngine) hasActionCondition(conditions []PolicyCondition) bool {
	for _, condition := range conditions {
		if condition.Field == "action" {
			return true
		}
	}
	return false
}

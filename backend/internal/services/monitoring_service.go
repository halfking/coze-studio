package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"cozerights-backend/internal/models"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// MonitoringService 监控服务
type MonitoringService struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewMonitoringService 创建监控服务
func NewMonitoringService(db *gorm.DB, logger *zap.Logger) *MonitoringService {
	return &MonitoringService{
		db:     db,
		logger: logger,
	}
}

// AlertRule 告警规则
type AlertRule struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	TenantID    uint      `json:"tenant_id" gorm:"index"`
	Name        string    `json:"name" gorm:"size:100;not null"`
	Description string    `json:"description" gorm:"size:500"`
	Type        string    `json:"type" gorm:"size:50;not null"` // quota, rate_limit, error_rate, cost
	Metric      string    `json:"metric" gorm:"size:100;not null"`
	Condition   string    `json:"condition" gorm:"type:text"` // JSON格式的条件
	Threshold   float64   `json:"threshold"`
	Severity    string    `json:"severity" gorm:"size:20;default:'warning'"` // info, warning, critical
	Enabled     bool      `json:"enabled" gorm:"default:true"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Alert 告警记录
type Alert struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	TenantID    uint      `json:"tenant_id" gorm:"index"`
	RuleID      uint      `json:"rule_id" gorm:"index"`
	Title       string    `json:"title" gorm:"size:200;not null"`
	Message     string    `json:"message" gorm:"type:text"`
	Severity    string    `json:"severity" gorm:"size:20"`
	Status      string    `json:"status" gorm:"size:20;default:'active'"` // active, resolved, suppressed
	Value       float64   `json:"value"`
	Threshold   float64   `json:"threshold"`
	Metadata    string    `json:"metadata" gorm:"type:text"`
	ResolvedAt  *time.Time `json:"resolved_at"`
	CreatedAt   time.Time `json:"created_at" gorm:"index"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// MetricData 指标数据
type MetricData struct {
	TenantID    uint                   `json:"tenant_id"`
	Metric      string                 `json:"metric"`
	Value       float64                `json:"value"`
	Labels      map[string]string      `json:"labels"`
	Metadata    map[string]interface{} `json:"metadata"`
	Timestamp   time.Time              `json:"timestamp"`
}

// AlertCondition 告警条件
type AlertCondition struct {
	Operator  string  `json:"operator"` // gt, gte, lt, lte, eq, ne
	Threshold float64 `json:"threshold"`
	Duration  string  `json:"duration"` // 持续时间，如 "5m", "1h"
}

// NotificationChannel 通知渠道
type NotificationChannel struct {
	ID       uint   `json:"id" gorm:"primaryKey"`
	TenantID uint   `json:"tenant_id" gorm:"index"`
	Name     string `json:"name" gorm:"size:100;not null"`
	Type     string `json:"type" gorm:"size:50;not null"` // email, webhook, slack
	Config   string `json:"config" gorm:"type:text"`      // JSON格式的配置
	Enabled  bool   `json:"enabled" gorm:"default:true"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CheckAlerts 检查告警
func (ms *MonitoringService) CheckAlerts(ctx context.Context) error {
	// 获取所有启用的告警规则
	var rules []AlertRule
	if err := ms.db.Where("enabled = ?", true).Find(&rules).Error; err != nil {
		return fmt.Errorf("failed to get alert rules: %w", err)
	}

	for _, rule := range rules {
		if err := ms.checkAlertRule(ctx, rule); err != nil {
			ms.logger.Error("Failed to check alert rule",
				zap.Error(err),
				zap.Uint("rule_id", rule.ID),
				zap.String("rule_name", rule.Name))
		}
	}

	return nil
}

// checkAlertRule 检查单个告警规则
func (ms *MonitoringService) checkAlertRule(ctx context.Context, rule AlertRule) error {
	// 获取指标数据
	value, err := ms.getMetricValue(rule.TenantID, rule.Metric)
	if err != nil {
		return fmt.Errorf("failed to get metric value: %w", err)
	}

	// 解析告警条件
	var condition AlertCondition
	if err := json.Unmarshal([]byte(rule.Condition), &condition); err != nil {
		return fmt.Errorf("failed to parse alert condition: %w", err)
	}

	// 检查是否触发告警
	triggered := ms.evaluateCondition(value, condition)

	if triggered {
		// 检查是否已经有活跃的告警
		var existingAlert Alert
		err := ms.db.Where("rule_id = ? AND status = ?", rule.ID, "active").
			First(&existingAlert).Error

		if err == gorm.ErrRecordNotFound {
			// 创建新告警
			return ms.createAlert(rule, value, condition.Threshold)
		} else if err != nil {
			return fmt.Errorf("failed to check existing alert: %w", err)
		}
		// 如果已有活跃告警，不重复创建
	} else {
		// 检查是否需要解决告警
		return ms.resolveAlerts(rule.ID)
	}

	return nil
}

// getMetricValue 获取指标值
func (ms *MonitoringService) getMetricValue(tenantID uint, metric string) (float64, error) {
	switch metric {
	case "api_requests_per_minute":
		return ms.getAPIRequestsPerMinute(tenantID)
	case "quota_usage_percentage":
		return ms.getQuotaUsagePercentage(tenantID)
	case "error_rate_percentage":
		return ms.getErrorRatePercentage(tenantID)
	case "monthly_cost":
		return ms.getMonthlyCost(tenantID)
	case "active_users":
		return ms.getActiveUsers(tenantID)
	case "storage_usage_gb":
		return ms.getStorageUsage(tenantID)
	default:
		return 0, fmt.Errorf("unknown metric: %s", metric)
	}
}

// getAPIRequestsPerMinute 获取每分钟API请求数
func (ms *MonitoringService) getAPIRequestsPerMinute(tenantID uint) (float64, error) {
	var count int64
	oneMinuteAgo := time.Now().Add(-time.Minute)
	
	err := ms.db.Model(&models.AuditLog{}).
		Where("tenant_id = ? AND created_at >= ?", tenantID, oneMinuteAgo).
		Count(&count).Error
	
	return float64(count), err
}

// getQuotaUsagePercentage 获取配额使用百分比
func (ms *MonitoringService) getQuotaUsagePercentage(tenantID uint) (float64, error) {
	// 简化实现：假设API调用配额为10000/月
	var count int64
	startOfMonth := time.Now().AddDate(0, 0, -time.Now().Day()+1)
	
	err := ms.db.Model(&models.AuditLog{}).
		Where("tenant_id = ? AND created_at >= ?", tenantID, startOfMonth).
		Count(&count).Error
	
	if err != nil {
		return 0, err
	}
	
	quota := 10000.0 // 假设的配额
	return (float64(count) / quota) * 100, nil
}

// getErrorRatePercentage 获取错误率百分比
func (ms *MonitoringService) getErrorRatePercentage(tenantID uint) (float64, error) {
	oneHourAgo := time.Now().Add(-time.Hour)
	
	var totalCount, errorCount int64
	
	// 总请求数
	err := ms.db.Model(&models.AuditLog{}).
		Where("tenant_id = ? AND created_at >= ?", tenantID, oneHourAgo).
		Count(&totalCount).Error
	if err != nil {
		return 0, err
	}
	
	// 错误请求数
	err = ms.db.Model(&models.AuditLog{}).
		Where("tenant_id = ? AND created_at >= ? AND status = ?", tenantID, oneHourAgo, "failed").
		Count(&errorCount).Error
	if err != nil {
		return 0, err
	}
	
	if totalCount == 0 {
		return 0, nil
	}
	
	return (float64(errorCount) / float64(totalCount)) * 100, nil
}

// getMonthlyCost 获取月度费用
func (ms *MonitoringService) getMonthlyCost(tenantID uint) (float64, error) {
	startOfMonth := time.Now().AddDate(0, 0, -time.Now().Day()+1)
	
	var totalCost float64
	err := ms.db.Model(&UsageRecord{}).
		Select("COALESCE(SUM(cost), 0)").
		Where("tenant_id = ? AND created_at >= ?", tenantID, startOfMonth).
		Scan(&totalCost).Error
	
	return totalCost, err
}

// getActiveUsers 获取活跃用户数
func (ms *MonitoringService) getActiveUsers(tenantID uint) (float64, error) {
	oneDayAgo := time.Now().Add(-24 * time.Hour)
	
	var count int64
	err := ms.db.Model(&models.AuditLog{}).
		Select("COUNT(DISTINCT user_id)").
		Where("tenant_id = ? AND created_at >= ?", tenantID, oneDayAgo).
		Scan(&count).Error
	
	return float64(count), err
}

// getStorageUsage 获取存储使用量(GB)
func (ms *MonitoringService) getStorageUsage(tenantID uint) (float64, error) {
	// 简化实现：返回模拟数据
	return 2.5, nil
}

// evaluateCondition 评估告警条件
func (ms *MonitoringService) evaluateCondition(value float64, condition AlertCondition) bool {
	switch condition.Operator {
	case "gt":
		return value > condition.Threshold
	case "gte":
		return value >= condition.Threshold
	case "lt":
		return value < condition.Threshold
	case "lte":
		return value <= condition.Threshold
	case "eq":
		return value == condition.Threshold
	case "ne":
		return value != condition.Threshold
	default:
		return false
	}
}

// createAlert 创建告警
func (ms *MonitoringService) createAlert(rule AlertRule, value, threshold float64) error {
	alert := Alert{
		TenantID:  rule.TenantID,
		RuleID:    rule.ID,
		Title:     fmt.Sprintf("Alert: %s", rule.Name),
		Message:   fmt.Sprintf("Metric %s has value %.2f, which exceeds threshold %.2f", rule.Metric, value, threshold),
		Severity:  rule.Severity,
		Status:    "active",
		Value:     value,
		Threshold: threshold,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := ms.db.Create(&alert).Error; err != nil {
		return fmt.Errorf("failed to create alert: %w", err)
	}

	ms.logger.Warn("Alert triggered",
		zap.Uint("tenant_id", rule.TenantID),
		zap.String("rule_name", rule.Name),
		zap.String("metric", rule.Metric),
		zap.Float64("value", value),
		zap.Float64("threshold", threshold))

	// 发送通知
	go ms.sendNotification(context.Background(), alert, rule)

	return nil
}

// resolveAlerts 解决告警
func (ms *MonitoringService) resolveAlerts(ruleID uint) error {
	now := time.Now()
	err := ms.db.Model(&Alert{}).
		Where("rule_id = ? AND status = ?", ruleID, "active").
		Updates(map[string]interface{}{
			"status":      "resolved",
			"resolved_at": &now,
			"updated_at":  now,
		}).Error

	return err
}

// sendNotification 发送通知
func (ms *MonitoringService) sendNotification(ctx context.Context, alert Alert, rule AlertRule) {
	// 获取通知渠道
	var channels []NotificationChannel
	err := ms.db.Where("tenant_id = ? AND enabled = ?", rule.TenantID, true).
		Find(&channels).Error
	if err != nil {
		ms.logger.Error("Failed to get notification channels", zap.Error(err))
		return
	}

	for _, channel := range channels {
		if err := ms.sendToChannel(ctx, alert, rule, channel); err != nil {
			ms.logger.Error("Failed to send notification",
				zap.Error(err),
				zap.Uint("channel_id", channel.ID),
				zap.String("channel_type", channel.Type))
		}
	}
}

// sendToChannel 发送到指定渠道
func (ms *MonitoringService) sendToChannel(ctx context.Context, alert Alert, rule AlertRule, channel NotificationChannel) error {
	switch channel.Type {
	case "email":
		return ms.sendEmailNotification(ctx, alert, rule, channel)
	case "webhook":
		return ms.sendWebhookNotification(ctx, alert, rule, channel)
	case "slack":
		return ms.sendSlackNotification(ctx, alert, rule, channel)
	default:
		return fmt.Errorf("unsupported notification channel type: %s", channel.Type)
	}
}

// sendEmailNotification 发送邮件通知
func (ms *MonitoringService) sendEmailNotification(ctx context.Context, alert Alert, rule AlertRule, channel NotificationChannel) error {
	// TODO: 实现邮件发送
	ms.logger.Info("Email notification sent",
		zap.Uint("alert_id", alert.ID),
		zap.String("title", alert.Title))
	return nil
}

// sendWebhookNotification 发送Webhook通知
func (ms *MonitoringService) sendWebhookNotification(ctx context.Context, alert Alert, rule AlertRule, channel NotificationChannel) error {
	// TODO: 实现Webhook发送
	ms.logger.Info("Webhook notification sent",
		zap.Uint("alert_id", alert.ID),
		zap.String("title", alert.Title))
	return nil
}

// sendSlackNotification 发送Slack通知
func (ms *MonitoringService) sendSlackNotification(ctx context.Context, alert Alert, rule AlertRule, channel NotificationChannel) error {
	// TODO: 实现Slack发送
	ms.logger.Info("Slack notification sent",
		zap.Uint("alert_id", alert.ID),
		zap.String("title", alert.Title))
	return nil
}

// GetTenantMetrics 获取租户指标
func (ms *MonitoringService) GetTenantMetrics(ctx context.Context, tenantID uint) (map[string]interface{}, error) {
	metrics := make(map[string]interface{})

	// 获取各种指标
	apiRequests, _ := ms.getAPIRequestsPerMinute(tenantID)
	quotaUsage, _ := ms.getQuotaUsagePercentage(tenantID)
	errorRate, _ := ms.getErrorRatePercentage(tenantID)
	monthlyCost, _ := ms.getMonthlyCost(tenantID)
	activeUsers, _ := ms.getActiveUsers(tenantID)
	storageUsage, _ := ms.getStorageUsage(tenantID)

	metrics["api_requests_per_minute"] = apiRequests
	metrics["quota_usage_percentage"] = quotaUsage
	metrics["error_rate_percentage"] = errorRate
	metrics["monthly_cost"] = monthlyCost
	metrics["active_users"] = activeUsers
	metrics["storage_usage_gb"] = storageUsage

	// 获取活跃告警数量
	var activeAlerts int64
	ms.db.Model(&Alert{}).
		Where("tenant_id = ? AND status = ?", tenantID, "active").
		Count(&activeAlerts)
	metrics["active_alerts"] = activeAlerts

	return metrics, nil
}

// GetAlerts 获取告警列表
func (ms *MonitoringService) GetAlerts(ctx context.Context, tenantID uint, status string, limit int) ([]Alert, error) {
	var alerts []Alert
	query := ms.db.Where("tenant_id = ?", tenantID)
	
	if status != "" {
		query = query.Where("status = ?", status)
	}
	
	err := query.Order("created_at DESC").Limit(limit).Find(&alerts).Error
	return alerts, err
}

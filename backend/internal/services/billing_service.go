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

// BillingService 计费服务
type BillingService struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewBillingService 创建计费服务
func NewBillingService(db *gorm.DB, logger *zap.Logger) *BillingService {
	return &BillingService{
		db:     db,
		logger: logger,
	}
}

// BillingPlan 计费计划
type BillingPlan struct {
	ID          uint                   `json:"id" gorm:"primaryKey"`
	TenantID    uint                   `json:"tenant_id" gorm:"index"`
	Name        string                 `json:"name" gorm:"size:100;not null"`
	Description string                 `json:"description" gorm:"size:500"`
	Type        string                 `json:"type" gorm:"size:50;not null"` // free, pay_as_you_go, subscription
	Currency    string                 `json:"currency" gorm:"size:10;default:'USD'"`
	Pricing     string                 `json:"pricing" gorm:"type:text"` // JSON格式的定价规则
	Quotas      string                 `json:"quotas" gorm:"type:text"`  // JSON格式的配额设置
	Features    string                 `json:"features" gorm:"type:text"` // JSON格式的功能列表
	Active      bool                   `json:"active" gorm:"default:true"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// UsageRecord 使用量记录
type UsageRecord struct {
	ID          uint                   `json:"id" gorm:"primaryKey"`
	TenantID    uint                   `json:"tenant_id" gorm:"index"`
	UserID      uint                   `json:"user_id" gorm:"index"`
	WorkspaceID *uint                  `json:"workspace_id" gorm:"index"`
	Resource    string                 `json:"resource" gorm:"size:50;not null;index"`
	Action      string                 `json:"action" gorm:"size:50;not null"`
	ResourceID  string                 `json:"resource_id" gorm:"size:100"`
	Quantity    int64                  `json:"quantity" gorm:"not null"`
	Unit        string                 `json:"unit" gorm:"size:20;not null"`
	Cost        float64                `json:"cost" gorm:"type:decimal(10,4)"`
	Currency    string                 `json:"currency" gorm:"size:10;default:'USD'"`
	Metadata    string                 `json:"metadata" gorm:"type:text"`
	BilledAt    *time.Time             `json:"billed_at"`
	CreatedAt   time.Time              `json:"created_at" gorm:"index"`
}

// BillingInvoice 账单
type BillingInvoice struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	TenantID    uint      `json:"tenant_id" gorm:"index"`
	InvoiceNo   string    `json:"invoice_no" gorm:"size:50;unique;not null"`
	Period      string    `json:"period" gorm:"size:20;not null"` // YYYY-MM
	Status      string    `json:"status" gorm:"size:20;default:'pending'"` // pending, paid, overdue, cancelled
	TotalAmount float64   `json:"total_amount" gorm:"type:decimal(10,2)"`
	Currency    string    `json:"currency" gorm:"size:10;default:'USD'"`
	DueDate     time.Time `json:"due_date"`
	PaidAt      *time.Time `json:"paid_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// BillingInvoiceItem 账单项目
type BillingInvoiceItem struct {
	ID          uint    `json:"id" gorm:"primaryKey"`
	InvoiceID   uint    `json:"invoice_id" gorm:"index"`
	Resource    string  `json:"resource" gorm:"size:50;not null"`
	Description string  `json:"description" gorm:"size:200"`
	Quantity    int64   `json:"quantity"`
	Unit        string  `json:"unit" gorm:"size:20"`
	UnitPrice   float64 `json:"unit_price" gorm:"type:decimal(10,4)"`
	Amount      float64 `json:"amount" gorm:"type:decimal(10,2)"`
	Currency    string  `json:"currency" gorm:"size:10;default:'USD'"`
}

// PricingRule 定价规则
type PricingRule struct {
	Resource    string             `json:"resource"`
	Action      string             `json:"action"`
	Unit        string             `json:"unit"`
	Tiers       []PricingTier      `json:"tiers"`
	FreeQuota   int64              `json:"free_quota"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// PricingTier 定价层级
type PricingTier struct {
	From      int64   `json:"from"`      // 起始数量
	To        int64   `json:"to"`        // 结束数量，-1表示无限制
	UnitPrice float64 `json:"unit_price"` // 单价
}

// QuotaSettings 配额设置
type QuotaSettings struct {
	Resource string `json:"resource"`
	Action   string `json:"action"`
	Limit    int64  `json:"limit"`
	Period   string `json:"period"` // daily, monthly, yearly
	Unit     string `json:"unit"`
}

// RecordUsage 记录使用量
func (bs *BillingService) RecordUsage(ctx context.Context, req UsageRecordRequest) error {
	// 获取租户的计费计划
	plan, err := bs.getTenantBillingPlan(req.TenantID)
	if err != nil {
		return fmt.Errorf("failed to get billing plan: %w", err)
	}

	// 计算费用
	cost, err := bs.calculateCost(plan, req)
	if err != nil {
		return fmt.Errorf("failed to calculate cost: %w", err)
	}

	// 序列化元数据
	metadataJSON := ""
	if req.Metadata != nil {
		metadataBytes, _ := json.Marshal(req.Metadata)
		metadataJSON = string(metadataBytes)
	}

	// 创建使用量记录
	record := UsageRecord{
		TenantID:    req.TenantID,
		UserID:      req.UserID,
		WorkspaceID: req.WorkspaceID,
		Resource:    req.Resource,
		Action:      req.Action,
		ResourceID:  req.ResourceID,
		Quantity:    req.Quantity,
		Unit:        req.Unit,
		Cost:        cost,
		Currency:    plan.Currency,
		Metadata:    metadataJSON,
		CreatedAt:   time.Now(),
	}

	if err := bs.db.Create(&record).Error; err != nil {
		return fmt.Errorf("failed to create usage record: %w", err)
	}

	bs.logger.Info("Usage recorded for billing",
		zap.Uint("tenant_id", req.TenantID),
		zap.Uint("user_id", req.UserID),
		zap.String("resource", req.Resource),
		zap.String("action", req.Action),
		zap.Int64("quantity", req.Quantity),
		zap.Float64("cost", cost))

	return nil
}

// getTenantBillingPlan 获取租户计费计划
func (bs *BillingService) getTenantBillingPlan(tenantID uint) (*BillingPlan, error) {
	var plan BillingPlan
	err := bs.db.Where("tenant_id = ? AND active = ?", tenantID, true).
		First(&plan).Error
	
	if err == gorm.ErrRecordNotFound {
		// 如果没有找到计费计划，返回默认的免费计划
		return bs.getDefaultFreePlan(tenantID), nil
	}
	
	if err != nil {
		return nil, err
	}
	
	return &plan, nil
}

// getDefaultFreePlan 获取默认免费计划
func (bs *BillingService) getDefaultFreePlan(tenantID uint) *BillingPlan {
	return &BillingPlan{
		TenantID:    tenantID,
		Name:        "Free Plan",
		Type:        "free",
		Currency:    "USD",
		Pricing:     `[]`, // 免费计划无定价
		Quotas:      `[{"resource":"api","action":"call","limit":1000,"period":"monthly","unit":"requests"}]`,
		Features:    `["basic_features"]`,
		Active:      true,
	}
}

// calculateCost 计算费用
func (bs *BillingService) calculateCost(plan *BillingPlan, req UsageRecordRequest) (float64, error) {
	if plan.Type == "free" {
		return 0.0, nil
	}

	// 解析定价规则
	var pricingRules []PricingRule
	if err := json.Unmarshal([]byte(plan.Pricing), &pricingRules); err != nil {
		return 0.0, fmt.Errorf("failed to parse pricing rules: %w", err)
	}

	// 查找适用的定价规则
	var applicableRule *PricingRule
	for _, rule := range pricingRules {
		if rule.Resource == req.Resource && rule.Action == req.Action {
			applicableRule = &rule
			break
		}
	}

	if applicableRule == nil {
		// 没有找到定价规则，免费
		return 0.0, nil
	}

	// 检查免费配额
	if req.Quantity <= applicableRule.FreeQuota {
		return 0.0, nil
	}

	// 计算需要计费的数量
	billableQuantity := req.Quantity - applicableRule.FreeQuota

	// 根据层级定价计算费用
	return bs.calculateTieredCost(applicableRule.Tiers, billableQuantity), nil
}

// calculateTieredCost 计算层级定价费用
func (bs *BillingService) calculateTieredCost(tiers []PricingTier, quantity int64) float64 {
	if len(tiers) == 0 {
		return 0.0
	}

	totalCost := 0.0
	remainingQuantity := quantity

	for _, tier := range tiers {
		if remainingQuantity <= 0 {
			break
		}

		// 计算当前层级的数量
		tierQuantity := int64(0)
		if tier.To == -1 || remainingQuantity <= (tier.To-tier.From+1) {
			tierQuantity = remainingQuantity
		} else {
			tierQuantity = tier.To - tier.From + 1
		}

		// 计算当前层级的费用
		tierCost := float64(tierQuantity) * tier.UnitPrice
		totalCost += tierCost

		remainingQuantity -= tierQuantity
	}

	return totalCost
}

// GenerateInvoice 生成账单
func (bs *BillingService) GenerateInvoice(ctx context.Context, tenantID uint, period string) (*BillingInvoice, error) {
	// 检查是否已经生成过该期间的账单
	var existingInvoice BillingInvoice
	err := bs.db.Where("tenant_id = ? AND period = ?", tenantID, period).
		First(&existingInvoice).Error
	
	if err == nil {
		return &existingInvoice, nil
	}
	
	if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to check existing invoice: %w", err)
	}

	// 获取该期间的使用量记录
	usageRecords, err := bs.getUsageRecordsForPeriod(tenantID, period)
	if err != nil {
		return nil, fmt.Errorf("failed to get usage records: %w", err)
	}

	// 计算总金额和创建账单项目
	totalAmount := 0.0
	var invoiceItems []BillingInvoiceItem

	// 按资源分组统计
	resourceUsage := make(map[string]float64)
	resourceQuantity := make(map[string]int64)
	
	for _, record := range usageRecords {
		key := fmt.Sprintf("%s:%s", record.Resource, record.Action)
		resourceUsage[key] += record.Cost
		resourceQuantity[key] += record.Quantity
		totalAmount += record.Cost
	}

	// 创建账单项目
	for key, amount := range resourceUsage {
		parts := strings.Split(key, ":")
		if len(parts) != 2 {
			continue
		}
		
		item := BillingInvoiceItem{
			Resource:    parts[0],
			Description: fmt.Sprintf("%s %s usage", parts[0], parts[1]),
			Quantity:    resourceQuantity[key],
			Unit:        "units", // 简化处理
			UnitPrice:   amount / float64(resourceQuantity[key]),
			Amount:      amount,
			Currency:    "USD",
		}
		invoiceItems = append(invoiceItems, item)
	}

	// 生成账单号
	invoiceNo := fmt.Sprintf("INV-%d-%s", tenantID, period)

	// 创建账单
	invoice := BillingInvoice{
		TenantID:    tenantID,
		InvoiceNo:   invoiceNo,
		Period:      period,
		Status:      "pending",
		TotalAmount: totalAmount,
		Currency:    "USD",
		DueDate:     time.Now().AddDate(0, 1, 0), // 30天后到期
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 开始事务
	tx := bs.db.Begin()
	
	// 创建账单
	if err := tx.Create(&invoice).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create invoice: %w", err)
	}

	// 创建账单项目
	for i := range invoiceItems {
		invoiceItems[i].InvoiceID = invoice.ID
		if err := tx.Create(&invoiceItems[i]).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to create invoice item: %w", err)
		}
	}

	// 标记使用量记录为已计费
	if err := tx.Model(&UsageRecord{}).
		Where("tenant_id = ? AND created_at >= ? AND created_at < ? AND billed_at IS NULL", 
			tenantID, bs.getPeriodStart(period), bs.getPeriodEnd(period)).
		Update("billed_at", time.Now()).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to mark usage records as billed: %w", err)
	}

	tx.Commit()

	bs.logger.Info("Invoice generated",
		zap.Uint("tenant_id", tenantID),
		zap.String("period", period),
		zap.String("invoice_no", invoiceNo),
		zap.Float64("total_amount", totalAmount))

	return &invoice, nil
}

// getUsageRecordsForPeriod 获取指定期间的使用量记录
func (bs *BillingService) getUsageRecordsForPeriod(tenantID uint, period string) ([]UsageRecord, error) {
	var records []UsageRecord
	
	startTime := bs.getPeriodStart(period)
	endTime := bs.getPeriodEnd(period)
	
	err := bs.db.Where("tenant_id = ? AND created_at >= ? AND created_at < ? AND billed_at IS NULL",
		tenantID, startTime, endTime).
		Find(&records).Error
	
	return records, err
}

// getPeriodStart 获取期间开始时间
func (bs *BillingService) getPeriodStart(period string) time.Time {
	// 期间格式: YYYY-MM
	t, _ := time.Parse("2006-01", period)
	return t
}

// getPeriodEnd 获取期间结束时间
func (bs *BillingService) getPeriodEnd(period string) time.Time {
	start := bs.getPeriodStart(period)
	return start.AddDate(0, 1, 0)
}

// GetTenantUsageSummary 获取租户使用量摘要
func (bs *BillingService) GetTenantUsageSummary(ctx context.Context, tenantID uint, period string) (map[string]interface{}, error) {
	startTime := bs.getPeriodStart(period)
	endTime := bs.getPeriodEnd(period)

	// 查询使用量统计
	var results []struct {
		Resource string  `json:"resource"`
		Action   string  `json:"action"`
		Quantity int64   `json:"quantity"`
		Cost     float64 `json:"cost"`
	}

	err := bs.db.Model(&UsageRecord{}).
		Select("resource, action, SUM(quantity) as quantity, SUM(cost) as cost").
		Where("tenant_id = ? AND created_at >= ? AND created_at < ?", tenantID, startTime, endTime).
		Group("resource, action").
		Find(&results).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get usage summary: %w", err)
	}

	// 构建摘要
	summary := map[string]interface{}{
		"tenant_id": tenantID,
		"period":    period,
		"usage":     results,
		"total_cost": func() float64 {
			total := 0.0
			for _, r := range results {
				total += r.Cost
			}
			return total
		}(),
	}

	return summary, nil
}

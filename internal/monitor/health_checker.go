package monitor

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/awd-platform/awd-arena/internal/model"

	"gorm.io/gorm"
)

// ServiceHealthChecker 服务健康检查器
type ServiceHealthChecker struct {
	db        *gorm.DB
	config    *model.HealthCheckConfig
	alertChan chan AlertEvent
	checkers  map[uint]*ServiceChecker
	mu        sync.RWMutex
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
}

// AlertEvent 告警事件
type AlertEvent struct {
	ServiceID   uint      `json:"service_id"`
	ServiceName string    `json:"service_name"`
	Status      string    `json:"status"`
	OldStatus   string    `json:"old_status"`
	CheckedAt   time.Time `json:"checked_at"`
	ErrorMsg    string    `json:"error_msg"`
}

// ServiceChecker 单个服务的检查器
type ServiceChecker struct {
	service          *model.TargetService
	config           *model.HealthCheckConfig
	db               *gorm.DB
	alertChan        chan AlertEvent
	stopChan         chan struct{}
	consecutiveFails int
	lastStatus       string
	lastNotified     bool
}

// NewServiceHealthChecker 创建健康检查器
func NewServiceHealthChecker(db *gorm.DB, config *model.HealthCheckConfig, alertChan chan AlertEvent) *ServiceHealthChecker {
	if config == nil {
		config = model.DefaultHealthCheckConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &ServiceHealthChecker{
		db:        db,
		config:    config,
		alertChan: alertChan,
		checkers:  make(map[uint]*ServiceChecker),
		ctx:       ctx,
		cancel:    cancel,
	}
}

// Start 启动健康检查
func (hc *ServiceHealthChecker) Start() error {
	// 加载所有启用的服务
	var services []model.TargetService
	if err := hc.db.Where("enabled = ?", true).Find(&services).Error; err != nil {
		return fmt.Errorf("failed to load services: %w", err)
	}

	for _, service := range services {
		hc.AddService(&service)
	}

	return nil
}

// Stop 停止健康检查
func (hc *ServiceHealthChecker) Stop() {
	hc.cancel()
	hc.wg.Wait()

	hc.mu.Lock()
	defer hc.mu.Unlock()

	for _, checker := range hc.checkers {
		close(checker.stopChan)
	}
	hc.checkers = make(map[uint]*ServiceChecker)
}

// AddService 添加服务监控
func (hc *ServiceHealthChecker) AddService(service *model.TargetService) {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	if _, exists := hc.checkers[service.ID]; exists {
		return
	}

	checker := &ServiceChecker{
		service:    service,
		config:     hc.config,
		db:         hc.db,
		alertChan:  hc.alertChan,
		stopChan:   make(chan struct{}),
		lastStatus: model.HealthStatusUnknown,
	}

	hc.checkers[service.ID] = checker

	hc.wg.Add(1)
	go checker.run(hc.ctx, &hc.wg)
}

// RemoveService 移除服务监控
func (hc *ServiceHealthChecker) RemoveService(serviceID uint) {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	if checker, exists := hc.checkers[serviceID]; exists {
		close(checker.stopChan)
		delete(hc.checkers, serviceID)
	}
}

// run 运行服务检查
func (sc *ServiceChecker) run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	ticker := time.NewTicker(time.Duration(sc.config.CheckInterval) * time.Second)
	defer ticker.Stop()

	// 立即执行第一次检查
	sc.check()

	for {
		select {
		case <-ctx.Done():
			return
		case <-sc.stopChan:
			return
		case <-ticker.C:
			sc.check()
		}
	}
}

// check 执行单次健康检查
func (sc *ServiceChecker) check() {
	now := time.Now()

	result := sc.performCheck()

	// 记录检查结果
	healthRecord := &model.ServiceHealth{
		ServiceID:    sc.service.ID,
		Status:       result.Status,
		CheckedAt:    now,
		ResponseTime: result.ResponseTime,
		ErrorMsg:     result.ErrorMsg,
		CreatedAt:    now,
	}

	if err := sc.db.Create(healthRecord).Error; err != nil {
		fmt.Printf("failed to save health record: %v\n", err)
	}

	// 状态变更检测和告警
	sc.handleStatusChange(result, now)
}

// CheckResult 检查结果
type CheckResult struct {
	Status       string
	ResponseTime int64
	ErrorMsg     string
}

// performCheck 执行实际的检查
func (sc *ServiceChecker) performCheck() *CheckResult {
	var result CheckResult
	start := time.Now()

	switch sc.service.Protocol {
	case "http", "https":
		result = sc.checkHTTP()
	case "tcp":
		result = sc.checkTCP()
	default:
		result = CheckResult{
			Status:   model.HealthStatusUnknown,
			ErrorMsg: fmt.Sprintf("unsupported protocol: %s", sc.service.Protocol),
		}
	}

	result.ResponseTime = time.Since(start).Milliseconds()
	return &result
}

// checkHTTP HTTP/HTTPS 检查
func (sc *ServiceChecker) checkHTTP() CheckResult {
	client := &http.Client{
		Timeout: time.Duration(sc.config.Timeout) * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	url := fmt.Sprintf("%s://%s:%d%s",
		sc.service.Protocol,
		sc.service.Host,
		sc.service.Port,
		sc.service.Path,
	)

	resp, err := client.Get(url)
	if err != nil {
		return CheckResult{
			Status:   model.HealthStatusUnhealthy,
			ErrorMsg: err.Error(),
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 500 {
		return CheckResult{
			Status: model.HealthStatusHealthy,
		}
	}

	return CheckResult{
		Status:   model.HealthStatusUnhealthy,
		ErrorMsg: fmt.Sprintf("HTTP status: %d", resp.StatusCode),
	}
}

// checkTCP TCP 检查
func (sc *ServiceChecker) checkTCP() CheckResult {
	address := fmt.Sprintf("[%s]:%d", sc.service.Host, sc.service.Port)

	conn, err := net.DialTimeout("tcp", address,
		time.Duration(sc.config.Timeout)*time.Second)
	if err != nil {
		return CheckResult{
			Status:   model.HealthStatusUnhealthy,
			ErrorMsg: err.Error(),
		}
	}
	defer conn.Close()

	return CheckResult{
		Status: model.HealthStatusHealthy,
	}
}

// handleStatusChange 处理状态变更
func (sc *ServiceChecker) handleStatusChange(result *CheckResult, checkedAt time.Time) {
	oldStatus := sc.lastStatus
	newStatus := result.Status

	// 更新失败计数
	if newStatus == model.HealthStatusUnhealthy {
		sc.consecutiveFails++
	} else {
		sc.consecutiveFails = 0
	}

	// 检查是否需要发送告警
	shouldAlert := false
	alertStatus := newStatus

	if newStatus == model.HealthStatusUnhealthy &&
		sc.consecutiveFails >= sc.config.FailureCount &&
		!sc.lastNotified {
		// 服务宕机告警
		shouldAlert = true
		sc.lastNotified = true
	} else if oldStatus == model.HealthStatusUnhealthy &&
		newStatus == model.HealthStatusHealthy &&
		sc.config.RecoveryNotify &&
		sc.lastNotified {
		// 服务恢复通知
		shouldAlert = true
		alertStatus = "recovered"
		sc.lastNotified = false
	}

	// 发送告警
	if shouldAlert && sc.alertChan != nil {
		select {
		case sc.alertChan <- AlertEvent{
			ServiceID:   sc.service.ID,
			ServiceName: sc.service.Name,
			Status:      alertStatus,
			OldStatus:   oldStatus,
			CheckedAt:   checkedAt,
			ErrorMsg:    result.ErrorMsg,
		}:
		default:
			// 告警通道满，丢弃
		}
	}

	sc.lastStatus = newStatus
}

// GetServiceStatus 获取服务当前状态
func (hc *ServiceHealthChecker) GetServiceStatus(serviceID uint) (string, error) {
	hc.mu.RLock()
	defer hc.mu.RUnlock()

	checker, exists := hc.checkers[serviceID]
	if !exists {
		return "", fmt.Errorf("service not found")
	}

	return checker.lastStatus, nil
}

// GetAllStatuses 获取所有服务状态
func (hc *ServiceHealthChecker) GetAllStatuses() map[uint]string {
	hc.mu.RLock()
	defer hc.mu.RUnlock()

	statuses := make(map[uint]string)
	for id, checker := range hc.checkers {
		statuses[id] = checker.lastStatus
	}

	return statuses
}

// GetHealthStats 获取服务健康统计
func (hc *ServiceHealthChecker) GetHealthStats(serviceID uint, since time.Time) (*model.ServiceHealthSummary, error) {
	var service model.TargetService
	if err := hc.db.First(&service, serviceID).Error; err != nil {
		return nil, err
	}

	var totalChecks, healthyChecks int64
	var avgResponseTime float64

	// 统计检查次数
	if err := hc.db.Model(&model.ServiceHealth{}).
		Where("service_id = ? AND checked_at >= ?", serviceID, since).
		Count(&totalChecks).Error; err != nil {
		return nil, err
	}

	// 统计健康次数
	if err := hc.db.Model(&model.ServiceHealth{}).
		Where("service_id = ? AND checked_at >= ? AND status = ?",
			serviceID, since, model.HealthStatusHealthy).
		Count(&healthyChecks).Error; err != nil {
		return nil, err
	}

	// 计算平均响应时间
	type AvgResult struct {
		Avg float64
	}
	var avgResult AvgResult
	if err := hc.db.Model(&model.ServiceHealth{}).
		Select("AVG(response_time) as avg").
		Where("service_id = ? AND checked_at >= ?", serviceID, since).
		Scan(&avgResult).Error; err != nil {
		return nil, err
	}
	avgResponseTime = avgResult.Avg

	// 获取最新状态
	var latestHealth model.ServiceHealth
	if err := hc.db.Where("service_id = ?", serviceID).
		Order("checked_at DESC").
		First(&latestHealth).Error; err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	var uptimePercent float64
	if totalChecks > 0 {
		uptimePercent = float64(healthyChecks) / float64(totalChecks) * 100
	}

	return &model.ServiceHealthSummary{
		ServiceID:       serviceID,
		ServiceName:     service.Name,
		CurrentStatus:   latestHealth.Status,
		LastCheckedAt:   latestHealth.CheckedAt,
		UptimePercent:   uptimePercent,
		TotalChecks:     totalChecks,
		HealthyChecks:   healthyChecks,
		AvgResponseTime: int64(avgResponseTime),
	}, nil
}

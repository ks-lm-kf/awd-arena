package monitor

import (
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/awd-platform/awd-arena/internal/model"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}

	// 自动迁移
	err = db.AutoMigrate(&model.TargetService{}, &model.ServiceHealth{})
	if err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	return db
}

func createTestService(t *testing.T, db *gorm.DB, name, protocol, host string, port int) *model.TargetService {
	service := &model.TargetService{
		Name:     name,
		Protocol: protocol,
		Host:     host,
		Port:     port,
		Path:     "/",
		Enabled:  true,
	}

	if err := db.Create(service).Error; err != nil {
		t.Fatalf("failed to create service: %v", err)
	}

	return service
}

// 辅助函数：从URL中提取端口号
func getPortFromURL(url string) int {
	// URL格式为 http://127.0.0.1:PORT
	parts := strings.Split(url, ":")
	if len(parts) >= 3 {
		portStr := strings.Split(parts[2], "/")[0]
		if p, err := strconv.Atoi(portStr); err == nil {
			return p
		}
	}
	return 8080
}

func TestNewServiceHealthChecker(t *testing.T) {
	db := setupTestDB(t)

	t.Run("with default config", func(t *testing.T) {
		hc := NewServiceHealthChecker(db, nil, nil)
		if hc == nil {
			t.Fatal("expected health checker to be created")
		}
		if hc.config.CheckInterval != 30 {
			t.Errorf("expected default check interval 30, got %d", hc.config.CheckInterval)
		}
	})

	t.Run("with custom config", func(t *testing.T) {
		config := &model.HealthCheckConfig{
			CheckInterval: 5,
			Timeout:       3,
		}
		hc := NewServiceHealthChecker(db, config, nil)
		if hc.config.CheckInterval != 5 {
			t.Errorf("expected check interval 5, got %d", hc.config.CheckInterval)
		}
	})
}

func TestServiceHealthChecker_StartStop(t *testing.T) {
	db := setupTestDB(t)

	// 创建测试服务
	createTestService(t, db, "test-service", "http", "localhost", 8080)

	hc := NewServiceHealthChecker(db, nil, nil)

	if err := hc.Start(); err != nil {
		t.Fatalf("failed to start health checker: %v", err)
	}

	// 等待一小段时间
	time.Sleep(100 * time.Millisecond)

	hc.Stop()

	// 验证可以安全停止
	time.Sleep(100 * time.Millisecond)
}

func TestServiceHealthChecker_AddRemoveService(t *testing.T) {
	db := setupTestDB(t)
	hc := NewServiceHealthChecker(db, nil, nil)

	service := &model.TargetService{
		Name:     "test",
		Protocol: "http",
		Host:     "localhost",
		Port:     8080,
		Enabled:  true,
	}
	db.Create(service)

	hc.AddService(service)

	if len(hc.checkers) != 1 {
		t.Errorf("expected 1 checker, got %d", len(hc.checkers))
	}

	// 添加相同的服务应该不会增加计数
	hc.AddService(service)
	if len(hc.checkers) != 1 {
		t.Errorf("expected 1 checker after duplicate add, got %d", len(hc.checkers))
	}

	hc.RemoveService(service.ID)
	if len(hc.checkers) != 0 {
		t.Errorf("expected 0 checkers after remove, got %d", len(hc.checkers))
	}
}

func TestServiceChecker_CheckHTTP(t *testing.T) {
	// 启动测试HTTP服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	db := setupTestDB(t)

	// 创建检查器
	sc := &ServiceChecker{
		service: &model.TargetService{
			Protocol: "http",
			Host:     "127.0.0.1",
			Port:     getPortFromURL(server.URL),
			Path:     "/",
		},
		config: model.DefaultHealthCheckConfig(),
		db:     db,
	}

	result := sc.performCheck()

	if result.Status != model.HealthStatusHealthy {
		t.Errorf("expected healthy status, got %s, error: %s", result.Status, result.ErrorMsg)
	}

	if result.ResponseTime < 0 {
		t.Errorf("expected non-negative response time, got %d", result.ResponseTime)
	}
}

func TestServiceChecker_CheckHTTPFailure(t *testing.T) {
	db := setupTestDB(t)

	sc := &ServiceChecker{
		service: &model.TargetService{
			Protocol: "http",
			Host:     "localhost",
			Port:     59999, // 不存在的端口
			Path:     "/",
		},
		config: model.DefaultHealthCheckConfig(),
		db:     db,
	}

	result := sc.performCheck()

	if result.Status != model.HealthStatusUnhealthy {
		t.Errorf("expected unhealthy status, got %s", result.Status)
	}

	if result.ErrorMsg == "" {
		t.Error("expected error message for failed check")
	}
}

func TestServiceChecker_CheckTCP(t *testing.T) {
	// 启动测试TCP服务器
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start tcp server: %v", err)
	}
	defer listener.Close()

	port := listener.Addr().(*net.TCPAddr).Port

	db := setupTestDB(t)

	sc := &ServiceChecker{
		service: &model.TargetService{
			Protocol: "tcp",
			Host:     "127.0.0.1",
			Port:     port,
		},
		config: model.DefaultHealthCheckConfig(),
		db:     db,
	}

	result := sc.performCheck()

	if result.Status != model.HealthStatusHealthy {
		t.Errorf("expected healthy status, got %s", result.Status)
	}
}

func TestServiceChecker_CheckTCPFailure(t *testing.T) {
	db := setupTestDB(t)

	sc := &ServiceChecker{
		service: &model.TargetService{
			Protocol: "tcp",
			Host:     "localhost",
			Port:     59999, // 不存在的端口
		},
		config: model.DefaultHealthCheckConfig(),
		db:     db,
	}

	result := sc.performCheck()

	if result.Status != model.HealthStatusUnhealthy {
		t.Errorf("expected unhealthy status, got %s", result.Status)
	}
}

func TestServiceChecker_CheckUnsupportedProtocol(t *testing.T) {
	db := setupTestDB(t)

	sc := &ServiceChecker{
		service: &model.TargetService{
			Protocol: "udp",
			Host:     "localhost",
			Port:     53,
		},
		config: model.DefaultHealthCheckConfig(),
		db:     db,
	}

	result := sc.performCheck()

	if result.Status != model.HealthStatusUnknown {
		t.Errorf("expected unknown status, got %s", result.Status)
	}
}

func TestServiceChecker_HandleStatusChange(t *testing.T) {
	db := setupTestDB(t)
	alertChan := make(chan AlertEvent, 10)

	sc := &ServiceChecker{
		service: &model.TargetService{
			Model:    gorm.Model{ID: 1},
			Name:     "test-service",
			Protocol: "http",
			Host:     "localhost",
			Port:     8080,
		},
		config:     model.DefaultHealthCheckConfig(),
		db:         db,
		alertChan:  alertChan,
		lastStatus: model.HealthStatusHealthy,
	}

	// 测试状态变化到不健康，达到告警阈值
	for i := 0; i < sc.config.FailureCount; i++ {
		sc.handleStatusChange(&CheckResult{
			Status:   model.HealthStatusUnhealthy,
			ErrorMsg: "connection refused",
		}, time.Now())
	}

	// 应该收到告警
	select {
	case alert := <-alertChan:
		if alert.Status != model.HealthStatusUnhealthy {
			t.Errorf("expected unhealthy alert, got %s", alert.Status)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("expected alert event")
	}

	// 测试恢复告警
	sc.handleStatusChange(&CheckResult{
		Status: model.HealthStatusHealthy,
	}, time.Now())

	select {
	case alert := <-alertChan:
		if alert.Status != "recovered" {
			t.Errorf("expected recovered alert, got %s", alert.Status)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("expected recovery alert")
	}
}

func TestServiceHealthChecker_GetServiceStatus(t *testing.T) {
	db := setupTestDB(t)
	hc := NewServiceHealthChecker(db, nil, nil)

	service := &model.TargetService{
		Name:     "test",
		Protocol: "http",
		Host:     "localhost",
		Port:     8080,
		Enabled:  true,
	}
	db.Create(service)

	// 不存在的服务
	_, err := hc.GetServiceStatus(999)
	if err == nil {
		t.Error("expected error for non-existent service")
	}

	hc.AddService(service)

	// 存在的服务应该能获取状态
	status, err := hc.GetServiceStatus(service.ID)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if status != model.HealthStatusUnknown {
		t.Errorf("expected initial status unknown, got %s", status)
	}
}

func TestServiceHealthChecker_GetAllStatuses(t *testing.T) {
	db := setupTestDB(t)
	hc := NewServiceHealthChecker(db, nil, nil)

	service1 := &model.TargetService{
		Name:     "test1",
		Protocol: "http",
		Host:     "localhost",
		Port:     8080,
		Enabled:  true,
	}
	service2 := &model.TargetService{
		Name:     "test2",
		Protocol: "http",
		Host:     "localhost",
		Port:     8081,
		Enabled:  true,
	}
	db.Create(service1)
	db.Create(service2)

	statuses := hc.GetAllStatuses()
	if len(statuses) != 0 {
		t.Errorf("expected 0 statuses, got %d", len(statuses))
	}

	hc.AddService(service1)
	hc.AddService(service2)

	statuses = hc.GetAllStatuses()
	if len(statuses) != 2 {
		t.Errorf("expected 2 statuses, got %d", len(statuses))
	}
}

func TestServiceHealthChecker_GetHealthStats(t *testing.T) {
	db := setupTestDB(t)
	hc := NewServiceHealthChecker(db, nil, nil)

	service := &model.TargetService{
		Name:     "test",
		Protocol: "http",
		Host:     "localhost",
		Port:     8080,
		Enabled:  true,
	}
	db.Create(service)

	// 插入一些健康记录
	now := time.Now()
	for i := 0; i < 10; i++ {
		status := model.HealthStatusHealthy
		if i%3 == 0 {
			status = model.HealthStatusUnhealthy
		}
		health := &model.ServiceHealth{
			ServiceID:    service.ID,
			Status:       status,
			CheckedAt:    now.Add(-time.Duration(i) * time.Minute),
			ResponseTime: 100,
			CreatedAt:    now.Add(-time.Duration(i) * time.Minute),
		}
		db.Create(health)
	}

	stats, err := hc.GetHealthStats(service.ID, now.Add(-15*time.Minute))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if stats.ServiceID != service.ID {
		t.Errorf("expected service id %d, got %d", service.ID, stats.ServiceID)
	}

	if stats.TotalChecks != 10 {
		t.Errorf("expected 10 total checks, got %d", stats.TotalChecks)
	}

	// 10次中有4次不健康(i=0,3,6,9)，6次健康
	if stats.HealthyChecks != 6 {
		t.Errorf("expected 6 healthy checks, got %d", stats.HealthyChecks)
	}

	// 可用率应该是 60%
	expectedUptime := 60.0
	if stats.UptimePercent < expectedUptime-0.1 || stats.UptimePercent > expectedUptime+0.1 {
		t.Errorf("expected uptime around %.1f, got %.1f", expectedUptime, stats.UptimePercent)
	}
}

func TestServiceHealthChecker_SaveHealthRecord(t *testing.T) {
	db := setupTestDB(t)

	// 创建服务
	service := createTestService(t, db, "test", "http", "localhost", 8080)

	// 启动测试服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	hc := NewServiceHealthChecker(db, &model.HealthCheckConfig{
		CheckInterval: 1,
		Timeout:       1,
		FailureCount:  1,
	}, nil)

	hc.AddService(&model.TargetService{
		Model:    gorm.Model{ID: service.ID},
		Name:     service.Name,
		Protocol: "http",
		Host:     "127.0.0.1",
		Port:     getPortFromURL(server.URL),
		Path:     "/",
		Enabled:  true,
	})

	// 等待一次检查
	time.Sleep(200 * time.Millisecond)
	hc.Stop()

	// 验证记录已保存
	var count int64
	db.Model(&model.ServiceHealth{}).Where("service_id = ?", service.ID).Count(&count)
	if count == 0 {
		t.Error("expected health records to be saved")
	}
}

func TestServiceHealthChecker_ConcurrentOperations(t *testing.T) {
	db := setupTestDB(t)
	hc := NewServiceHealthChecker(db, &model.HealthCheckConfig{
		CheckInterval: 10,
		Timeout:       1,
	}, nil)

	var wg sync.WaitGroup

	// 并发添加服务
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			service := &model.TargetService{
				Name:     string(rune('A' + id)),
				Protocol: "http",
				Host:     "localhost",
				Port:     8080 + id,
				Enabled:  true,
			}
			db.Create(service)
			hc.AddService(service)
		}(i)
	}

	// 并发获取状态
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			hc.GetAllStatuses()
		}()
	}

	wg.Wait()

	// 验证没有竞态条件
	hc.Stop()
}

func TestServiceChecker_Timeout(t *testing.T) {
	db := setupTestDB(t)

	// 启动一个慢速服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	sc := &ServiceChecker{
		service: &model.TargetService{
			Protocol: "http",
			Host:     "127.0.0.1",
			Port:     getPortFromURL(server.URL),
			Path:     "/",
		},
		config: &model.HealthCheckConfig{
			Timeout: 1, // 1秒超时
		},
		db: db,
	}

	start := time.Now()
	result := sc.performCheck()
	elapsed := time.Since(start)

	// 应该在1秒超时，加上一些余量
	if elapsed > 1500*time.Millisecond {
		t.Errorf("check took too long: %v", elapsed)
	}

	if result.Status != model.HealthStatusUnhealthy {
		t.Errorf("expected unhealthy due to timeout, got %s", result.Status)
	}
}

func TestDefaultHealthCheckConfig(t *testing.T) {
	config := model.DefaultHealthCheckConfig()

	if config.CheckInterval != 30 {
		t.Errorf("expected default check interval 30, got %d", config.CheckInterval)
	}

	if config.Timeout != 10 {
		t.Errorf("expected default timeout 10, got %d", config.Timeout)
	}

	if config.MaxRetries != 3 {
		t.Errorf("expected default max retries 3, got %d", config.MaxRetries)
	}

	if config.FailureCount != 3 {
		t.Errorf("expected default failure count 3, got %d", config.FailureCount)
	}

	if !config.RecoveryNotify {
		t.Error("expected recovery notify to be true by default")
	}
}

package handler

import (
"github.com/awd-platform/awd-arena/internal/model"
"github.com/gofiber/fiber/v3"
"gorm.io/gorm"
"strconv"
)

// AuditHandler handles audit-related requests
var AuditHandler *auditHandler

func init() {
AuditHandler = &auditHandler{}
}

type auditHandler struct{}

// GetAuditLogs returns all audit logs with filtering
// GET /api/audit/logs
func (h *auditHandler) GetAuditLogs(c fiber.Ctx) error {
db := c.Locals("db").(*gorm.DB)

page, _ := strconv.Atoi(c.Query("page", "1"))
pageSize, _ := strconv.Atoi(c.Query("page_size", "50"))
action := c.Query("action")
resourceType := c.Query("resource_type")
userID := c.Query("user_id")
startDate := c.Query("start_date")
endDate := c.Query("end_date")

if page < 1 {
page = 1
}
if pageSize < 1 || pageSize > 100 {
pageSize = 50
}

var logs []model.AdminLog
var total int64

query := db.Model(&model.AdminLog{})

if action != "" {
query = query.Where("action = ?", action)
}
if resourceType != "" {
query = query.Where("resource_type = ?", resourceType)
}
if userID != "" {
query = query.Where("user_id = ?", userID)
}
if startDate != "" {
query = query.Where("created_at >= ?", startDate)
}
if endDate != "" {
query = query.Where("created_at <= ?", endDate)
}

query.Count(&total)

offset := (page - 1) * pageSize
if err := query.Order("created_at DESC").Limit(pageSize).Offset(offset).Find(&logs).Error; err != nil {
return c.Status(500).JSON(fiber.Map{"code": 500, "message": err.Error()})
}

return c.JSON(fiber.Map{
"code":    0,
"message": "ok",
"data": fiber.Map{
"items":     logs,
"total":     total,
"page":      page,
"page_size": pageSize,
},
})
}

// GetAuditStats returns audit statistics
// GET /api/audit/stats
func (h *auditHandler) GetAuditStats(c fiber.Ctx) error {
db := c.Locals("db").(*gorm.DB)

var stats struct {
TotalLogs     int64 `json:"total_logs"`
LoginCount    int64 `json:"login_count"`
CreateCount   int64 `json:"create_count"`
UpdateCount   int64 `json:"update_count"`
DeleteCount   int64 `json:"delete_count"`
}

// Total logs
db.Model(&model.AdminLog{}).Count(&stats.TotalLogs)

// Count by action type
db.Model(&model.AdminLog{}).Where("action = ?", "login").Count(&stats.LoginCount)
db.Model(&model.AdminLog{}).Where("action = ?", "create").Count(&stats.CreateCount)
db.Model(&model.AdminLog{}).Where("action = ?", "update").Count(&stats.UpdateCount)
db.Model(&model.AdminLog{}).Where("action = ?", "delete").Count(&stats.DeleteCount)

return c.JSON(fiber.Map{
"code":    0,
"message": "ok",
"data":    stats,
})
}

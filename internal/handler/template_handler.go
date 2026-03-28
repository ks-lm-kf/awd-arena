package handler

import (
	"fmt"
	"strconv"
	"time"

	"github.com/awd-platform/awd-arena/internal/database"
	"github.com/awd-platform/awd-arena/internal/model"
	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

// TemplateHandler 题库模板处理器
type TemplateHandler struct {
	db *gorm.DB
}

// NewTemplateHandler 创建模板处理器
func NewTemplateHandler() *TemplateHandler {
	return &TemplateHandler{db: database.GetDB()}
}

// List 获取模板列表
// GET /api/v1/templates
func (h *TemplateHandler) List(c fiber.Ctx) error {
	query := struct {
		Category   string `form:"category"`
		Difficulty string `form:"difficulty"`
		Page       int    `form:"page,default=1"`
		PageSize   int    `form:"page_size,default=20"`
		Status     string `form:"status"`
	}{
		Page:     1,
		PageSize: 20,
	}
	
	if err := c.Bind().Query(&query); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"code":    400,
			"message": "invalid query parameters",
		})
	}
	
	db := h.db.Model(&model.ChallengeTemplate{})
	
	// 过滤条件
	if query.Category != "" {
		db = db.Where("category = ?", query.Category)
	}
	if query.Difficulty != "" {
		db = db.Where("difficulty = ?", query.Difficulty)
	}
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}
	if query.Keyword != "" {
		db = db.Where("name LIKE ? OR description LIKE ?", 
			"%"+query.Keyword+"%", "%"+query.Keyword+"%")
	}
	
	// 统计总数
	var total int64
	db.Count(&total)
	
	// 分页查询
	var templates []model.ChallengeTemplate
	offset := (query.Page - 1) * query.PageSize
	if err := db.Order("created_at DESC").
		Offset(offset).
		Limit(query.PageSize).
		Find(&templates).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"code":    500,
			"message": "failed to query templates",
		})
	}
	
	return c.JSON(fiber.Map{
		"code":    0,
		"message": "ok",
		"data": fiber.Map{
			"list":     templates,
			"total":    total,
			"page":     query.Page,
			"pageSize": query.PageSize,
		},
	})
}

// Get 获取单个模板详情
// GET /api/v1/templates/:id
func (h *TemplateHandler) Get(c fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"code":    400,
			"message": "invalid template id",
		})
	}
	
	var template model.ChallengeTemplate
	if err := h.db.First(&template, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(404).JSON(fiber.Map{
				"code":    404,
				"message": "template not found",
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"code":    500,
			"message": "failed to query template",
		})
	}
	
	return c.JSON(fiber.Map{
		"code":    0,
		"message": "ok",
		"data":    template,
	})
}

// Create 创建模板
// POST /api/v1/templates
func (h *TemplateHandler) Create(c fiber.Ctx) error {
	var template model.ChallengeTemplate
	if err := c.Bind().Body(&template); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"code":    400,
			"message": "invalid request body",
		})
	}
	
	// 验证必填字段
	if template.Name == "" {
		return c.Status(400).JSON(fiber.Map{
			"code":    400,
			"message": "template name is required",
		})
	}
	if template.Category == "" {
		return c.Status(400).JSON(fiber.Map{
			"code":    400,
			"message": "category is required",
		})
	}
	
	// 检查名称是否已存在
	var count int64
	h.db.Model(&model.ChallengeTemplate{}).Where("name = ?", template.Name).Count(&count)
	if count > 0 {
		return c.Status(400).JSON(fiber.Map{
			"code":    400,
			"message": "template name already exists",
		})
	}
	
	// 设置默认值
	template.CreatedAt = time.Now()
	template.UpdatedAt = time.Now()
	if template.Status == "" {
		template.Status = "draft"
	}
	if template.Difficulty == "" {
		template.Difficulty = "medium"
	}
	
	// 保存到数据库
	if err := h.db.Create(&template).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"code":    500,
			"message": "failed to create template",
		})
	}
	
	return c.Status(201).JSON(fiber.Map{
		"code":    0,
		"message": "template created successfully",
		"data":    template,
	})
}

// Update 更新模板
// PUT /api/v1/templates/:id
func (h *TemplateHandler) Update(c fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"code":    400,
			"message": "invalid template id",
		})
	}
	
	var template model.ChallengeTemplate
	if err := h.db.First(&template, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(404).JSON(fiber.Map{
				"code":    404,
				"message": "template not found",
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"code":    500,
			"message": "failed to query template",
		})
	}
	
	var updateData model.ChallengeTemplate
	if err := c.Bind().Body(&updateData); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"code":    400,
			"message": "invalid request body",
		})
	}
	
	// 检查名称是否与其他模板冲突
	if updateData.Name != "" && updateData.Name != template.Name {
		var count int64
		h.db.Model(&model.ChallengeTemplate{}).
			Where("name = ? AND id != ?", updateData.Name, id).
			Count(&count)
		if count > 0 {
			return c.Status(400).JSON(fiber.Map{
				"code":    400,
				"message": "template name already exists",
			})
		}
		template.Name = updateData.Name
	}
	
	// 更新字段
	if updateData.Category != "" {
		template.Category = updateData.Category
	}
	if updateData.Description != "" {
		template.Description = updateData.Description
	}
	if updateData.ImageConfig.Name != "" {
		template.ImageConfig = updateData.ImageConfig
	}
	if len(updateData.ServicePorts) > 0 {
		template.ServicePorts = updateData.ServicePorts
	}
	if updateData.VulnConfig.Type != "" {
		template.VulnConfig = updateData.VulnConfig
	}
	if updateData.FlagConfig.Type != "" {
		template.FlagConfig = updateData.FlagConfig
	}
	if updateData.Difficulty != "" {
		template.Difficulty = updateData.Difficulty
	}
	if updateData.BaseScore > 0 {
		template.BaseScore = updateData.BaseScore
	}
	if updateData.CPULimit > 0 {
		template.CPULimit = updateData.CPULimit
	}
	if updateData.MemLimit > 0 {
		template.MemLimit = updateData.MemLimit
	}
	if updateData.Hints != "" {
		template.Hints = updateData.Hints
	}
	if updateData.Status != "" {
		template.Status = updateData.Status
	}
	
	template.UpdatedAt = time.Now()
	
	if err := h.db.Save(&template).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"code":    500,
			"message": "failed to update template",
		})
	}
	
	return c.JSON(fiber.Map{
		"code":    0,
		"message": "template updated successfully",
		"data":    template,
	})
}

// Delete 删除模板
// DELETE /api/v1/templates/:id
func (h *TemplateHandler) Delete(c fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"code":    400,
			"message": "invalid template id",
		})
	}
	
	result := h.db.Delete(&model.ChallengeTemplate{}, id)
	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"code":    500,
			"message": "failed to delete template",
		})
	}
	
	if result.RowsAffected == 0 {
		return c.Status(404).JSON(fiber.Map{
			"code":    404,
			"message": "template not found",
		})
	}
	
	return c.JSON(fiber.Map{
		"code":    0,
		"message": "template deleted successfully",
	})
}

// Preview 预览模板配置
// GET /api/v1/templates/:id/preview
func (h *TemplateHandler) Preview(c fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"code":    400,
			"message": "invalid template id",
		})
	}
	
	var template model.ChallengeTemplate
	if err := h.db.First(&template, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(404).JSON(fiber.Map{
				"code":    404,
				"message": "template not found",
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"code":    500,
			"message": "failed to query template",
		})
	}
	
	// 生成预览数据
	preview := model.TemplatePreview{
		
		DockerCommand:     h.generateDockerCommand(&template),
		PortMapping:       h.generatePortMapping(&template),
		EnvList:           h.generateEnvList(&template),
	}
	
	return c.JSON(fiber.Map{
		"code":    0,
		"message": "ok",
		"data":    preview,
	})
}

// Export 导出模板为JSON
// GET /api/v1/templates/:id/export
func (h *TemplateHandler) Export(c fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"code":    400,
			"message": "invalid template id",
		})
	}
	
	var template model.ChallengeTemplate
	if err := h.db.First(&template, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(404).JSON(fiber.Map{
				"code":    404,
				"message": "template not found",
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"code":    500,
			"message": "failed to query template",
		})
	}
	
	export := model.TemplateExport{
		Version:    "1.0",
		ExportedAt: time.Now(),
		Template:   template,
	}
	
	// 设置下载头
	c.Set("Content-Type", "application/json")
	c.Set("Content-Disposition", 
		fmt.Sprintf("attachment; filename=template_%s_%d.json", 
			template.Name, time.Now().Unix()))
	
	return c.JSON(export)
}

// Import 导入模板
// POST /api/v1/templates/import
func (h *TemplateHandler) Import(c fiber.Ctx) error {
	var importReq model.TemplateImport
	if err := c.Bind().Body(&importReq); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"code":    400,
			"message": "invalid request body",
		})
	}
	
	template := importReq.Template.Template
	
	// 验证版本
	if importReq.Template.Version == "" {
		return c.Status(400).JSON(fiber.Map{
			"code":    400,
			"message": "missing version field",
		})
	}
	
	// 检查模板是否已存在
	var existingTemplate model.ChallengeTemplate
	err := h.db.Where("name = ?", template.Name).First(&existingTemplate).Error
	
	if err == nil {
		// 模板已存在
		if !importReq.Overwrite {
			return c.Status(400).JSON(fiber.Map{
				"code":    400,
				"message": "template with this name already exists",
			})
		}
		// 覆盖现有模板
		template.ID = existingTemplate.ID
		template.CreatedAt = existingTemplate.CreatedAt
		template.UpdatedAt = time.Now()
		if err := h.db.Save(&template).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{
				"code":    500,
				"message": "failed to update template",
			})
		}
	} else if err == gorm.ErrRecordNotFound {
		// 创建新模板
		template.ID = 0
		template.CreatedAt = time.Now()
		template.UpdatedAt = time.Now()
		if err := h.db.Create(&template).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{
				"code":    500,
				"message": "failed to create template",
			})
		}
	} else {
		return c.Status(500).JSON(fiber.Map{
			"code":    500,
			"message": "database error",
		})
	}
	
	return c.JSON(fiber.Map{
		"code":    0,
		"message": "template imported successfully",
		"data":    template,
	})
}

// Duplicate 复制模板
// POST /api/v1/templates/:id/duplicate
func (h *TemplateHandler) Duplicate(c fiber.Ctx) error {
	id, err := strconv.ParseInt(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"code":    400,
			"message": "invalid template id",
		})
	}
	
	var template model.ChallengeTemplate
	if err := h.db.First(&template, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(404).JSON(fiber.Map{
				"code":    404,
				"message": "template not found",
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"code":    500,
			"message": "failed to query template",
		})
	}
	
	// 创建副本
	newTemplate := template
	newTemplate.ID = 0
	newTemplate.Name = template.Name + " (Copy)"
	newTemplate.CreatedAt = time.Now()
	newTemplate.UpdatedAt = time.Now()
	newTemplate.Status = "draft"
	
	if err := h.db.Create(&newTemplate).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"code":    500,
			"message": "failed to duplicate template",
		})
	}
	
	return c.JSON(fiber.Map{
		"code":    0,
		"message": "template duplicated successfully",
		"data":    newTemplate,
	})
}

// BatchDelete 批量删除模板
// POST /api/v1/templates/batch-delete
func (h *TemplateHandler) BatchDelete(c fiber.Ctx) error {
	var req struct {
		IDs []int64 `json:"ids"`
	}
	
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"code":    400,
			"message": "invalid request body",
		})
	}
	
	if len(req.IDs) == 0 {
		return c.Status(400).JSON(fiber.Map{
			"code":    400,
			"message": "no template ids provided",
		})
	}
	
	if err := h.db.Delete(&model.ChallengeTemplate{}, req.IDs).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"code":    500,
			"message": "failed to delete templates",
		})
	}
	
	return c.JSON(fiber.Map{
		"code":    0,
		"message": "templates deleted successfully",
	})
}

// 辅助函数：生成Docker命令
func (h *TemplateHandler) generateDockerCommand(template *model.ChallengeTemplate) string {
	cmd := fmt.Sprintf("docker run -d")
	
	// 名称
	cmd += fmt.Sprintf(" --name %s", template.Name)
	
	// 资源限制
	if template.MemLimit > 0 {
		cmd += fmt.Sprintf(" --memory=%dm", template.MemLimit)
	}
	if template.CPULimit > 0 {
		cmd += fmt.Sprintf(" --cpus=%.1f", template.CPULimit)
	}
	
	// 端口映射
	for _, port := range template.ServicePorts {
		cmd += fmt.Sprintf(" -p %d:%d/%s", port.Port, port.Port, port.Protocol)
	}
	
	// 环境变量
	for k, v := range template.ImageConfig.Environment {
		cmd += fmt.Sprintf(" -e %s=%s", k, v)
	}
	
	// 特权模式
	if template.ImageConfig.Privileged {
		cmd += " --privileged"
	}
	
	// 网络模式
	if template.ImageConfig.NetworkMode != "" {
		cmd += fmt.Sprintf(" --network=%s", template.ImageConfig.NetworkMode)
	}
	
	// 镜像
	image := template.ImageConfig.Name
	if template.ImageConfig.ImageTag != "" {
		image += ":" + template.ImageConfig.ImageTag
	}
	cmd += fmt.Sprintf(" %s", image)
	
	return cmd
}

// 辅助函数：生成端口映射
func (h *TemplateHandler) generatePortMapping(template *model.ChallengeTemplate) map[int]int {
	mapping := make(map[int]int)
	for _, port := range template.ServicePorts {
		mapping[port.Port] = port.Port
	}
	return mapping
}

// 辅助函数：生成环境变量列表
func (h *TemplateHandler) generateEnvList(template *model.ChallengeTemplate) []string {
	envList := make([]string, 0, len(template.ImageConfig.Environment))
	for k, v := range template.ImageConfig.Environment {
		envList = append(envList, fmt.Sprintf("%s=%s", k, v))
	}
	return envList
}

// BatchExport 批量导出模板
// POST /api/v1/templates/batch-export
func (h *TemplateHandler) BatchExport(c fiber.Ctx) error {
	var req struct {
		IDs []int64 `json:"ids"`
	}
	
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"code":    400,
			"message": "invalid request body",
		})
	}
	
	if len(req.IDs) == 0 {
		return c.Status(400).JSON(fiber.Map{
			"code":    400,
			"message": "no template ids provided",
		})
	}
	
	var templates []model.ChallengeTemplate
	if err := h.db.Find(&templates, req.IDs).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"code":    500,
			"message": "failed to query templates",
		})
	}
	
	exports := make([]model.TemplateExport, len(templates))
	for i, template := range templates {
		exports[i] = model.TemplateExport{
			Version:    "1.0",
			ExportedAt: time.Now(),
			Template:   template,
		}
	}
	
	// 设置下载头
	c.Set("Content-Type", "application/json")
	c.Set("Content-Disposition", 
		fmt.Sprintf("attachment; filename=templates_export_%d.json", time.Now().Unix()))
	
	return c.JSON(fiber.Map{
		"version":     "1.0",
		"exported_at": time.Now(),
		"count":       len(exports),
		"templates":   exports,
	})
}


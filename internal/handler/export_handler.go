package handler

import (
	"time"

	"github.com/awd-platform/awd-arena/internal/model"
	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)


var ExportHandler = &exportHandler{}
// ExportHandler 导出处理器
type exportHandler struct{}

// ExportRankingCSV 导出排名CSV
// GET /api/v1/games/:id/export/ranking/csv
func (h *exportHandler) ExportRankingCSV(c fiber.Ctx) error {
	gameID := parseID(c.Params("id"))
	
	db := c.Locals("db").(*gorm.DB)
	
	// 获取排行榜数据
	var gameTeams []model.GameTeam
	if err := db.Where("game_id = ?", gameID).Preload("Team").Find(&gameTeams).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	
	// 获取分数
	var roundScores []model.RoundScore
	if err := db.Where("game_id = ?", gameID).Order("round desc").Find(&roundScores).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	
	// 生成 CSV
	csv := "排名,队伍名称,总分,攻击分,防御分,轮次\n"
	
	// 构建分数映射
	scoreMap := make(map[int64]model.RoundScore)
	for _, score := range roundScores {
		if _, exists := scoreMap[score.TeamID]; !exists {
			scoreMap[score.TeamID] = score
		}
	}
	
	// 生成数据行
	for _, gt := range gameTeams {
		if score, exists := scoreMap[gt.TeamID]; exists {
			csv += formatCSVRow([]interface{}{
				score.Rank,
				gt.Team.Name,
				score.TotalScore,
				score.AttackScore,
				score.DefenseScore,
				score.Round,
			})
		}
	}
	
	c.Set("Content-Type", "text/csv; charset=utf-8")
	c.Set("Content-Disposition", "attachment; filename=ranking_"+time.Now().Format("20060102_150405")+".csv")
	return c.SendString(csv)
}

// ExportRankingPDF 导出排名PDF（简化版）
// GET /api/v1/games/:id/export/ranking/pdf
func (h *exportHandler) ExportRankingPDF(c fiber.Ctx) error {
	gameID := parseID(c.Params("id"))
	
	db := c.Locals("db").(*gorm.DB)
	
	// 获取排行榜数据
	var gameTeams []model.GameTeam
	if err := db.Where("game_id = ?", gameID).Preload("Team").Find(&gameTeams).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	
	// 获取分数
	var roundScores []model.RoundScore
	if err := db.Where("game_id = ?", gameID).Order("rank asc").Find(&roundScores).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	
	// 生成简单的 HTML 报告（避免复杂的 PDF 库）
	html := generateHTMLReport(gameTeams, roundScores)
	
	c.Set("Content-Type", "text/html; charset=utf-8")
	c.Set("Content-Disposition", "attachment; filename=ranking_"+time.Now().Format("20060102_150405")+".html")
	return c.SendString(html)
}

// ExportAttackLog 导出攻击日志
// GET /api/v1/games/:id/export/attacks
func (h *exportHandler) ExportAttackLog(c fiber.Ctx) error {
	gameID := parseID(c.Params("id"))
	
	db := c.Locals("db").(*gorm.DB)
	
	// 查询攻击日志（如果有）
	var attacks []model.AttackLog
	if err := db.Where("game_id = ?", gameID).Order("timestamp desc").Limit(1000).Find(&attacks).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	
	// 生成 CSV
	csv := "时间,攻击队伍,目标队伍,目标IP,攻击类型,严重性\n"
	for _, attack := range attacks {
		csv += formatCSVRow([]interface{}{
			attack.Timestamp.Format("2006-01-02 15:04:05"),
			attack.AttackerTeam,
			attack.TargetTeam,
			attack.TargetIP,
			attack.AttackType,
			attack.Severity,
		})
	}
	
	c.Set("Content-Type", "text/csv; charset=utf-8")
	c.Set("Content-Disposition", "attachment; filename=attack_log_"+time.Now().Format("20060102_150405")+".csv")
	return c.SendString(csv)
}

// ExportAll 批量导出
// GET /api/v1/games/:id/export/all
func (h *exportHandler) ExportAll(c fiber.Ctx) error {
	_ = parseID(c.Params("id")) // gameID not needed for this endpoint
	
	// 返回所有导出链接
	return c.JSON(fiber.Map{
		"code": 0,
		"message": "ok",
		"data": fiber.Map{
			"ranking_csv":    "/api/v1/games/" + c.Params("id") + "/export/ranking/csv",
			"ranking_pdf":    "/api/v1/games/" + c.Params("id") + "/export/ranking/pdf",
			"attack_log":     "/api/v1/games/" + c.Params("id") + "/export/attacks",
		},
	})
}

// Helper functions

func formatCSVRow(data []interface{}) string {
	row := ""
	for i, item := range data {
		if i > 0 {
			row += ","
		}
		row += toString(item)
	}
	return row + "\n"
}

func toString(item interface{}) string {
	switch v := item.(type) {
	case string:
		return v
	case int:
		return formatInt(v)
	case int64:
		return formatInt64(v)
	case float64:
		return formatFloat(v)
	default:
		return ""
	}
}

func generateHTMLReport(teams []model.GameTeam, scores []model.RoundScore) string {
	html := `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>AWD Arena - 排名报告</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        table { border-collapse: collapse; width: 100%; margin-top: 20px; }
        th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
        th { background-color: #4CAF50; color: white; }
        tr:nth-child(even) { background-color: #f2f2f2; }
        h1 { color: #333; }
    </style>
</head>
<body>
    <h1>AWD Arena - 比赛排名报告</h1>
    <p>生成时间: ` + time.Now().Format("2006-01-02 15:04:05") + `</p>
    <table>
        <tr>
            <th>排名</th>
            <th>队伍名称</th>
            <th>总分</th>
            <th>攻击分</th>
            <th>防御分</th>
        </tr>
`
	
	for _, score := range scores {
		teamName := "Unknown"
		for _, team := range teams {
			if team.TeamID == score.TeamID {
				teamName = team.Team.Name
				break
			}
		}
		
		html += `        <tr>
            <td>` + formatInt(score.Rank) + `</td>
            <td>` + teamName + `</td>
            <td>` + formatFloat(score.TotalScore) + `</td>
            <td>` + formatFloat(score.AttackScore) + `</td>
            <td>` + formatFloat(score.DefenseScore) + `</td>
        </tr>
`
	}
	
	html += `    </table>
</body>
</html>`
	
	return html
}

func formatInt(i int) string {
	return formatInt64(int64(i))
}

func formatInt64(i int64) string {
	// Simple int to string conversion
	if i == 0 {
		return "0"
	}
	
	negative := i < 0
	if negative {
		i = -i
	}
	
	var result string
	for i > 0 {
		digit := i % 10
		result = string('0'+digit) + result
		i /= 10
	}
	
	if negative {
		result = "-" + result
	}
	
	return result
}

func formatFloat(f float64) string {
	// Simple float formatting (2 decimal places)
	intPart := int64(f)
	fracPart := int64((f - float64(intPart)) * 100)
	
	result := formatInt64(intPart)
	
	if fracPart < 0 {
		fracPart = -fracPart
	}
	
	result += "." + formatInt64(fracPart/10) + formatInt64(fracPart%10)
	
	return result
}

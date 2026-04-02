package handler

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"time"

	"github.com/awd-platform/awd-arena/internal/database"
	"github.com/awd-platform/awd-arena/internal/model"
	"github.com/gofiber/fiber/v3"
	"github.com/jung-kurt/gofpdf"
)

var ExportHandler = &exportHandler{}

type exportHandler struct{}

// ExportRankingCSV exports rankings as CSV
// GET /api/v1/games/:id/export/ranking/csv
func (h *exportHandler) ExportRankingCSV(c fiber.Ctx) error {
	gameID, _ := parseID(c.Params("id"))
	db := database.GetDB()
	if db == nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": "database not available"})
	}

	var roundScores []model.RoundScore
	if err := db.Where("game_id = ?", gameID).Order("rank asc").Find(&roundScores).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": err.Error()})
	}

	teamMap := make(map[int64]string)
	var teams []model.Team
	db.Find(&teams)
	for _, t := range teams {
		teamMap[t.ID] = t.Name
	}

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	writer.Write([]string{"Rank", "Team", "TotalScore", "AttackScore", "DefenseScore", "Round"})
	for _, s := range roundScores {
		writer.Write([]string{
			fmt.Sprintf("%d", s.Rank),
			teamMap[s.TeamID],
			fmt.Sprintf("%.2f", s.TotalScore),
			fmt.Sprintf("%.2f", s.AttackScore),
			fmt.Sprintf("%.2f", s.DefenseScore),
			fmt.Sprintf("%d", s.Round),
		})
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": err.Error()})
	}

	c.Set("Content-Type", "text/csv; charset=utf-8")
	c.Set("Content-Disposition", "attachment; filename=ranking_"+time.Now().Format("20060102_150405")+".csv")
	return c.Send(buf.Bytes())
}

// ExportRankingPDF exports rankings as a real PDF document.
// GET /api/v1/games/:id/export/ranking/pdf
func (h *exportHandler) ExportRankingPDF(c fiber.Ctx) error {
	gameID, _ := parseID(c.Params("id"))
	db := database.GetDB()
	if db == nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": "database not available"})
	}

	var roundScores []model.RoundScore
	if err := db.Where("game_id = ?", gameID).Order("rank asc").Find(&roundScores).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": err.Error()})
	}

	teamMap := make(map[int64]string)
	var teams []model.Team
	db.Find(&teams)
	for _, t := range teams {
		teamMap[t.ID] = t.Name
	}

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(0, 10, "AWD Arena Ranking Report")
	pdf.Ln(14)
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(0, 6, "Generated: "+time.Now().Format("2006-01-02 15:04:05"))
	pdf.Ln(10)

	pdf.SetFont("Arial", "B", 10)
	colWidths := []float64{15, 50, 30, 30, 30, 30}
	headers := []string{"Rank", "Team", "Total", "Attack", "Defense", "Round"}
	for i, h := range headers {
		pdf.CellFormat(colWidths[i], 8, h, "1", 0, "C", false, 0, "")
	}
	pdf.Ln(-1)

	pdf.SetFont("Arial", "", 10)
	for _, s := range roundScores {
		cells := []string{
			fmt.Sprintf("%d", s.Rank),
			teamMap[s.TeamID],
			fmt.Sprintf("%.2f", s.TotalScore),
			fmt.Sprintf("%.2f", s.AttackScore),
			fmt.Sprintf("%.2f", s.DefenseScore),
			fmt.Sprintf("%d", s.Round),
		}
		for i, cell := range cells {
			pdf.CellFormat(colWidths[i], 7, cell, "1", 0, "C", false, 0, "")
		}
		pdf.Ln(-1)
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": "pdf generation failed"})
	}

	c.Set("Content-Type", "application/pdf")
	c.Set("Content-Disposition", "attachment; filename=ranking_"+time.Now().Format("20060102_150405")+".pdf")
	return c.Send(buf.Bytes())
}

// ExportAttackLog exports attack logs as CSV
// GET /api/v1/games/:id/export/attacks
func (h *exportHandler) ExportAttackLog(c fiber.Ctx) error {
	gameID, _ := parseID(c.Params("id"))
	db := database.GetDB()
	if db == nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": "database not available"})
	}

	var submissions []model.FlagSubmission
	db.Where("game_id = ?", gameID).Order("submitted_at desc").Limit(1000).Find(&submissions)

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	writer.Write([]string{"Time", "AttackerTeam", "TargetTeam", "Correct", "Points", "Round"})
	for _, s := range submissions {
		correct := "no"
		if s.IsCorrect {
			correct = "yes"
		}
		writer.Write([]string{
			s.SubmittedAt.Format("2006-01-02 15:04:05"),
			fmt.Sprintf("%d", s.AttackerTeam),
			fmt.Sprintf("%d", s.TargetTeam),
			correct,
			fmt.Sprintf("%.2f", s.PointsEarned),
			fmt.Sprintf("%d", s.Round),
		})
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": err.Error()})
	}

	c.Set("Content-Type", "text/csv; charset=utf-8")
	c.Set("Content-Disposition", "attachment; filename=attacks_"+time.Now().Format("20060102_150405")+".csv")
	return c.Send(buf.Bytes())
}

// ExportAll returns all export links
// GET /api/v1/games/:id/export/all
func (h *exportHandler) ExportAll(c fiber.Ctx) error {
	id := c.Params("id")
	return c.JSON(fiber.Map{
		"code": 0,
		"data": fiber.Map{
			"ranking_csv": "/api/v1/games/" + id + "/export/ranking/csv",
			"ranking_pdf": "/api/v1/games/" + id + "/export/ranking/pdf",
			"attack_log":  "/api/v1/games/" + id + "/export/attacks",
		},
	})
}

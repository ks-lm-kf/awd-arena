package handler

import (
    "time"

    "github.com/awd-platform/awd-arena/internal/database"
    "github.com/awd-platform/awd-arena/internal/model"
    "github.com/gofiber/fiber/v3"
)

var ExportHandler = &exportHandler{}

type exportHandler struct{}

// ExportRankingCSV exports rankings as CSV
// GET /api/v1/games/:id/export/ranking/csv
func (h *exportHandler) ExportRankingCSV(c fiber.Ctx) error {
    gameID := parseID(c.Params("id"))
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

    csv := "Rank,Team,TotalScore,AttackScore,DefenseScore,Round\n"
    for _, s := range roundScores {
        csv += itoa(s.Rank) + "," + teamMap[s.TeamID] + "," + ftoa(s.TotalScore) + "," + ftoa(s.AttackScore) + "," + ftoa(s.DefenseScore) + "," + itoa(s.Round) + "\n"
    }

    c.Set("Content-Type", "text/csv; charset=utf-8")
    c.Set("Content-Disposition", "attachment; filename=ranking_"+time.Now().Format("20060102_150405")+".csv")
    return c.SendString(csv)
}

// ExportRankingPDF exports rankings as HTML report
// GET /api/v1/games/:id/export/ranking/pdf
func (h *exportHandler) ExportRankingPDF(c fiber.Ctx) error {
    gameID := parseID(c.Params("id"))
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

    html := `<!DOCTYPE html><html><head><meta charset="UTF-8"><title>AWD Arena Ranking</title>
    <style>body{font-family:Arial,sans-serif;margin:20px}table{border-collapse:collapse;width:100%;margin-top:20px}
    th,td{border:1px solid #ddd;padding:8px;text-align:left}th{background:#4CAF50;color:white}
    tr:nth-child(even){background:#f2f2f2}</style></head><body><h1>AWD Arena Ranking Report</h1>
    <p>Generated: ` + time.Now().Format("2006-01-02 15:04:05") + `</p>
    <table><tr><th>Rank</th><th>Team</th><th>Total</th><th>Attack</th><th>Defense</th></tr>`

    for _, s := range roundScores {
        html += "<tr><td>" + itoa(s.Rank) + "</td><td>" + teamMap[s.TeamID] + "</td><td>" + ftoa(s.TotalScore) + "</td><td>" + ftoa(s.AttackScore) + "</td><td>" + ftoa(s.DefenseScore) + "</td></tr>"
    }
    html += "</table></body></html>"

    c.Set("Content-Type", "text/html; charset=utf-8")
    c.Set("Content-Disposition", "attachment; filename=ranking_"+time.Now().Format("20060102_150405")+".html")
    return c.SendString(html)
}

// ExportAttackLog exports attack logs as CSV
// GET /api/v1/games/:id/export/attacks
func (h *exportHandler) ExportAttackLog(c fiber.Ctx) error {
    gameID := parseID(c.Params("id"))
    db := database.GetDB()
    if db == nil {
        return c.Status(500).JSON(fiber.Map{"code": 500, "message": "database not available"})
    }

    var submissions []model.FlagSubmission
    db.Where("game_id = ?", gameID).Order("submitted_at desc").Limit(1000).Find(&submissions)

    csv := "Time,AttackerTeam,TargetTeam,Correct,Points,Round\n"
    for _, s := range submissions {
        csv += s.SubmittedAt.Format("2006-01-02 15:04:05") + "," + itoa64(s.AttackerTeam) + "," + itoa64(s.TargetTeam) + ","
        if s.IsCorrect { csv += "yes" } else { csv += "no" }
        csv += "," + ftoa(s.PointsEarned) + "," + itoa(s.Round) + "\n"
    }

    c.Set("Content-Type", "text/csv; charset=utf-8")
    c.Set("Content-Disposition", "attachment; filename=attacks_"+time.Now().Format("20060102_150405")+".csv")
    return c.SendString(csv)
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

func parseID(s string) int64 {
    var n int64
    for _, c := range s {
        if c >= '0' && c <= '9' {
            n = n*10 + int64(c-'0')
        }
    }
    return n
}

func itoa(n int) string { return itoa64(int64(n)) }

func itoa64(n int64) string {
    if n == 0 { return "0" }
    neg := n < 0
    if neg { n = -n }
    s := ""
    for n > 0 { s = string('0'+n%10) + s; n /= 10 }
    if neg { s = "-" + s }
    return s
}

func ftoa(f float64) string {
    i := int64(f)
    frac := int64((f - float64(i)) * 100)
    if frac < 0 { frac = -frac }
    return itoa64(i) + "." + string('0'+frac/10%10) + string('0'+frac%10)
}

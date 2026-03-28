package export

import (
	"fmt"
	"os"
	"time"

	"github.com/awd-platform/awd-arena/internal/model"
)

// ExportRankingHTML exports rankings to an HTML file (used as PDF substitute).
func ExportRankingHTML(filepath string, entries []model.RoundScore, teams map[int64]string) error {
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	html := `<!DOCTYPE html><html><head><meta charset="UTF-8"><title>AWD Arena Ranking</title>
<style>
body{font-family:Arial,sans-serif;margin:20px;background:#1a1a2e;color:#eee}
table{border-collapse:collapse;width:100%;margin-top:20px}
th,td{border:1px solid #333;padding:8px;text-align:left}
th{background:#16213e;color:#0f3460;font-weight:bold}
tr:nth-child(even){background:#16213e}
h1{color:#e94560}
</style></head><body><h1>AWD Arena Ranking Report</h1>
<p>Generated: ` + time.Now().Format("2006-01-02 15:04:05") + `</p>
<table><tr><th>Rank</th><th>Team</th><th>Total</th><th>Attack</th><th>Defense</th></tr>
`

	for _, e := range entries {
		html += fmt.Sprintf("<tr><td>%d</td><td>%s</td><td>%.2f</td><td>%.2f</td><td>%.2f</td></tr>\n",
			e.Rank, teams[e.TeamID], e.TotalScore, e.AttackScore, e.DefenseScore)
	}
	html += "</table></body></html>"

	_, err = file.WriteString(html)
	return err
}

// ExportAttackLogHTML exports attack logs to HTML.
func ExportAttackLogHTML(filepath string, submissions []model.FlagSubmission) error {
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	html := `<!DOCTYPE html><html><head><meta charset="UTF-8"><title>Attack Log</title>
<style>
body{font-family:Arial,sans-serif;margin:20px;background:#1a1a2e;color:#eee}
table{border-collapse:collapse;width:100%;margin-top:20px}
th,td{border:1px solid #333;padding:6px;text-align:left}
th{background:#16213e}
tr:nth-child(even){background:#16213e}
.correct{color:#4caf50}.wrong{color:#f44336}
</style></head><body><h1>Attack Log Report</h1>
<p>Generated: ` + time.Now().Format("2006-01-02 15:04:05") + `</p>
<table><tr><th>Time</th><th>Attacker</th><th>Target</th><th>Result</th><th>Points</th><th>Round</th></tr>
`

	for _, s := range submissions {
		cls := "wrong"
		label := "FAIL"
		if s.IsCorrect {
			cls = "correct"
			label = "OK"
		}
		html += fmt.Sprintf("<tr><td>%s</td><td>%d</td><td>%d</td><td class=\"%s\">%s</td><td>%.2f</td><td>%d</td></tr>\n",
			s.SubmittedAt.Format("15:04:05"), s.AttackerTeam, s.TargetTeam, cls, label, s.PointsEarned, s.Round)
	}
	html += "</table></body></html>"

	_, err = file.WriteString(html)
	return err
}

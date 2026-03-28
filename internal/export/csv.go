package export

import (
	"encoding/csv"
	"fmt"
	"os"
	"time"

	"github.com/awd-platform/awd-arena/internal/model"
)

// ExportRankingCSV exports rankings to a CSV file.
func ExportRankingCSV(filepath string, entries []model.RoundScore, teams map[int64]string) error {
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write([]string{"Rank", "Team", "TotalScore", "AttackScore", "DefenseScore", "Round"})
	for _, e := range entries {
		teamName := teams[e.TeamID]
		writer.Write([]string{
			fmt.Sprintf("%d", e.Rank),
			teamName,
			fmt.Sprintf("%.2f", e.TotalScore),
			fmt.Sprintf("%.2f", e.AttackScore),
			fmt.Sprintf("%.2f", e.DefenseScore),
			fmt.Sprintf("%d", e.Round),
		})
	}
	return nil
}

// ExportAttackLogCSV exports attack logs to a CSV file.
func ExportAttackLogCSV(filepath string, submissions []model.FlagSubmission) error {
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

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
	return nil
}

// ExportStatsCSV exports competition statistics to CSV.
func ExportStatsCSV(filepath string, stats *model.CompetitionStats) error {
	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write([]string{"Metric", "Value"})
	writer.Write([]string{"Total Teams", fmt.Sprintf("%d", stats.TotalTeams)})
	writer.Write([]string{"Total Rounds", fmt.Sprintf("%d", stats.TotalRounds)})
	writer.Write([]string{"Total Attacks", fmt.Sprintf("%d", stats.TotalAttacks)})
	writer.Write([]string{"Avg Attacks/Round", fmt.Sprintf("%.2f", stats.AvgAttacksPerRound)})
	writer.Write([]string{"Most Attacked", stats.MostAttackedTeam})
	writer.Write([]string{"Top Attacker", stats.TopAttacker})
	writer.Write([]string{"Exported", time.Now().Format("2006-01-02 15:04:05")})
	return nil
}

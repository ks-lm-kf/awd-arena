package export

import (
	"bytes"
	"encoding/csv"
	"fmt"

	"github.com/awd-platform/awd-arena/internal/model"
)

// GenerateRankingCSV 生成排行榜CSV
func GenerateRankingCSV(rankings []model.Ranking) []byte {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// 写入UTF-8 BOM确保Excel正确识别编码
	buf.Write([]byte{0xEF, 0xBB, 0xBF})

	// 写入表头
	header := []string{"排名", "队伍名称", "总分", "攻击次数", "防御次数", "一血次数", "更新时间"}
	if err := writer.Write(header); err != nil {
		return nil
	}

	// 写入数据
	for _, ranking := range rankings {
		record := []string{
			fmt.Sprintf("%d", ranking.Rank),
			ranking.TeamName,
			fmt.Sprintf("%.2f", ranking.Score),
			fmt.Sprintf("%d", ranking.Attacks),
			fmt.Sprintf("%d", ranking.Defenses),
			fmt.Sprintf("%d", ranking.FirstBlood),
			ranking.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
		if err := writer.Write(record); err != nil {
			return nil
		}
	}

	writer.Flush()
	return buf.Bytes()
}

// GenerateAttackLogCSV 生成攻击日志CSV
func GenerateAttackLogCSV(attacks []model.AttackLog) []byte {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// 写入UTF-8 BOM
	buf.Write([]byte{0xEF, 0xBB, 0xBF})

	// 写入表头
	header := []string{"时间", "轮次", "攻击方", "目标队伍", "目标IP", "目标端口", "协议", "攻击类型", "严重程度", "来源IP"}
	if err := writer.Write(header); err != nil {
		return nil
	}

	// 写入数据
	for _, attack := range attacks {
		record := []string{
			attack.Timestamp.Format("2006-01-02 15:04:05"),
			fmt.Sprintf("%d", attack.Round),
			attack.AttackerTeam,
			attack.TargetTeam,
			attack.TargetIP,
			fmt.Sprintf("%d", attack.TargetPort),
			attack.Protocol,
			attack.AttackType,
			attack.Severity,
			attack.SourceIP,
		}
		if err := writer.Write(record); err != nil {
			return nil
		}
	}

	writer.Flush()
	return buf.Bytes()
}

// GenerateRoundCSV 生成轮次排名CSV
func GenerateRoundCSV(rankings []model.RoundRanking) []byte {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	buf.Write([]byte{0xEF, 0xBB, 0xBF})

	header := []string{"轮次", "队伍名称", "得分", "更新时间"}
	if err := writer.Write(header); err != nil {
		return nil
	}

	for _, ranking := range rankings {
		record := []string{
			fmt.Sprintf("%d", ranking.Round),
			ranking.TeamName,
			fmt.Sprintf("%.2f", ranking.Score),
			ranking.UpdatedAt.Format("2006-01-02 15:04:05"),
		}
		if err := writer.Write(record); err != nil {
			return nil
		}
	}

	writer.Flush()
	return buf.Bytes()
}

// GenerateStatisticsCSV 生成统计信息CSV
func GenerateStatisticsCSV(stats model.CompetitionStats) []byte {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	buf.Write([]byte{0xEF, 0xBB, 0xBF})

	// 写入基本信息
	writer.Write([]string{"比赛统计信息"})
	writer.Write([]string{""})
	writer.Write([]string{"参赛队伍总数", fmt.Sprintf("%d", stats.TotalTeams)})
	writer.Write([]string{"总攻击次数", fmt.Sprintf("%d", stats.TotalAttacks)})
	writer.Write([]string{"平均得分", fmt.Sprintf("%.2f", stats.AvgScore)})
	writer.Write([]string{"领先队伍", stats.TopTeam})
	writer.Write([]string{"最高得分", fmt.Sprintf("%.2f", stats.TopScore)})

	writer.Flush()
	return buf.Bytes()
}

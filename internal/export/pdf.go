package export

import (
	"archive/zip"
	"bytes"
	"fmt"
	"time"

	"github.com/awd-platform/awd-arena/internal/model"

	"github.com/jung-kurt/gofpdf"
)

// GenerateRankingPDF 生成排行榜PDF
func GenerateRankingPDF(rankings []model.Ranking) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetAutoPageBreak(true, 10)

	// 加载字体（支持中文）
	pdf.AddUTF8Font("simhei", "", "simhei.ttf")
	pdf.SetFont("simhei", "", 16)

	// 标题
	pdf.CellFormat(190, 10, "AWD Arena - 排行榜", "", 0, "C", false, 0, "")
	pdf.Ln(15)

	// 生成时间
	pdf.SetFont("simhei", "", 10)
	pdf.CellFormat(190, 5, fmt.Sprintf("生成时间: %s", time.Now().Format("2006-01-02 15:04:05")), "", 0, "R", false, 0, "")
	pdf.Ln(10)

	// 表头
	pdf.SetFont("simhei", "", 11)
	pdf.SetFillColor(200, 220, 255)
	
	colWidths := []float64{15, 50, 25, 25, 25, 25, 25}
	headers := []string{"排名", "队伍名称", "总分", "攻击次数", "防御次数", "一血次数", "更新时间"}
	
	for i, header := range headers {
		pdf.CellFormat(colWidths[i], 8, header, "1", 0, "C", true, 0, "")
	}
	pdf.Ln(-1)

	// 表格数据
	pdf.SetFont("simhei", "", 10)
	for _, ranking := range rankings {
		pdf.CellFormat(colWidths[0], 7, fmt.Sprintf("%d", ranking.Rank), "1", 0, "C", false, 0, "")
		pdf.CellFormat(colWidths[1], 7, 		pdf.CellFormat(colWidths[2], 7, fmt.Sprintf("%.1f", 		pdf.CellFormat(colWidths[3], 7, fmt.Sprintf("%d", ranking.Attacks), "1", 0, "C", false, 0, "")
		pdf.CellFormat(colWidths[4], 7, fmt.Sprintf("%d", ranking.Defenses), "1", 0, "C", false, 0, "")
		pdf.CellFormat(colWidths[5], 7, fmt.Sprintf("%d", ranking.FirstBlood), "1", 0, "C", false, 0, "")
		pdf.CellFormat(colWidths[6], 7, 		pdf.Ln(-1)
	}

	// 生成PDF字节
	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// GenerateFullReport 生成完整比赛报告PDF
func GenerateFullReport(rankings []model.Ranking, attacks []model.AttackLog) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetAutoPageBreak(true, 10)

	// 加载字体
	pdf.AddUTF8Font("simhei", "", "simhei.ttf")
	pdf.SetFont("simhei", "", 16)

	// 封面标题
	pdf.CellFormat(190, 10, "AWD Arena", "", 0, "C", false, 0, "")
	pdf.Ln(12)
	pdf.SetFont("simhei", "", 24)
	pdf.CellFormat(190, 15, "比赛报告", "", 0, "C", false, 0, "")
	pdf.Ln(20)
	
	pdf.SetFont("simhei", "", 12)
	pdf.CellFormat(190, 8, fmt.Sprintf("报告生成时间: %s", time.Now().Format("2006-01-02 15:04:05")), "", 0, "C", false, 0, "")
	pdf.Ln(30)

	// 第一部分：排行榜
	pdf.SetFont("simhei", "", 16)
	pdf.CellFormat(190, 10, "一、排行榜", "", 0, "L", false, 0, "")
	pdf.Ln(10)

	pdf.SetFont("simhei", "", 11)
	pdf.SetFillColor(200, 220, 255)
	
	colWidths := []float64{15, 50, 30, 30, 30, 35}
	headers := []string{"排名", "队伍名称", "总分", "攻击次数", "防御次数", "一血次数"}
	
	for i, header := range headers {
		pdf.CellFormat(colWidths[i], 8, header, "1", 0, "C", true, 0, "")
	}
	pdf.Ln(-1)

	pdf.SetFont("simhei", "", 10)
	for _, ranking := range rankings {
		pdf.CellFormat(colWidths[0], 7, fmt.Sprintf("%d", ranking.Rank), "1", 0, "C", false, 0, "")
		pdf.CellFormat(colWidths[1], 7, 		pdf.CellFormat(colWidths[2], 7, fmt.Sprintf("%.1f", 		pdf.CellFormat(colWidths[3], 7, fmt.Sprintf("%d", ranking.Attacks), "1", 0, "C", false, 0, "")
		pdf.CellFormat(colWidths[4], 7, fmt.Sprintf("%d", ranking.Defenses), "1", 0, "C", false, 0, "")
		pdf.CellFormat(colWidths[5], 7, fmt.Sprintf("%d", ranking.FirstBlood), "1", 0, "C", false, 0, "")
		pdf.Ln(-1)
	}

	pdf.Ln(15)

	// 第二部分：攻击统计数据
	pdf.SetFont("simhei", "", 16)
	pdf.CellFormat(190, 10, "二、攻击统计", "", 0, "L", false, 0, "")
	pdf.Ln(10)

	// 统计数据
	totalAttacks := len(attacks)
	severityCount := make(map[string]int)
	attackTypeCount := make(map[string]int)
	
	for _, a := range attacks {
		severityCount[a.Severity]++
		attackTypeCount[a.AttackType]++
	}

	pdf.SetFont("simhei", "", 11)
	pdf.CellFormat(190, 7, fmt.Sprintf("总攻击记录数: %d", totalAttacks), "", 0, "L", false, 0, "")
	pdf.Ln(8)
	
	pdf.CellFormat(190, 7, "按严重程度分类:", "", 0, "L", false, 0, "")
	pdf.Ln(6)
	for sev, count := range severityCount {
		pdf.CellFormat(190, 6, fmt.Sprintf("  %s: %d", sev, count), "", 0, "L", false, 0, "")
		pdf.Ln(5)
	}
	pdf.Ln(15)

	// 第三部分：攻击记录详情
	pdf.SetFont("simhei", "", 16)
	pdf.CellFormat(190, 10, "三、攻击记录（前20条）", "", 0, "L", false, 0, "")
	pdf.Ln(10)

	pdf.SetFont("simhei", "", 9)
	pdf.SetFillColor(200, 220, 255)
	
	attackColWidths := []float64{30, 15, 30, 30, 35, 25, 25}
	attackHeaders := []string{"时间", "轮次", "攻击方", "目标队伍", "目标IP", "攻击类型", "严重程度"}
	
	for i, header := range attackHeaders {
		pdf.CellFormat(attackColWidths[i], 7, header, "1", 0, "C", true, 0, "")
	}
	pdf.Ln(-1)

	pdf.SetFont("simhei", "", 8)
	maxRecords := 20
	if len(attacks) < maxRecords {
		maxRecords = len(attacks)
	}
	
	for i := 0; i < maxRecords; i++ {
		attack := attacks[i]
		pdf.CellFormat(attackColWidths[0], 6, attack.Timestamp.Format("01-02 15:04"), "1", 0, "C", false, 0, "")
		pdf.CellFormat(attackColWidths[1], 6, fmt.Sprintf("%d", attack.Round), "1", 0, "C", false, 0, "")
		pdf.CellFormat(attackColWidths[2], 6, truncateString(attack.AttackerTeam, 12), "1", 0, "L", false, 0, "")
		pdf.CellFormat(attackColWidths[3], 6, truncateString(attack.TargetTeam, 12), "1", 0, "L", false, 0, "")
		pdf.CellFormat(attackColWidths[4], 6, truncateString(attack.TargetIP, 15), "1", 0, "L", false, 0, "")
		pdf.CellFormat(attackColWidths[5], 6, truncateString(attack.AttackType, 10), "1", 0, "C", false, 0, "")
		pdf.CellFormat(attackColWidths[6], 6, attack.Severity, "1", 0, "C", false, 0, "")
		pdf.Ln(-1)
	}

	// 页脚
	pdf.Ln(10)
	pdf.SetFont("simhei", "", 9)
	pdf.CellFormat(190, 5, "--- 报告结束 ---", "", 0, "C", false, 0, "")

	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// ExportAllData 导出所有数据为ZIP
func ExportAllData(rankings []model.Ranking, attacks []model.AttackLog) ([]byte, error) {
	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)

	// 1. 排行榜CSV
	rankingCSV := GenerateRankingCSV(rankings)
	if err := addFileToZip(zipWriter, "ranking.csv", rankingCSV); err != nil {
		return nil, err
	}

	// 2. 攻击日志CSV
	attackLogCSV := GenerateAttackLogCSV(attacks)
	if err := addFileToZip(zipWriter, "attack_log.csv", attackLogCSV); err != nil {
		return nil, err
	}

	// 3. 排行榜PDF
	rankingPDF, err := GenerateRankingPDF(rankings)
	if err != nil {
		return nil, err
	}
	if err := addFileToZip(zipWriter, "ranking.pdf", rankingPDF); err != nil {
		return nil, err
	}

	// 4. 完整比赛报告PDF
	reportPDF, err := GenerateFullReport(rankings, attacks)
	if err != nil {
		return nil, err
	}
	if err := addFileToZip(zipWriter, "competition_report.pdf", reportPDF); err != nil {
		return nil, err
	}

	// 关闭ZIP
	if err := zipWriter.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// addFileToZip 添加文件到ZIP
func addFileToZip(zipWriter *zip.Writer, filename string, data []byte) error {
	writer, err := zipWriter.Create(filename)
	if err != nil {
		return err
	}
	_, err = writer.Write(data)
	return err
}

// truncateString 截断字符串
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

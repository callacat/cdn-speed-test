package output

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"

	"github.com/olekukonko/tablewriter"
	"github.com/your-username/cdn-speed-test/internal/models"
)

// RenderResults 将结果渲染到终端或CSV文件
func RenderResults(results []*models.IPInfo, format string, csvPath string) {
	if format == "csv" {
		saveToCSV(results, csvPath)
		fmt.Printf("✅ 结果已保存到 %s\n", csvPath)
	} else {
		renderTable(results)
	}
}

// renderTable 在终端渲染表格
func renderTable(results []*models.IPInfo) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"IP 地址", "平均延迟", "下载速度 (MB/s)", "丢包率"})
	table.SetBorders(tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})
	table.SetCenterSeparator("|")

	for _, res := range results {
		if res.IsAvailable {
			table.Append([]string{
				res.IP.String(),
				fmt.Sprintf("%.2f ms", float64(res.Latency.Microseconds())/1000.0),
				fmt.Sprintf("%.2f", res.DownloadSpeed),
				fmt.Sprintf("%.2f%%", res.PacketLoss*100),
			})
		}
	}
	table.Render()
}

// saveToCSV 将结果保存到CSV文件
func saveToCSV(results []*models.IPInfo, path string) {
	file, err := os.Create(path)
	if err != nil {
		fmt.Println("❌ 创建CSV文件失败:", err)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write([]string{"IP Address", "Avg Latency (ms)", "Download Speed (MB/s)", "Packet Loss (%)"})
	for _, res := range results {
		if res.IsAvailable {
			writer.Write([]string{
				res.IP.String(),
				strconv.FormatFloat(float64(res.Latency.Microseconds())/1000.0, 'f', 2, 64),
				strconv.FormatFloat(res.DownloadSpeed, 'f', 2, 64),
				strconv.FormatFloat(res.PacketLoss*100, 'f', 2, 64),
			})
		}
	}
}

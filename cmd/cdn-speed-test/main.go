package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/your-username/cdn-speed-test/internal/config"
	"github.com/your-username/cdn-speed-test/internal/ip_source"
	"github.com/your-username/cdn-speed-test/internal/output"
	"github.com/your-username/cdn-speed-test/internal/tester"
)

func main() {
	fmt.Println("🚀 通用 CDN 优选IP测试工具 v1.0.0")

	// 1. 加载配置
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		fmt.Println("❌ 加载配置失败:", err)
		os.Exit(1)
	}
	fmt.Println("✅ 配置加载成功")

	// 2. 获取IP列表
	ips, err := ip_source.GetIPs(cfg)
	if err != nil {
		fmt.Println("❌ 获取IP列表失败:", err)
		os.Exit(1)
	}
	fmt.Printf("✅ 成功获取 %d 个IP地址\n", len(ips))

	// 3. 阶段一：TCP延迟测试
	fmt.Println("\n--- 阶段一：TCP延迟和丢包率测试 ---")
	initialResults := tester.RunTCPPingTests(ips, cfg.Test.Concurrency, cfg.Test.Retries, cfg.Test.Timeout, cfg.Test.LatencyMax)
	fmt.Printf("✅ %d 个IP通过初步筛选\n", len(initialResults))

	// 4. 排序并选取TopN
	sort.Slice(initialResults, func(i, j int) bool {
		if initialResults[i].PacketLoss != initialResults[j].PacketLoss {
			return initialResults[i].PacketLoss < initialResults[j].PacketLoss
		}
		return initialResults[i].Latency < initialResults[j].Latency
	})

	topN := cfg.Test.TopN
	if len(initialResults) < topN {
		topN = len(initialResults)
	}
	topIPs := initialResults[:topN]
	fmt.Printf("✅ 选取延迟最低的 %d 个IP进入下一阶段测试\n", len(topIPs))

	// 5. 阶段二：HTTP可用性和速度测试
	fmt.Println("\n--- 阶段二：HTTP可用性和速度测试 ---")
	tester.RunHTTPTests(topIPs, cfg.Test.Concurrency, cfg.HTTP.TargetURL, cfg.HTTP.SpeedTestURL, cfg.Test.Timeout, cfg.HTTP.SpeedTestTimeout)
	fmt.Println("✅ HTTP测试完成")


	// 6. 最终排序
	sort.Slice(topIPs, func(i, j int) bool {
		// 速度优先，然后是延迟
		if topIPs[i].DownloadSpeed != topIPs[j].DownloadSpeed {
			return topIPs[i].DownloadSpeed > topIPs[j].DownloadSpeed
		}
		return topIPs[i].Latency < topIPs[j].Latency
	})

	// 7. 输出结果
	fmt.Println("\n--- 🏆 测试完成，最佳IP如下 ---")
	output.RenderResults(topIPs, cfg.Output.Format, cfg.Output.CSVPath)
}

package tester

import (
	"net"
	"sync"
	"time"

	"github.com/your-username/cdn-speed-test/internal/models"
)

// PingResult 包含单次ping的结果
type PingResult struct {
	Latency time.Duration
	Success bool
}

// TCPPing 对单个IP进行多次TCP Ping测试
func TCPPing(ip net.IP, retries int, timeout time.Duration) (time.Duration, float64) {
	var totalLatency time.Duration
	successCount := 0
	var wg sync.WaitGroup
	results := make(chan PingResult, retries)

	for i := 0; i < retries; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			start := time.Now()
			conn, err := net.DialTimeout("tcp", net.JoinHostPort(ip.String(), "443"), timeout)
			latency := time.Since(start)
			if err == nil {
				conn.Close()
				results <- PingResult{Latency: latency, Success: true}
			} else {
				results <- PingResult{Latency: 0, Success: false}
			}
		}()
	}

	wg.Wait()
	close(results)

	for res := range results {
		if res.Success {
			totalLatency += res.Latency
			successCount++
		}
	}

	if successCount == 0 {
		return 0, 1.0
	}

	avgLatency := totalLatency / time.Duration(successCount)
	packetLoss := 1.0 - (float64(successCount) / float64(retries))

	return avgLatency, packetLoss
}

// RunTCPPingTests 并发对所有IP进行TCP Ping测试
func RunTCPPingTests(ips []net.IP, concurrency int, retries int, timeout time.Duration, latencyMax int) []*models.IPInfo {
	var wg sync.WaitGroup
	ipChan := make(chan net.IP, len(ips))
	results := make(chan *models.IPInfo, len(ips))

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for ip := range ipChan {
				avgLatency, packetLoss := TCPPing(ip, retries, timeout)
				if avgLatency > 0 && avgLatency.Milliseconds() <= int64(latencyMax) && packetLoss == 0 {
					results <- &models.IPInfo{
						IP:         ip,
						Latency:    avgLatency,
						PacketLoss: packetLoss,
					}
				}
			}
		}()
	}

	for _, ip := range ips {
		ipChan <- ip
	}
	close(ipChan)

	wg.Wait()
	close(results)

	var finalResults []*models.IPInfo
	for res := range results {
		finalResults = append(finalResults, res)
	}

	return finalResults
}

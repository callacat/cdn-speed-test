package tester

import (
	"net"
	"sync"
	"time"

	"github.com/callacat/cdn-speed-test/internal/models"
	"github.com/schollz/progressbar/v3"
)

// runTCPing performs a single TCP "ping" to a given IP on port 443.
func runTCPing(ip net.IP, timeout time.Duration) (time.Duration, bool) {
	start := time.Now()
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(ip.String(), "443"), timeout)
	if err != nil {
		return 0, false
	}
	conn.Close()
	return time.Since(start), true
}

// testIPLatency performs multiple TCP pings to calculate average latency and packet loss.
func testIPLatency(ip net.IP, retries int, timeout time.Duration) (time.Duration, float64) {
	var totalLatency time.Duration
	successCount := 0

	for i := 0; i < retries; i++ {
		latency, success := runTCPing(ip, timeout)
		if success {
			totalLatency += latency
			successCount++
		}
	}

	if successCount == 0 {
		return 0, 1.0 // 100% packet loss
	}

	avgLatency := totalLatency / time.Duration(successCount)
	packetLoss := 1.0 - (float64(successCount) / float64(retries))

	return avgLatency, packetLoss
}

// RunTCPTests concurrently tests the latency of all provided IPs.
func RunTCPTests(ips []net.IP, concurrency int, retries int, timeout time.Duration, latencyMax int, bar *progressbar.ProgressBar) []*models.IPInfo {
	var wg sync.WaitGroup
	ipChan := make(chan net.IP, len(ips))
	resultsChan := make(chan *models.IPInfo, len(ips))

	// Start worker goroutines
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for ip := range ipChan {
				avgLatency, packetLoss := testIPLatency(ip, retries, timeout)

				// Filter out IPs that don't meet the criteria
				if packetLoss == 0 && avgLatency > 0 && avgLatency.Milliseconds() <= int64(latencyMax) {
					resultsChan <- &models.IPInfo{
						IP:         ip,
						Latency:    avgLatency,
						PacketLoss: packetLoss,
					}
				}
				bar.Add(1)
			}
		}()
	}

	// Feed IPs to the workers
	for _, ip := range ips {
		ipChan <- ip
	}
	close(ipChan)

	// Wait for all workers to finish
	wg.Wait()
	close(resultsChan)

	// Collect results
	var finalResults []*models.IPInfo
	for res := range resultsChan {
		finalResults = append(finalResults, res)
	}

	return finalResults
}

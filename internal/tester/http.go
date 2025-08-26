package tester

import (
	"context"
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/callacat/cdn-speed-test/internal/models"
	"github.com/schollz/progressbar/v3"
)

// testIPConnectivityAndSpeed performs HTTP availability and speed tests for a single IP.
// It updates the passed IPInfo struct directly.
func testIPConnectivityAndSpeed(ipInfo *models.IPInfo, targetURL, speedTestURL string, timeout, speedTestTimeout time.Duration) {
	// Create a custom transport that forces connections to our target IP.
	dialer := &net.Dialer{
		Timeout:   timeout,
		KeepAlive: 30 * time.Second,
	}
	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			// Override the address to use our specific IP and port 443.
			return dialer.DialContext(ctx, network, net.JoinHostPort(ipInfo.IP.String(), "443"))
		},
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: true}, // Ignore self-signed certs
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	// 1. Availability Test
	client := &http.Client{Transport: transport, Timeout: timeout}
	req, err := http.NewRequest("HEAD", targetURL, nil)
	if err != nil {
		ipInfo.IsAvailable = false
		return
	}

	resp, err := client.Do(req)
	if err != nil || (resp.StatusCode < 200 || resp.StatusCode >= 400) {
		ipInfo.IsAvailable = false
		return
	}
	resp.Body.Close()
	ipInfo.IsAvailable = true

	// 2. Speed Test
	speedClient := &http.Client{Transport: transport, Timeout: speedTestTimeout}
	req, err = http.NewRequest("GET", speedTestURL, nil)
	if err != nil {
		return // Don't mark as unavailable, just failed the speed test
	}

	start := time.Now()
	speedResp, err := speedClient.Do(req)
	if err != nil {
		return
	}
	defer speedResp.Body.Close()

	bytes, err := io.Copy(io.Discard, speedResp.Body)
	if err != nil {
		return
	}
	duration := time.Since(start)

	if duration.Seconds() > 0 {
		// Calculate speed in MB/s
		ipInfo.DownloadSpeed = (float64(bytes) / 1024 / 1024) / duration.Seconds()
	}
}

// RunHTTPTests concurrently performs HTTP tests on a list of IPs.
// This function modifies the IPInfo structs in the provided slice directly.
func RunHTTPTests(ipInfos []*models.IPInfo, concurrency int, targetURL, speedTestURL string, timeout, speedTestTimeout time.Duration, bar *progressbar.ProgressBar) {
	var wg sync.WaitGroup
	ipChan := make(chan *models.IPInfo, len(ipInfos))

	// Start worker goroutines
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for ipInfo := range ipChan {
				testIPConnectivityAndSpeed(ipInfo, targetURL, speedTestURL, timeout, speedTestTimeout)
				bar.Add(1)
			}
		}()
	}

	// Feed IPs to the workers
	for _, ipInfo := range ipInfos {
		ipChan <- ipInfo
	}
	close(ipChan)

	// Wait for all workers to finish
	wg.Wait()
}

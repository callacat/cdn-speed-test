package tester

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/callacat/cdn-speed-test/internal/models"
)

// CheckConnectivityAndSpeed 对单个IP进行HTTP连通性和速度测试
func CheckConnectivityAndSpeed(ipInfo *models.IPInfo, targetURL, speedTestURL string, timeout, speedTestTimeout time.Duration) {
	// 创建自定义Transport
	dialer := &net.Dialer{
		Timeout:   timeout,
		KeepAlive: 30 * time.Second,
	}
	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			// 强制使用指定的IP地址
			return dialer.DialContext(ctx, network, net.JoinHostPort(ipInfo.IP.String(), "443"))
		},
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: true}, // 忽略证书验证
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   timeout,
	}

	// 1. 连通性测试
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

	// 2. 速度测试
	speedClient := &http.Client{
		Transport: transport,
		Timeout:   speedTestTimeout,
	}
	req, err = http.NewRequest("GET", speedTestURL, nil)
	if err != nil {
		return
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
		// MB/s
		ipInfo.DownloadSpeed = (float64(bytes) / 1024 / 1024) / duration.Seconds()
	}
}

// RunHTTPTests 并发对IP列表进行HTTP测试
func RunHTTPTests(ipInfos []*models.IPInfo, concurrency int, targetURL, speedTestURL string, timeout, speedTestTimeout time.Duration) {
	var wg sync.WaitGroup
	ipChan := make(chan *models.IPInfo, len(ipInfos))

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for ipInfo := range ipChan {
				CheckConnectivityAndSpeed(ipInfo, targetURL, speedTestURL, timeout, speedTestTimeout)
			}
		}()
	}

	for _, ipInfo := range ipInfos {
		ipChan <- ipInfo
	}
	close(ipChan)

	wg.Wait()
}

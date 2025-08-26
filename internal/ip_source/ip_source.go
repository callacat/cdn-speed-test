package ip_source

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/callacat/cdn-speed-test/internal/config"
)

// GetIPs 根据配置获取IP地址列表
func GetIPs(cfg *config.Config) ([]net.IP, error) {
	// 优先从本地文件读取
	if _, err := os.Stat(cfg.IPSource.LocalFiles.IPv4); err == nil {
		fmt.Println("🔍 从本地文件加载IP列表:", cfg.IPSource.LocalFiles.IPv4)
		return readIPsFromFile(cfg.IPSource.LocalFiles.IPv4)
	}

	// 如果文件不存在，尝试从API获取
	if cfg.IPSource.APIURL != "" {
		fmt.Println("🌐 从API加载IP列表:", cfg.IPSource.APIURL)
		return getIPsFromAPI(cfg.IPSource.APIURL)
	}

	return nil, fmt.Errorf("IP来源文件 %s 不存在，且未配置API URL", cfg.IPSource.LocalFiles.IPv4)
}

// readIPsFromFile 从文件中读取IP
func readIPsFromFile(path string) ([]net.IP, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var ips []net.IP
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if ip := net.ParseIP(line); ip != nil {
			ips = append(ips, ip)
		} else if _, ipNet, err := net.ParseCIDR(line); err == nil {
			// 展开CIDR
			for ip := ipNet.IP.Mask(ipNet.Mask); ipNet.Contains(ip); inc(ip) {
				// 复制IP以避免修改原始IP
				newIP := make(net.IP, len(ip))
				copy(newIP, ip)
				ips = append(ips, newIP)
			}
		}
	}
	return ips, scanner.Err()
}

// getIPsFromAPI 从API获取IP
func getIPsFromAPI(apiURL string) ([]net.IP, error) {
	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API请求失败，状态码: %d", resp.StatusCode)
	}

	var ips []net.IP
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if ip := net.ParseIP(line); ip != nil {
			ips = append(ips, ip)
		}
	}
	return ips, scanner.Err()
}

// inc 用于增加IP地址
func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

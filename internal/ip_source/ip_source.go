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

// GetIPs retrieves the list of IP addresses based on the configuration.
func GetIPs(cfg *config.Config) ([]net.IP, error) {
	// Prioritize reading from the local file.
	if _, err := os.Stat(cfg.IPSource.LocalFiles.IPv4); err == nil {
		fmt.Println("🔍 Loading IP list from local file:", cfg.IPSource.LocalFiles.IPv4)
		return readIPsFromFile(cfg.IPSource.LocalFiles.IPv4)
	}

	// If the file doesn't exist, try fetching from the API.
	if cfg.IPSource.APIURL != "" {
		fmt.Println("🌐 Loading IP list from API:", cfg.IPSource.APIURL)
		return getIPsFromAPI(cfg.IPSource.APIURL)
	}

	return nil, fmt.Errorf("IP source file %s not found and no API URL configured", cfg.IPSource.LocalFiles.IPv4)
}

// readIPsFromFile reads IPs and CIDR ranges from a file.
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
		if line == "" {
			continue
		}

		// Check if it's a CIDR range
		if _, ipNet, err := net.ParseCIDR(line); err == nil {
			// Expand CIDR
			for ip := ipNet.IP.Mask(ipNet.Mask); ipNet.Contains(ip); inc(ip) {
				newIP := make(net.IP, len(ip))
				copy(newIP, ip)
				ips = append(ips, newIP)
			}
		} else if ip := net.ParseIP(line); ip != nil {
			// It's a single IP
			ips = append(ips, ip)
		}
	}
	return ips, scanner.Err()
}

// getIPsFromAPI fetches IPs from a remote URL.
func getIPsFromAPI(apiURL string) ([]net.IP, error) {
	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status code: %d", resp.StatusCode)
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

// inc increments an IP address. Used for CIDR expansion.
func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

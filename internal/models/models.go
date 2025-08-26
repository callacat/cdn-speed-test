package models

import (
	"net"
	"time"
)

// IPInfo 包含一个IP地址及其测试结果
type IPInfo struct {
	IP            net.IP
	Latency       time.Duration // 平均延迟
	PacketLoss    float64       // 丢包率 (0.0 - 1.0)
	IsAvailable   bool          // HTTP是否可用
	DownloadSpeed float64       // 下载速度 (MB/s)
	GeoInfo       string        // 地理位置信息 (暂未实现)
}

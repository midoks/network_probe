package modules

import (
	"fmt"
	"net"
	"time"
)

// TcpingConfig 表示 TCP ping 配置
type TcpingConfig struct {
	Host    string
	Port    int
	Count   int
	Timeout time.Duration
}

// NewTcpingConfig 创建一个新的 TCP ping 配置
func NewTcpingConfig() *TcpingConfig {
	return &TcpingConfig{
		Count:   4,
		Timeout: 3 * time.Second,
	}
}

// TcpingResult 表示 TCP ping 结果
type TcpingResult struct {
	Host               string
	IP                 string
	Port               int
	Attempts           int
	SuccessfulAttempts int
	PacketLoss         float64
	MinRTT             float64
	MaxRTT             float64
	AvgRTT             float64
}

// TcpingService 表示 TCP ping 服务
type TcpingService struct{}

// NewTcpingService 创建一个新的 TCP ping 服务
func NewTcpingService() *TcpingService {
	return &TcpingService{}
}

// Tcping 执行 TCP ping 操作
func (s *TcpingService) Tcping(config *TcpingConfig) (*TcpingResult, error) {
	// 解析主机地址
	hostAddr, err := net.ResolveIPAddr("ip", config.Host)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve host: %v", err)
	}

	// 实现 TCP ping 逻辑
	var successfulAttempts int
	var totalRTT time.Duration
	var minRTT, maxRTT time.Duration

	for i := 0; i < config.Count; i++ {
		start := time.Now()

		// 尝试连接到指定端口
		addr := fmt.Sprintf("%s:%d", hostAddr.String(), config.Port)
		conn, err := net.DialTimeout("tcp", addr, config.Timeout)
		if err != nil {
			continue
		}
		conn.Close()

		// 计算 RTT
		rtt := time.Since(start)
		totalRTT += rtt

		if minRTT == 0 || rtt < minRTT {
			minRTT = rtt
		}
		if rtt > maxRTT {
			maxRTT = rtt
		}

		successfulAttempts++
	}

	// 计算丢包率
	packetLoss := float64(config.Count-successfulAttempts) / float64(config.Count) * 100

	// 计算平均 RTT
	var avgRTT float64
	if successfulAttempts > 0 {
		avgRTT = totalRTT.Seconds() * 1000 / float64(successfulAttempts)
	}

	return &TcpingResult{
		Host:               config.Host,
		IP:                 hostAddr.String(),
		Port:               config.Port,
		Attempts:           config.Count,
		SuccessfulAttempts: successfulAttempts,
		PacketLoss:         packetLoss,
		MinRTT:             minRTT.Seconds() * 1000,
		MaxRTT:             maxRTT.Seconds() * 1000,
		AvgRTT:             avgRTT,
	}, nil
}

// ScanPorts 扫描指定主机的多个端口
func (s *TcpingService) ScanPorts(host string, ports []int, timeout time.Duration) (map[int]bool, error) {
	// 解析主机地址
	hostAddr, err := net.ResolveIPAddr("ip", host)
	if err != nil {
		return nil, err
	}

	results := make(map[int]bool)

	for _, port := range ports {
		addr := fmt.Sprintf("%s:%d", hostAddr.String(), port)
		conn, err := net.DialTimeout("tcp", addr, timeout)
		if err != nil {
			results[port] = false
			continue
		}
		conn.Close()
		results[port] = true
	}

	return results, nil
}

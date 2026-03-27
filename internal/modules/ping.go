package modules

import (
	"errors"
	"fmt"
	"net"
	"os"
	"time"
)

// PingConfig 表示 ping 配置
type PingConfig struct {
	Host    string
	Count   int
	Timeout time.Duration
}

// NewPingConfig 创建一个新的 ping 配置
func NewPingConfig() *PingConfig {
	return &PingConfig{
		Count:   4,
		Timeout: 2 * time.Second,
	}
}

// PingResult 表示 ping 结果
type PingResult struct {
	Host               string
	IP                 string
	Attempts           int
	SuccessfulAttempts int
	PacketLoss         float64
	MinRTT             float64
	MaxRTT             float64
	AvgRTT             float64
}

// PingService 表示 ping 服务
type PingService struct{}

// NewPingService 创建一个新的 ping 服务
func NewPingService() *PingService {
	return &PingService{}
}

// Ping 执行 ping 操作
func (s *PingService) Ping(config *PingConfig) (*PingResult, error) {
	// 解析主机地址
	hostAddr, err := net.ResolveIPAddr("ip", config.Host)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve host: %v", err)
	}

	// 检查是否为 root 用户（在类 Unix 系统中）
	if os.Geteuid() != 0 {
		return nil, errors.New("ping requires root privileges. Please run with sudo or as root")
	}

	// 实现 ping 逻辑
	var successfulAttempts int
	var totalRTT time.Duration
	var minRTT, maxRTT time.Duration

	for i := 0; i < config.Count; i++ {
		start := time.Now()

		// 创建 ICMP 连接
		conn, err := net.DialTimeout("ip4:icmp", hostAddr.String(), config.Timeout)
		if err != nil {
			continue
		}
		defer conn.Close()

		// 发送 ICMP 回显请求
		echoRequest := []byte{
			8, 0, 0, 0, // 类型、代码、校验和
			0, 0, 0, 0, // 标识符、序列号
		}

		_, err = conn.Write(echoRequest)
		if err != nil {
			continue
		}

		// 接收响应
		conn.SetReadDeadline(time.Now().Add(config.Timeout))
		buf := make([]byte, 1024)
		_, err = conn.Read(buf)
		if err != nil {
			continue
		}

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

	return &PingResult{
		Host:               config.Host,
		IP:                 hostAddr.String(),
		Attempts:           config.Count,
		SuccessfulAttempts: successfulAttempts,
		PacketLoss:         packetLoss,
		MinRTT:             minRTT.Seconds() * 1000,
		MaxRTT:             maxRTT.Seconds() * 1000,
		AvgRTT:             avgRTT,
	}, nil
}

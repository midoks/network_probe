package modules

import (
	"fmt"
	"net"
	"time"
)

// TracerouteConfig 表示 traceroute 配置
type TracerouteConfig struct {
	Host     string
	MaxHops  int
	Protocol string
}

// NewTracerouteConfig 创建一个新的 traceroute 配置
func NewTracerouteConfig() *TracerouteConfig {
	return &TracerouteConfig{
		MaxHops:  30,
		Protocol: "icmp",
	}
}

// TracerouteHop 表示 traceroute 中的一个跳点
type TracerouteHop struct {
	Hop      int
	IP       string
	Hostname string
	RTT      float64
	Error    string
}

// TracerouteResult 表示 traceroute 结果
type TracerouteResult struct {
	Host string
	Hops []TracerouteHop
}

// TracerouteService 表示 traceroute 服务
type TracerouteService struct{}

// NewTracerouteService 创建一个新的 traceroute 服务
func NewTracerouteService() *TracerouteService {
	return &TracerouteService{}
}

// Traceroute 执行 traceroute 操作
func (s *TracerouteService) Traceroute(config *TracerouteConfig) (*TracerouteResult, error) {
	// 解析主机地址
	hostAddr, err := net.ResolveIPAddr("ip", config.Host)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve host: %v", err)
	}

	result := &TracerouteResult{
		Host: config.Host,
		Hops: make([]TracerouteHop, 0, config.MaxHops),
	}

	// 实现 traceroute 逻辑
	for ttl := 1; ttl <= config.MaxHops; ttl++ {
		hop := TracerouteHop{
			Hop: ttl,
		}

		// 创建 ICMP 连接
		conn, err := net.DialTimeout("ip4:icmp", hostAddr.String(), 2*time.Second)
		if err != nil {
			hop.Error = err.Error()
			result.Hops = append(result.Hops, hop)
			continue
		}

		// 发送 ICMP 回显请求
		echoRequest := []byte{
			8, 0, 0, 0, // 类型、代码、校验和
			0, 0, 0, 0, // 标识符、序列号
		}

		start := time.Now()
		_, err = conn.Write(echoRequest)
		if err != nil {
			conn.Close()
			hop.Error = err.Error()
			result.Hops = append(result.Hops, hop)
			continue
		}

		// 接收响应
		conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		rtt := time.Since(start)
		conn.Close()

		if err != nil {
			hop.Error = err.Error()
			result.Hops = append(result.Hops, hop)
			continue
		}

		if n > 0 {
			hop.IP = hostAddr.String()
			hop.RTT = rtt.Seconds() * 1000

			// 尝试解析主机名
			hostnames, err := net.LookupAddr(hop.IP)
			if err == nil && len(hostnames) > 0 {
				hop.Hostname = hostnames[0]
			}

			// 简单实现，直接添加跳点
		}

		// 添加跳点到结果
		result.Hops = append(result.Hops, hop)
	}

	return result, nil
}

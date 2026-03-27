package modules

import (
	"fmt"
	"net"
	"syscall"
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

	// 使用 UDP 协议实现 traceroute，这样普通用户也能执行
	return s.tracerouteUsingUDP(config, hostAddr)
}

// tracerouteUsingUDP 使用 UDP 协议执行 traceroute 操作
func (s *TracerouteService) tracerouteUsingUDP(config *TracerouteConfig, hostAddr *net.IPAddr) (*TracerouteResult, error) {
	result := &TracerouteResult{
		Host: config.Host,
		Hops: make([]TracerouteHop, 0, config.MaxHops),
	}

	// 创建 ICMP 监听套接字
	icmpConn, err := net.ListenPacket("ip4:icmp", "0.0.0.0")
	if err != nil {
		return nil, fmt.Errorf("failed to listen for ICMP: %v", err)
	}
	defer icmpConn.Close()

	// 为每个 TTL 值发送数据包
	for ttl := 1; ttl <= config.MaxHops; ttl++ {
		hop := TracerouteHop{
			Hop: ttl,
		}

		// 创建 UDP 连接
		udpAddr := &net.UDPAddr{
			IP:   hostAddr.IP,
			Port: 33434 + ttl, // 使用标准 traceroute 端口范围
		}

		// 创建 UDP 套接字
		udpConn, err := net.DialUDP("udp", nil, udpAddr)
		if err != nil {
			hop.Error = err.Error()
			result.Hops = append(result.Hops, hop)
			continue
		}

		// 设置 TTL
		rawConn, err := udpConn.SyscallConn()
		if err == nil {
			rawConn.Control(func(fd uintptr) {
				syscall.SetsockoptInt(int(fd), syscall.IPPROTO_IP, syscall.IP_TTL, ttl)
			})
		}

		// 发送数据
		start := time.Now()
		_, err = udpConn.Write([]byte("traceroute probe"))
		udpConn.Close()

		if err != nil {
			hop.Error = err.Error()
			result.Hops = append(result.Hops, hop)
			continue
		}

		// 接收 ICMP 响应
		icmpConn.SetReadDeadline(time.Now().Add(2 * time.Second))
		buf := make([]byte, 1024)
		n, addr, err := icmpConn.ReadFrom(buf)

		if err != nil {
			hop.Error = "Request timed out"
			result.Hops = append(result.Hops, hop)
			continue
		}

		if n > 0 {
			// 计算 RTT
			rtt := time.Since(start).Seconds() * 1000

			// 获取响应地址
			respAddr, ok := addr.(*net.IPAddr)
			if ok {
				hop.IP = respAddr.String()
				hop.RTT = rtt

				// 尝试解析主机名
				hostnames, err := net.LookupAddr(hop.IP)
				if err == nil && len(hostnames) > 0 {
					hop.Hostname = hostnames[0]
				}

				// 检查是否到达目标
				if respAddr.IP.Equal(hostAddr.IP) {
					result.Hops = append(result.Hops, hop)
					break
				}
			}
		}

		// 添加跳点到结果
		result.Hops = append(result.Hops, hop)
	}

	return result, nil
}

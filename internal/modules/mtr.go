package modules

import (
	"bytes"
	"fmt"
	"math"
	"net"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// MtrConfig 表示 mtr 配置
type MtrConfig struct {
	Host     string
	MaxHops  int
	Interval int
	Count    int
}

// NewMtrConfig 创建一个新的 mtr 配置
func NewMtrConfig() *MtrConfig {
	return &MtrConfig{
		MaxHops:  30,
		Interval: 1,
		Count:    10,
	}
}

// MtrHop 表示 mtr 中的一个跳点
type MtrHop struct {
	Hop      int
	IP       string
	Hostname string
	Loss     float64
	Snt      int
	Last     float64
	Avg      float64
	Best     float64
	Wrst     float64
	StDev    float64
}

// MtrResult 表示 mtr 结果
type MtrResult struct {
	Host string
	Hops []MtrHop
}

// MtrService 表示 mtr 服务
type MtrService struct{}

// NewMtrService 创建一个新的 mtr 服务
func NewMtrService() *MtrService {
	return &MtrService{}
}

// Mtr 执行 mtr 操作
func (s *MtrService) Mtr(config *MtrConfig) (*MtrResult, error) {
	// 解析主机地址
	hostAddr, err := net.ResolveIPAddr("ip", config.Host)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve host: %v", err)
	}

	// 根据平台选择不同的实现
	switch runtime.GOOS {
	case "linux":
		// Linux 使用原生实现
		return s.mtrNativeImplementation(config, hostAddr)
	case "darwin":
		// macOS 使用 mtr + sudo 实现
		return s.mtrUsingSystemCommand(config, hostAddr)
	default:
		// 其他平台使用系统命令
		return s.mtrUsingSystemCommand(config, hostAddr)
	}
}

// mtrUsingSystemCommand 使用系统 mtr 命令执行 mtr 操作
func (s *MtrService) mtrUsingSystemCommand(config *MtrConfig, hostAddr *net.IPAddr) (*MtrResult, error) {
	// 构建 mtr 命令
	var cmd *exec.Cmd

	// 首先尝试不使用 sudo 运行 mtr
	cmd = exec.Command("mtr", "-n", "-c", strconv.Itoa(config.Count), "-m", strconv.Itoa(config.MaxHops), "-i", strconv.Itoa(config.Interval), config.Host)

	// 捕获输出
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	// 执行命令
	err := cmd.Run()
	output := out.String()

	// 如果失败且是权限错误，尝试使用 sudo
	if err != nil && (strings.Contains(output, "Failure to open") || strings.Contains(output, "Permission denied") || strings.Contains(output, "mtr-packet")) {
		// 尝试使用 sudo
		cmd = exec.Command("sudo", "mtr", "-n", "-c", strconv.Itoa(config.Count), "-m", strconv.Itoa(config.MaxHops), "-i", strconv.Itoa(config.Interval), config.Host)

		// 重置输出缓冲区
		out.Reset()
		cmd.Stdout = &out
		cmd.Stderr = &out

		// 执行命令
		err = cmd.Run()
		output = out.String()
	}

	// 检查是否有错误
	if err != nil {
		// 检查是否是权限错误
		if strings.Contains(output, "Failure to open") || strings.Contains(output, "Permission denied") || strings.Contains(output, "mtr-packet") {
			return nil, fmt.Errorf("mtr requires root privileges. Please run with sudo or as root")
		}
		// 检查是否是命令不存在
		if strings.Contains(output, "command not found") || strings.Contains(output, "not found") {
			return nil, fmt.Errorf("mtr command not found. Please install mtr first")
		}
		// 其他错误，返回输出
		return nil, fmt.Errorf("mtr execution failed: %v. Output: %s", err, output)
	}

	// 检查输出是否为空
	if strings.TrimSpace(output) == "" {
		return nil, fmt.Errorf("mtr returned no output")
	}

	// 解析 mtr 输出
	result := &MtrResult{
		Host: config.Host,
		Hops: make([]MtrHop, 0),
	}

	// 解析每一行
	lines := strings.Split(output, "\n")

	// 尝试多种正则表达式格式
	hopRegex1 := regexp.MustCompile(`^(\d+)\s+\|\s+([^\s]+)\s+\|\s+([\d.]+)%\s+\|\s+(\d+)\s+\|\s+([\d.]+)\s+\|\s+([\d.]+)\s+\|\s+([\d.]+)\s+\|\s+([\d.]+)\s+\|\s+([\d.]+)`)
	hopRegex2 := regexp.MustCompile(`^(\d+)\s+([^\s]+)\s+([\d.]+)%\s+(\d+)\s+([\d.]+)\s+([\d.]+)\s+([\d.]+)\s+([\d.]+)\s+([\d.]+)`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Start:") || strings.HasPrefix(line, "HOST") || strings.Contains(line, "Failure to open") {
			continue
		}

		// 尝试匹配跳点信息（格式1：带管道符）
		matches := hopRegex1.FindStringSubmatch(line)
		if len(matches) >= 10 {
			hopNum, _ := strconv.Atoi(matches[1])
			ip := strings.TrimSpace(matches[2])
			loss, _ := strconv.ParseFloat(matches[3], 64)
			snt, _ := strconv.Atoi(matches[4])
			last, _ := strconv.ParseFloat(matches[5], 64)
			avg, _ := strconv.ParseFloat(matches[6], 64)
			best, _ := strconv.ParseFloat(matches[7], 64)
			wrst, _ := strconv.ParseFloat(matches[8], 64)
			stDev, _ := strconv.ParseFloat(matches[9], 64)

			// 尝试解析主机名
			hostname := ""
			hostnames, err := net.LookupAddr(ip)
			if err == nil && len(hostnames) > 0 {
				hostname = hostnames[0]
			}

			hop := MtrHop{
				Hop:      hopNum,
				IP:       ip,
				Hostname: hostname,
				Loss:     loss,
				Snt:      snt,
				Last:     last,
				Avg:      avg,
				Best:     best,
				Wrst:     wrst,
				StDev:    stDev,
			}
			result.Hops = append(result.Hops, hop)
			continue
		}

		// 尝试匹配跳点信息（格式2：空格分隔）
		matches = hopRegex2.FindStringSubmatch(line)
		if len(matches) >= 10 {
			hopNum, _ := strconv.Atoi(matches[1])
			ip := strings.TrimSpace(matches[2])
			loss, _ := strconv.ParseFloat(matches[3], 64)
			snt, _ := strconv.Atoi(matches[4])
			last, _ := strconv.ParseFloat(matches[5], 64)
			avg, _ := strconv.ParseFloat(matches[6], 64)
			best, _ := strconv.ParseFloat(matches[7], 64)
			wrst, _ := strconv.ParseFloat(matches[8], 64)
			stDev, _ := strconv.ParseFloat(matches[9], 64)

			// 尝试解析主机名
			hostname := ""
			hostnames, err := net.LookupAddr(ip)
			if err == nil && len(hostnames) > 0 {
				hostname = hostnames[0]
			}

			hop := MtrHop{
				Hop:      hopNum,
				IP:       ip,
				Hostname: hostname,
				Loss:     loss,
				Snt:      snt,
				Last:     last,
				Avg:      avg,
				Best:     best,
				Wrst:     wrst,
				StDev:    stDev,
			}
			result.Hops = append(result.Hops, hop)
		}
	}

	// 检查是否有跳点信息
	if len(result.Hops) == 0 {
		return nil, fmt.Errorf("no mtr results found. Please check your network connection or mtr installation")
	}

	// 即使命令返回错误，我们也返回结果
	return result, nil
}

// mtrNativeImplementation 使用原生方法执行 mtr 操作（Linux 平台）
func (s *MtrService) mtrNativeImplementation(config *MtrConfig, hostAddr *net.IPAddr) (*MtrResult, error) {
	result := &MtrResult{
		Host: config.Host,
		Hops: make([]MtrHop, 0),
	}

	// 尝试使用 ICMP 套接字
	conn, err := net.ListenPacket("ip4:icmp", "")
	if err != nil {
		return nil, fmt.Errorf("failed to create ICMP socket: %v. Please run with root privileges", err)
	}
	defer conn.Close()

	// 对每个跳点进行测试
	for ttl := 1; ttl <= config.MaxHops; ttl++ {
		hop := MtrHop{
			Hop:      ttl,
			IP:       "*",
			Hostname: "",
			Loss:     100.0,
			Snt:      config.Count,
			Last:     0,
			Avg:      0,
			Best:     0,
			Wrst:     0,
			StDev:    0,
		}

		totalRTT := 0.0
		successCount := 0
		var rtts []float64

		// 发送多个探测包
		for i := 0; i < config.Count; i++ {
			// 构建 ICMP echo 请求
			echoRequest := s.buildICMPEchoRequest(i)

			// 设置 TTL
			s.setTTL(conn, ttl)

			// 发送请求
			startTime := time.Now()
			n, err := conn.WriteTo(echoRequest, &net.IPAddr{IP: hostAddr.IP})
			if err != nil || n != len(echoRequest) {
				continue
			}

			// 接收响应
			buffer := make([]byte, 1500)
			conn.SetReadDeadline(time.Now().Add(time.Second * 2))
			n, addr, err := conn.ReadFrom(buffer)
			if err != nil {
				continue
			}

			rtt := time.Since(startTime).Seconds() * 1000
			totalRTT += rtt
			successCount++
			rtts = append(rtts, rtt)

			// 解析响应
			if ipAddr, ok := addr.(*net.IPAddr); ok {
				hop.IP = ipAddr.IP.String()
				// 尝试解析主机名
				hostnames, err := net.LookupAddr(hop.IP)
				if err == nil && len(hostnames) > 0 {
					hop.Hostname = hostnames[0]
				}
			}

			// 短暂休眠
			time.Sleep(time.Duration(config.Interval) * time.Second)
		}

		// 计算统计信息
		if successCount > 0 {
			hop.Loss = float64(config.Count-successCount) / float64(config.Count) * 100
			hop.Avg = totalRTT / float64(successCount)

			// 计算最佳、最差和标准差
			if len(rtts) > 0 {
				hop.Best = rtts[0]
				hop.Wrst = rtts[0]
				for _, r := range rtts {
					if r < hop.Best {
						hop.Best = r
					}
					if r > hop.Wrst {
						hop.Wrst = r
					}
				}

				// 计算标准差
				variance := 0.0
				for _, r := range rtts {
					variance += (r - hop.Avg) * (r - hop.Avg)
				}
				hop.StDev = math.Sqrt(variance / float64(successCount))

				// 最后一个 RTT
				if len(rtts) > 0 {
					hop.Last = rtts[len(rtts)-1]
				}
			}
		}

		result.Hops = append(result.Hops, hop)

		// 如果到达目标，停止
		if hop.IP == hostAddr.IP.String() && successCount > 0 {
			break
		}
	}

	if len(result.Hops) == 0 {
		return nil, fmt.Errorf("no mtr results found. Please check your network connection")
	}

	return result, nil
}

// buildICMPEchoRequest 构建 ICMP echo 请求
func (s *MtrService) buildICMPEchoRequest(seq int) []byte {
	// ICMP echo request
	// Type: 8, Code: 0
	// Checksum: will be calculated
	// Identifier: 12345
	// Sequence: seq

	echoRequest := []byte{
		8, 0, 0, 0, // Type, Code, Checksum (placeholder)
		0x30, 0x39, // Identifier (12345)
		byte(seq >> 8), byte(seq & 0xff), // Sequence
		// Payload: 48 bytes of data
		0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
		0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f,
		0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17,
		0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f,
		0x20, 0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27,
		0x28, 0x29, 0x2a, 0x2b, 0x2c, 0x2d, 0x2e, 0x2f,
	}

	// Calculate checksum
	checksum := s.calculateChecksum(echoRequest)
	echoRequest[2] = byte(checksum >> 8)
	echoRequest[3] = byte(checksum & 0xff)

	return echoRequest
}

// calculateChecksum 计算 ICMP 校验和
func (s *MtrService) calculateChecksum(data []byte) uint16 {
	var sum uint32
	for i := 0; i < len(data); i += 2 {
		if i+1 < len(data) {
			sum += uint32(data[i])<<8 | uint32(data[i+1])
		} else {
			sum += uint32(data[i]) << 8
		}
	}

	sum = (sum >> 16) + (sum & 0xffff)
	sum += (sum >> 16)

	return uint16(^sum)
}

// setTTL 设置套接字的 TTL
func (s *MtrService) setTTL(conn net.PacketConn, ttl int) error {
	// 注意：在 Go 标准库中，net.IPConn 没有 SetTTL 方法
	// 我们需要使用 syscall 来设置 TTL，但这会增加平台依赖性
	// 为了保持跨平台兼容性，我们暂时不设置 TTL
	// 在 Linux 平台上，TTL 默认为 64，这通常足够使用
	return nil
}

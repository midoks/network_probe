package modules

import (
	"bytes"
	"fmt"
	"net"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
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

	// 使用系统 mtr 命令
	return s.mtrUsingSystemCommand(config, hostAddr)
}

// mtrUsingSystemCommand 使用系统 mtr 命令执行 mtr 操作
func (s *MtrService) mtrUsingSystemCommand(config *MtrConfig, hostAddr *net.IPAddr) (*MtrResult, error) {
	// 构建 mtr 命令
	cmd := exec.Command("mtr", "-n", "-c", strconv.Itoa(config.Count), "-m", strconv.Itoa(config.MaxHops), "-i", strconv.Itoa(config.Interval), config.Host)

	// 捕获输出
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	// 执行命令
	err := cmd.Run()
	output := out.String()

	// 检查是否有错误
	if err != nil && strings.Contains(output, "Failure to open") {
		return nil, fmt.Errorf("mtr requires root privileges. Please run with sudo or as root")
	}

	// 解析 mtr 输出
	result := &MtrResult{
		Host: config.Host,
		Hops: make([]MtrHop, 0),
	}

	// 解析每一行
	lines := strings.Split(output, "\n")
	hopRegex := regexp.MustCompile(`^(\d+)\s+([^\s]+)\s+([\d.]+)%\s+(\d+)\s+([\d.]+)\s+([\d.]+)\s+([\d.]+)\s+([\d.]+)\s+([\d.]+)`)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "Start:") || strings.HasPrefix(line, "HOST") || strings.Contains(line, "Failure to open") {
			continue
		}

		// 尝试匹配跳点信息
		matches := hopRegex.FindStringSubmatch(line)
		if len(matches) >= 10 {
			hopNum, _ := strconv.Atoi(matches[1])
			ip := matches[2]
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
	if len(result.Hops) == 0 && err == nil {
		return nil, fmt.Errorf("no mtr results found. Please check your network connection")
	}

	// 即使命令返回错误，我们也返回结果
	return result, nil
}

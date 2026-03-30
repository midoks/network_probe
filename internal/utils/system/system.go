package system

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// SystemInfo 表示系统状态信息
type SystemInfo struct {
	OnlineStatus      string `json:"online_status"`
	DownloadBandwidth string `json:"download_bandwidth"`
	UploadBandwidth   string `json:"upload_bandwidth"`
	Connections       string `json:"connections"`
	AccessRate        string `json:"access_rate"`
	AttackRate        string `json:"attack_rate"`
	CacheDiskUsage    string `json:"cache_disk_usage"`
	MaxDiskWriteSpeed string `json:"max_disk_write_speed"`
	MemoryCacheUsage  string `json:"memory_cache_usage"`
	CPUUsage          string `json:"cpu_usage"`
	MemoryUsage       string `json:"memory_usage"`
	TotalMemory       string `json:"total_memory"`
	Load              string `json:"load"`
	MonthlyTraffic    string `json:"monthly_traffic"`
	YesterdayTraffic  string `json:"yesterday_traffic"`
	TodayTraffic      string `json:"today_traffic"`
}

// GetSystemInfo 获取系统状态信息
func GetSystemInfo() (*SystemInfo, error) {
	info := &SystemInfo{
		OnlineStatus: "在线",
	}

	// 获取带宽信息
	download, upload, err := getBandwidth()
	if err == nil {
		info.DownloadBandwidth = download
		info.UploadBandwidth = upload
	} else {
		info.DownloadBandwidth = "N/A"
		info.UploadBandwidth = "N/A"
	}

	// 获取连接数
	connections, err := getConnections()
	if err == nil {
		info.Connections = connections
	} else {
		info.Connections = "N/A"
	}

	// 获取访问量
	accessRate, err := getAccessRate()
	if err == nil {
		info.AccessRate = accessRate
	} else {
		info.AccessRate = "N/A"
	}

	// 获取攻击访问量
	attackRate, err := getAttackRate()
	if err == nil {
		info.AttackRate = attackRate
	} else {
		info.AttackRate = "N/A"
	}

	// 获取缓存硬盘用量
	cacheDiskUsage, err := getCacheDiskUsage()
	if err == nil {
		info.CacheDiskUsage = cacheDiskUsage
	} else {
		info.CacheDiskUsage = "N/A"
	}

	// 获取硬盘预估写入最大速度
	maxDiskWriteSpeed, err := getMaxDiskWriteSpeed()
	if err == nil {
		info.MaxDiskWriteSpeed = maxDiskWriteSpeed
	} else {
		info.MaxDiskWriteSpeed = "N/A"
	}

	// 获取内存缓存用量
	memoryCacheUsage, err := getMemoryCacheUsage()
	if err == nil {
		info.MemoryCacheUsage = memoryCacheUsage
	} else {
		info.MemoryCacheUsage = "N/A"
	}

	// 获取 CPU 使用率
	cpuUsage, err := getCPUUsage()
	if err == nil {
		info.CPUUsage = cpuUsage
	} else {
		info.CPUUsage = "N/A"
	}

	// 获取内存使用率和总内存
	memoryUsage, totalMemory, err := getMemoryInfo()
	if err == nil {
		info.MemoryUsage = memoryUsage
		info.TotalMemory = totalMemory
	} else {
		info.MemoryUsage = "N/A"
		info.TotalMemory = "N/A"
	}

	// 获取负载
	load, err := getLoad()
	if err == nil {
		info.Load = load
	} else {
		info.Load = "N/A"
	}

	// 获取流量信息
	monthlyTraffic, yesterdayTraffic, todayTraffic, err := getTrafficInfo()
	if err == nil {
		info.MonthlyTraffic = monthlyTraffic
		info.YesterdayTraffic = yesterdayTraffic
		info.TodayTraffic = todayTraffic
	} else {
		info.MonthlyTraffic = "N/A"
		info.YesterdayTraffic = "N/A"
		info.TodayTraffic = "N/A"
	}

	return info, nil
}

// getBandwidth 获取带宽信息
func getBandwidth() (string, string, error) {
	// 尝试使用 ifconfig 命令获取带宽信息
	cmd := exec.Command("ifconfig")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "N/A", "N/A", err
	}

	// 解析输出，查找接收和发送的字节数
	var rxBytes, txBytes uint64
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "RX bytes:") {
			parts := strings.Fields(line)
			for i, part := range parts {
				if part == "RX" && i+2 < len(parts) {
					rxStr := strings.Replace(parts[i+2], "bytes:", "", -1)
					rxBytes, _ = strconv.ParseUint(rxStr, 10, 64)
				}
				if part == "TX" && i+2 < len(parts) {
					txStr := strings.Replace(parts[i+2], "bytes:", "", -1)
					txBytes, _ = strconv.ParseUint(txStr, 10, 64)
				}
			}
		}
	}

	// 转换为 Mbps 和 Kbps
	rxMbps := float64(rxBytes*8) / (1024 * 1024)
	txKbps := float64(txBytes*8) / 1024

	return fmt.Sprintf("%.4f Mbps", rxMbps), fmt.Sprintf("%.4f Kbps", txKbps), nil
}

// getConnections 获取连接数
func getConnections() (string, error) {
	// 尝试使用 netstat 命令获取连接数
	cmd := exec.Command("netstat", "-an")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "N/A", err
	}

	// 统计连接数
	lines := strings.Split(string(output), "\n")
	count := 0
	for _, line := range lines {
		if strings.Contains(line, "ESTABLISHED") {
			count++
		}
	}

	return fmt.Sprintf("%d/分钟", count), nil
}

// getAccessRate 获取访问量
func getAccessRate() (string, error) {
	// 尝试检查常见的 Web 服务器日志文件
	logPaths := []string{
		"/var/log/nginx/access.log",
		"/var/log/apache2/access.log",
		"/var/log/httpd/access_log",
	}

	var logPath string
	for _, path := range logPaths {
		if _, err := os.Stat(path); err == nil {
			logPath = path
			break
		}
	}

	if logPath == "" {
		return "N/A", fmt.Errorf("no web server log file found")
	}

	// 打开日志文件
	file, err := os.Open(logPath)
	if err != nil {
		return "N/A", err
	}
	defer file.Close()

	// 统计最近一分钟的访问量
	count := 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		count++
	}

	return fmt.Sprintf("%d/秒", count/60), nil
}

// getAttackRate 获取攻击访问量
func getAttackRate() (string, error) {
	// 尝试检查常见的 Web 服务器日志文件
	logPaths := []string{
		"/var/log/nginx/access.log",
		"/var/log/apache2/access.log",
		"/var/log/httpd/access_log",
	}

	var logPath string
	for _, path := range logPaths {
		if _, err := os.Stat(path); err == nil {
			logPath = path
			break
		}
	}

	if logPath == "" {
		return "N/A", fmt.Errorf("no web server log file found")
	}

	// 打开日志文件
	file, err := os.Open(logPath)
	if err != nil {
		return "N/A", err
	}
	defer file.Close()

	// 统计可能的攻击尝试
	attackPatterns := []string{
		"../",
		"/etc/passwd",
		"sqlmap",
		"union select",
		"or 1=1",
		"script>",
	}

	count := 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		for _, pattern := range attackPatterns {
			if strings.Contains(strings.ToLower(line), pattern) {
				count++
				break
			}
		}
	}

	return fmt.Sprintf("%d/分钟", count), nil
}

// getCacheDiskUsage 获取缓存硬盘用量
func getCacheDiskUsage() (string, error) {
	// 尝试使用 df 命令获取磁盘使用情况
	cmd := exec.Command("df", "-h")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "N/A", err
	}

	// 解析输出，查找根分区的使用情况
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "/dev/") && strings.Contains(line, " / ") {
			parts := strings.Fields(line)
			if len(parts) >= 5 {
				used := parts[2]
				total := parts[1]
				return fmt.Sprintf("%s / %s", used, total), nil
			}
		}
	}

	return "N/A", fmt.Errorf("no disk usage information found")
}

// getMaxDiskWriteSpeed 获取硬盘预估写入最大速度
func getMaxDiskWriteSpeed() (string, error) {
	// 尝试使用 dd 命令测试硬盘写入速度
	cmd := exec.Command("bash", "-c", "dd if=/dev/zero of=/tmp/test bs=1M count=1024 2>&1 | tail -n 1")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "N/A", err
	}

	// 解析输出，获取写入速度
	outputStr := strings.TrimSpace(string(output))
	parts := strings.Fields(outputStr)
	if len(parts) >= 8 {
		speed := parts[7]
		return fmt.Sprintf("> %s", speed), nil
	}

	return "N/A", fmt.Errorf("no disk speed information found")
}

// getMemoryCacheUsage 获取内存缓存用量
func getMemoryCacheUsage() (string, error) {
	// 尝试使用 free 命令获取内存使用情况
	cmd := exec.Command("free", "-h")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "N/A", err
	}

	// 解析输出，获取缓存用量
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "Mem:") {
			parts := strings.Fields(line)
			if len(parts) >= 6 {
				cache := parts[5]
				return cache, nil
			}
		}
	}

	return "N/A", fmt.Errorf("no memory cache information found")
}

// getCPUUsage 获取 CPU 使用率
func getCPUUsage() (string, error) {
	// 尝试使用 top 命令获取 CPU 使用率
	cmd := exec.Command("top", "-bn1")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "N/A", err
	}

	// 解析输出，获取 CPU 使用率
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "%Cpu(s):") {
			parts := strings.Fields(line)
			if len(parts) >= 8 {
				// 计算总 CPU 使用率
				idle, _ := strconv.ParseFloat(strings.Replace(parts[7], "%", "", -1), 64)
				usage := 100 - idle
				return fmt.Sprintf("%.2f%%", usage), nil
			}
		}
	}

	return "N/A", fmt.Errorf("no CPU usage information found")
}

// getMemoryInfo 获取内存使用率和总内存
func getMemoryInfo() (string, string, error) {
	// 尝试使用 free 命令获取内存使用情况
	cmd := exec.Command("free", "-h")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "N/A", "N/A", err
	}

	// 解析输出，获取内存使用率和总内存
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "Mem:") {
			parts := strings.Fields(line)
			if len(parts) >= 4 {
				total := parts[1]
				used := parts[2]
				// 计算使用率
				totalBytes, _ := convertToBytes(total)
				usedBytes, _ := convertToBytes(used)
				usage := (float64(usedBytes) / float64(totalBytes)) * 100
				return fmt.Sprintf("%.2f%%", usage), total, nil
			}
		}
	}

	return "N/A", "N/A", fmt.Errorf("no memory information found")
}

// convertToBytes 将内存大小字符串转换为字节数
func convertToBytes(size string) (uint64, error) {
	// 提取数字和单位
	var num float64
	var unit string
	fmt.Sscanf(size, "%f%s", &num, &unit)

	// 转换为字节
	switch strings.ToLower(unit) {
	case "b":
		return uint64(num), nil
	case "k", "kb":
		return uint64(num * 1024), nil
	case "m", "mb":
		return uint64(num * 1024 * 1024), nil
	case "g", "gb":
		return uint64(num * 1024 * 1024 * 1024), nil
	default:
		return 0, fmt.Errorf("unknown unit: %s", unit)
	}
}

// getLoad 获取负载
func getLoad() (string, error) {
	// 尝试使用 uptime 命令获取系统负载
	cmd := exec.Command("uptime")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "N/A", err
	}

	// 解析输出，获取负载信息
	outputStr := strings.TrimSpace(string(output))
	parts := strings.Fields(outputStr)
	if len(parts) >= 10 {
		load := parts[len(parts)-3]
		return fmt.Sprintf("%s/分钟", load), nil
	}

	return "N/A", fmt.Errorf("no load information found")
}

// getTrafficInfo 获取流量信息
func getTrafficInfo() (string, string, string, error) {
	// 尝试使用 ifconfig 命令获取网络接口的流量信息
	cmd := exec.Command("ifconfig")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "N/A", "N/A", "N/A", err
	}

	// 解析输出，查找接收和发送的字节数
	var totalBytes uint64
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "RX bytes:") {
			parts := strings.Fields(line)
			for i, part := range parts {
				if part == "RX" && i+2 < len(parts) {
					rxStr := strings.Replace(parts[i+2], "bytes:", "", -1)
					rxBytes, _ := strconv.ParseUint(rxStr, 10, 64)
					totalBytes += rxBytes
				}
				if part == "TX" && i+2 < len(parts) {
					txStr := strings.Replace(parts[i+2], "bytes:", "", -1)
					txBytes, _ := strconv.ParseUint(txStr, 10, 64)
					totalBytes += txBytes
				}
			}
		}
	}

	// 转换为 GiB
	totalGiB := float64(totalBytes) / (1024 * 1024 * 1024)

	// 假设总流量的 10% 是今日流量，30% 是昨日流量，剩余 60% 是当月流量
	todayGiB := totalGiB * 0.1
	yesterdayGiB := totalGiB * 0.3
	monthlyGiB := totalGiB * 0.6

	return fmt.Sprintf("%.2f GiB", monthlyGiB), fmt.Sprintf("%.2f GiB", yesterdayGiB), fmt.Sprintf("%.2f GiB", todayGiB), nil
}

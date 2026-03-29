package cli

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"network-probe/internal/api"
	"network-probe/internal/config"
	"network-probe/internal/modules"
	"network-probe/internal/utils/report"
	"network-probe/internal/utils/system"
	"network-probe/internal/version"
)

// Cli 表示命令行参数
type Cli struct {
	Command         string
	LogLevel        string
	Host            string
	Count           int
	Timeout         int
	Port            int
	URL             string
	Method          string
	FollowRedirects bool
	MaxHops         int
	Range           string
	ServiceName     string
	ReportEndpoints []string
}

// Run 运行 CLI
func Run() error {
	cli := &Cli{}

	var rootCmd = &cobra.Command{
		Use:   "network_probe",
		Short: "Network probe tool",
		Long:  "A comprehensive network probe tool with ping, traceroute, mtr, and more",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) > 0 {
				cli.Host = args[0]
			}
			if cli.Command == "" {
				// 没有提供命令，显示帮助信息
				cmd.Help()
				os.Exit(0)
			}
			switch cli.Command {
			case "ping":
				if err := handlePing(cli); err != nil {
					os.Exit(1)
				}
			case "tcping":
				if err := handleTcping(cli); err != nil {
					os.Exit(1)
				}
			case "website":
				if err := handleWebsite(cli); err != nil {
					os.Exit(1)
				}
			case "traceroute":
				if err := handleTraceroute(cli); err != nil {
					os.Exit(1)
				}
			case "dns":
				if err := handleDNS(cli); err != nil {
					os.Exit(1)
				}
			case "mtr":
				if err := handleMtr(cli); err != nil {
					os.Exit(1)
				}
			case "port-scan":
				if err := handlePortScan(cli); err != nil {
					os.Exit(1)
				}
			case "install":
				if err := handleInstall(cli); err != nil {
					os.Exit(1)
				}
			case "uninstall":
				if err := handleUninstall(cli); err != nil {
					os.Exit(1)
				}
			case "server":
				if err := handleServer(cli); err != nil {
					os.Exit(1)
				}
			case "version":
				fmt.Printf("network_probe version %s\n", version.Version)
			case "sysinfo":
				if err := handleSysinfo(cli); err != nil {
					os.Exit(1)
				}
			case "gc":
				if err := handleGC(cli); err != nil {
					os.Exit(1)
				}
			default:
				fmt.Printf("Unknown command: %s\n", cli.Command)
				os.Exit(1)
			}
		},
	}

	rootCmd.PersistentFlags().StringVarP(&cli.LogLevel, "log-level", "l", "info", "Log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().IntVarP(&cli.Count, "count", "c", 4, "Number of packets to send")
	rootCmd.PersistentFlags().IntVarP(&cli.Timeout, "timeout", "t", 1, "Timeout in seconds")
	rootCmd.PersistentFlags().IntVarP(&cli.MaxHops, "max-hops", "m", 30, "Maximum number of hops")

	// Ping 命令
	var pingCmd = &cobra.Command{
		Use:   "ping [host]",
		Short: "Ping a host",
		Run: func(cmd *cobra.Command, args []string) {
			cli.Command = "ping"
			if len(args) > 0 {
				cli.Host = args[0]
			}
			if err := handlePing(cli); err != nil {
				os.Exit(1)
			}
		},
	}

	// TCPing 命令
	var tcpingCmd = &cobra.Command{
		Use:   "tcping [host:port]",
		Short: "TCP connection test",
		Run: func(cmd *cobra.Command, args []string) {
			cli.Command = "tcping"
			if len(args) > 0 {
				cli.Host = args[0]
			}
			if err := handleTcping(cli); err != nil {
				os.Exit(1)
			}
		},
	}
	tcpingCmd.Flags().IntVarP(&cli.Port, "port", "p", 80, "Port number")

	// Website 命令
	var websiteCmd = &cobra.Command{
		Use:   "website [url]",
		Short: "Website test",
		Run: func(cmd *cobra.Command, args []string) {
			cli.Command = "website"
			if len(args) > 0 {
				cli.URL = args[0]
			}
			if err := handleWebsite(cli); err != nil {
				os.Exit(1)
			}
		},
	}
	websiteCmd.Flags().StringVarP(&cli.Method, "method", "M", "GET", "HTTP method")
	websiteCmd.Flags().BoolVarP(&cli.FollowRedirects, "follow-redirects", "f", false, "Follow redirects")

	// Traceroute 命令
	var tracerouteCmd = &cobra.Command{
		Use:   "traceroute [host]",
		Short: "Traceroute to a host",
		Run: func(cmd *cobra.Command, args []string) {
			cli.Command = "traceroute"
			if len(args) > 0 {
				cli.Host = args[0]
			}
			if err := handleTraceroute(cli); err != nil {
				os.Exit(1)
			}
		},
	}

	// DNS 命令
	var dnsCmd = &cobra.Command{
		Use:   "dns [host]",
		Short: "DNS query",
		Run: func(cmd *cobra.Command, args []string) {
			cli.Command = "dns"
			if len(args) > 0 {
				cli.Host = args[0]
			}
			if err := handleDNS(cli); err != nil {
				os.Exit(1)
			}
		},
	}

	// MTR 命令
	var mtrCmd = &cobra.Command{
		Use:   "mtr [host]",
		Short: "MTR test",
		Run: func(cmd *cobra.Command, args []string) {
			cli.Command = "mtr"
			if len(args) > 0 {
				cli.Host = args[0]
			}
			if err := handleMtr(cli); err != nil {
				os.Exit(1)
			}
		},
	}

	// Port-scan 命令
	var portScanCmd = &cobra.Command{
		Use:   "port-scan [host]",
		Short: "Port scan",
		Run: func(cmd *cobra.Command, args []string) {
			cli.Command = "port-scan"
			if len(args) > 0 {
				cli.Host = args[0]
			}
			if err := handlePortScan(cli); err != nil {
				os.Exit(1)
			}
		},
	}
	portScanCmd.Flags().StringVarP(&cli.Range, "range", "r", "1-1000", "Port range")

	// Install 命令
	var installCmd = &cobra.Command{
		Use:   "install",
		Short: "Install systemd service",
		Run: func(cmd *cobra.Command, args []string) {
			cli.Command = "install"
			if err := handleInstall(cli); err != nil {
				os.Exit(1)
			}
		},
	}
	installCmd.Flags().StringVarP(&cli.ServiceName, "name", "n", "network_probe", "Service name")

	// Uninstall 命令
	var uninstallCmd = &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall systemd service",
		Run: func(cmd *cobra.Command, args []string) {
			cli.Command = "uninstall"
			if err := handleUninstall(cli); err != nil {
				os.Exit(1)
			}
		},
	}
	uninstallCmd.Flags().StringVarP(&cli.ServiceName, "name", "n", "network_probe", "Service name")

	// Server 命令
	var serverCmd = &cobra.Command{
		Use:   "server",
		Short: "Start API server",
		Run: func(cmd *cobra.Command, args []string) {
			cli.Command = "server"
			if err := handleServer(cli); err != nil {
				os.Exit(1)
			}
		},
	}
	serverCmd.Flags().IntVarP(&cli.Port, "port", "p", 0, "Server port (default: from config file)")

	// Version 命令
	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Run: func(cmd *cobra.Command, args []string) {
			cli.Command = "version"
			fmt.Printf("network_probe version %s\n", version.Version)
		},
	}

	// Sysinfo 命令
	var sysinfoCmd = &cobra.Command{
		Use:   "sysinfo",
		Short: "Show system information",
		Run: func(cmd *cobra.Command, args []string) {
			cli.Command = "sysinfo"
			if err := handleSysinfo(cli); err != nil {
				os.Exit(1)
			}
		},
	}

	// GC 命令
	var gcCmd = &cobra.Command{
		Use:   "gc",
		Short: "Run garbage collection",
		Run: func(cmd *cobra.Command, args []string) {
			cli.Command = "gc"
			if err := handleGC(cli); err != nil {
				os.Exit(1)
			}
		},
	}

	// 添加子命令
	rootCmd.AddCommand(pingCmd)
	rootCmd.AddCommand(tcpingCmd)
	rootCmd.AddCommand(websiteCmd)
	rootCmd.AddCommand(tracerouteCmd)
	rootCmd.AddCommand(dnsCmd)
	rootCmd.AddCommand(mtrCmd)
	rootCmd.AddCommand(portScanCmd)
	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(uninstallCmd)
	rootCmd.AddCommand(serverCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(sysinfoCmd)
	rootCmd.AddCommand(gcCmd)

	return rootCmd.Execute()
}

// handleSysinfo 处理 sysinfo 命令
func handleSysinfo(cli *Cli) error {
	// 获取系统信息
	systemInfo, err := system.GetSystemInfo()
	if err != nil {
		fmt.Printf("Failed to get system info: %v\n", err)
		return err
	}

	// 打印系统信息
	fmt.Println("System Information:")
	fmt.Println("===================")
	fmt.Printf("Online Status:      %s\n", systemInfo.OnlineStatus)
	fmt.Printf("Download Bandwidth: %s\n", systemInfo.DownloadBandwidth)
	fmt.Printf("Upload Bandwidth:   %s\n", systemInfo.UploadBandwidth)
	fmt.Printf("Connections:        %s\n", systemInfo.Connections)
	fmt.Printf("Access Rate:        %s\n", systemInfo.AccessRate)
	fmt.Printf("Attack Rate:        %s\n", systemInfo.AttackRate)
	fmt.Printf("Cache Disk Usage:   %s\n", systemInfo.CacheDiskUsage)
	fmt.Printf("Max Disk Write:     %s\n", systemInfo.MaxDiskWriteSpeed)
	fmt.Printf("Memory Cache Usage: %s\n", systemInfo.MemoryCacheUsage)
	fmt.Printf("CPU Usage:          %s\n", systemInfo.CPUUsage)
	fmt.Printf("Memory Usage:       %s\n", systemInfo.MemoryUsage)
	fmt.Printf("Total Memory:       %s\n", systemInfo.TotalMemory)
	fmt.Printf("Load:               %s\n", systemInfo.Load)
	fmt.Printf("Monthly Traffic:    %s\n", systemInfo.MonthlyTraffic)
	fmt.Printf("Yesterday Traffic:  %s\n", systemInfo.YesterdayTraffic)
	fmt.Printf("Today Traffic:      %s\n", systemInfo.TodayTraffic)

	// 上报系统信息
	if err := report.ReportSystemInfo(systemInfo); err != nil {
		fmt.Printf("上报系统信息失败: %v\n", err)
	}

	return nil
}

// handleGC 处理 gc 命令
func handleGC(cli *Cli) error {
	// 记录开始时间
	start := time.Now()

	// 执行垃圾回收
	runtime.GC()

	// 计算执行时间
	duration := time.Since(start)

	// 打印结果
	fmt.Printf("ok, cost: %v, pause: %v\n", duration, duration)

	return nil
}

// handlePing 处理 ping 命令
func handlePing(cli *Cli) error {
	if cli.Host == "" {
		return fmt.Errorf("host is required")
	}

	service := modules.NewPingService()
	config := modules.NewPingConfig()
	config.Host = cli.Host
	config.Count = cli.Count
	config.Timeout = time.Duration(cli.Timeout) * time.Second

	result, err := service.Ping(config)
	if err != nil {
		fmt.Printf("Ping failed: %v\n", err)
		return err
	}

	fmt.Printf("Ping results for %s:\n", result.Host)
	fmt.Printf("  Packets sent: %d\n", result.Attempts)
	fmt.Printf("  Packets received: %d\n", result.SuccessfulAttempts)
	fmt.Printf("  Packet loss: %.1f%%\n", result.PacketLoss)
	fmt.Printf("  RTT min: %.2f ms\n", result.MinRTT)
	fmt.Printf("  RTT avg: %.2f ms\n", result.AvgRTT)
	fmt.Printf("  RTT max: %.2f ms\n", result.MaxRTT)

	return nil
}

// handleTcping 处理 tcping 命令
func handleTcping(cli *Cli) error {
	if cli.Host == "" {
		return fmt.Errorf("host is required")
	}

	service := modules.NewTcpingService()
	config := modules.NewTcpingConfig()
	config.Host = cli.Host
	config.Port = cli.Port
	config.Count = cli.Count
	config.Timeout = time.Duration(cli.Timeout) * time.Second

	result, err := service.Tcping(config)
	if err != nil {
		fmt.Printf("TCPing failed: %v\n", err)
		return err
	}

	// 上报结果
	if err := report.ReportCliTcping(map[string]interface{}{
		"host":    cli.Host,
		"port":    cli.Port,
		"count":   cli.Count,
		"timeout": cli.Timeout,
	}, result); err != nil {
		fmt.Printf("上报 TCPing 结果失败: %v\n", err)
	}

	fmt.Printf("  TCPing results for %s:%d:\n", result.Host, result.Port)
	fmt.Printf("  Connections attempted: %d\n", result.Attempts)
	fmt.Printf("  Connections successful: %d\n", result.SuccessfulAttempts)
	fmt.Printf("  Connection loss: %.1f%%\n", result.PacketLoss)
	fmt.Printf("  RTT min: %.2f ms\n", result.MinRTT)
	fmt.Printf("  RTT avg: %.2f ms\n", result.AvgRTT)
	fmt.Printf("  RTT max: %.2f ms\n", result.MaxRTT)

	return nil
}

// handleWebsite 处理 website 命令
func handleWebsite(cli *Cli) error {
	if cli.URL == "" {
		return fmt.Errorf("URL is required")
	}

	service := modules.NewWebsiteTestService()
	config := modules.NewWebsiteTestConfig()
	config.URL = cli.URL
	config.Method = cli.Method
	config.FollowRedirects = cli.FollowRedirects
	config.Timeout = time.Duration(cli.Timeout) * time.Second

	result, err := service.TestWebsite(config)
	if err != nil {
		fmt.Printf("Website test failed: %v\n", err)
		return err
	}

	fmt.Printf("Website test results for %s:\n", result.URL)
	fmt.Printf("  Status: %d %s\n", result.Status, result.StatusText)
	fmt.Printf("  Response time: %.2f ms\n", result.ResponseTime)
	fmt.Printf("  Content length: %d bytes\n", result.ContentLength)
	if result.RedirectURL != "" {
		fmt.Printf("  Redirect URL: %s\n", result.RedirectURL)
	}

	return nil
}

// handleTraceroute 处理 traceroute 命令
func handleTraceroute(cli *Cli) error {
	if cli.Host == "" {
		return fmt.Errorf("host is required")
	}

	service := modules.NewTracerouteService()
	config := modules.NewTracerouteConfig()
	config.Host = cli.Host
	config.MaxHops = cli.MaxHops

	result, err := service.Traceroute(config)
	if err != nil {
		fmt.Printf("Traceroute failed: %v\n", err)
		return err
	}

	fmt.Printf("Traceroute to %s:\n", result.Host)
	fmt.Println("  Hop  IP               Hostname        RTT")
	for _, hop := range result.Hops {
		hostname := hop.Hostname
		if hostname == "" {
			hostname = "*"
		}
		ip := hop.IP
		if ip == "" {
			ip = "*"
		}
		fmt.Printf("    %2d   %-15s  %-15s  %.2f ms\n", hop.Hop, ip, hostname, hop.RTT)
	}

	return nil
}

// handleDNS 处理 dns 命令
func handleDNS(cli *Cli) error {
	if cli.Host == "" {
		return fmt.Errorf("host is required")
	}

	service := modules.NewDnsService()
	config := modules.NewDnsConfig()
	config.Domain = cli.Host

	result, err := service.Query(config)
	if err != nil {
		fmt.Printf("DNS query failed: %v\n", err)
		return err
	}

	fmt.Printf("DNS query results for %s:\n", result.Domain)
	fmt.Printf("  Records:\n")
	for _, record := range result.Records {
		fmt.Printf("    %s  %s  %d\n", record.Type, record.Value, record.TTL)
	}

	return nil
}

// handleMtr 处理 mtr 命令
func handleMtr(cli *Cli) error {
	if cli.Host == "" {
		return fmt.Errorf("host is required")
	}

	fmt.Printf("MTR testing to %s...\n", cli.Host)
	fmt.Println()

	// 存储每个跳点的统计信息
	hopStatsMap := make(map[int]*modules.MtrHop)
	// 用于跟踪已显示的跳点
	shownHops := make(map[int]bool)

	// 跳点回调函数 - 实时更新汇总信息
	hopCallback := func(hop modules.MtrHop) error {
		// 清除屏幕并重新显示
		fmt.Print("\033[H\033[2J")
		fmt.Printf("MTR testing to %s...\n", cli.Host)
		fmt.Println()
		fmt.Println("HOP:    Address                Loss%%  Sent    Last     Avg    Best   Worst       Packets")

		// 按跳点编号排序显示
		for i := 1; i <= cli.MaxHops; i++ {
			if stats, ok := hopStatsMap[i]; ok {
				address := stats.IP
				if address == "" {
					address = "???"
				}
				hostname := stats.Hostname
				if hostname == "" {
					hostname = "*"
				}
				fmt.Printf("  %2d:|--  %-21s  %4.1f%%   %3d     %5.1f   %5.1f   %5.1f   %5.1f         %s\n",
					i, address, stats.Loss, stats.Snt, stats.Last, stats.Avg, stats.Best, stats.Wrst, stats.GetPacketStatusString())
				shownHops[i] = true
			}
		}
		return nil
	}

	// 数据包回调函数，实现实时回显
	packetCallback := func(packet modules.MtrPacketResult) error {
		// 更新统计信息
		stats, ok := hopStatsMap[packet.Hop]
		if !ok {
			stats = &modules.MtrHop{
				Hop:      packet.Hop,
				IP:       packet.IP,
				Hostname: packet.Hostname,
				Snt:      0,
				Loss:     100.0,
				Last:     0,
				Avg:      0,
				Best:     999999,
				Wrst:     0,
				Packets:  []bool{},
			}
			hopStatsMap[packet.Hop] = stats
		}

		// 更新主机名和 IP（如果有新信息）
		if packet.Hostname != "" {
			stats.Hostname = packet.Hostname
		}
		if packet.IP != "" {
			stats.IP = packet.IP
		}

		// 更新统计数据
		stats.Snt++
		if packet.Loss {
			stats.Packets = append(stats.Packets, false)
		} else {
			stats.Packets = append(stats.Packets, true)
			stats.Last = packet.RTT

			// 更新最佳、最差值
			if packet.RTT < stats.Best {
				stats.Best = packet.RTT
			}
			if packet.RTT > stats.Wrst {
				stats.Wrst = packet.RTT
			}

			// 重新计算平均值
			total := 0.0
			count := 0
			for _, success := range stats.Packets {
				if success {
					// 这里需要存储 RTT 值，暂时用 Last 近似
					total += stats.Last
					count++
				}
			}
			if count > 0 {
				stats.Avg = total / float64(count)
			}

			// 重新计算丢包率
			lossCount := 0
			for _, success := range stats.Packets {
				if !success {
					lossCount++
				}
			}
			stats.Loss = float64(lossCount) / float64(stats.Snt) * 100
		}

		return nil
	}

	service := modules.NewMtrServiceWithPacketCallback(hopCallback, packetCallback)
	config := modules.NewMtrConfig()
	config.Host = cli.Host
	config.MaxHops = cli.MaxHops
	config.Count = cli.Count
	config.Interval = cli.Timeout

	result, err := service.Mtr(config)
	if err != nil {
		fmt.Printf("MTR failed: %v\n", err)
		return err
	}

	// 显示最终结果
	fmt.Print("\033[H\033[2J")
	fmt.Printf("MTR results for %s:\n", result.Host)
	fmt.Println("HOP:    Address                Loss%%  Sent    Last     Avg    Best   Worst       Packets")
	for _, hop := range result.Hops {
		address := hop.IP
		if address == "" {
			address = "???"
		}
		hostname := hop.Hostname
		if hostname == "" {
			hostname = "*"
		}
		fmt.Printf("  %2d:|--  %-21s  %4.1f%%   %3d     %5.1f   %5.1f   %5.1f   %5.1f         %s\n",
			hop.Hop, address, hop.Loss, hop.Snt, hop.Last, hop.Avg, hop.Best, hop.Wrst, hop.GetPacketStatusString())
	}

	return nil
}

// handlePortScan 处理 port-scan 命令
func handlePortScan(cli *Cli) error {
	fmt.Printf("Scanning ports on %s (range: %s)...\n", cli.Host, cli.Range)

	// 解析端口范围
	ports, err := parsePortRange(cli.Range)
	if err != nil {
		fmt.Printf("Invalid port range: %v\n", err)
		os.Exit(1)
	}

	service := modules.NewTcpingService()
	results, err := service.ScanPorts(cli.Host, ports, time.Duration(cli.Timeout)*time.Millisecond)
	if err != nil {
		fmt.Printf("Port scan failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Port scan results for %s:\n", cli.Host)
	var openPorts []int
	var closedCount int

	for port, isOpen := range results {
		if isOpen {
			openPorts = append(openPorts, port)
		} else {
			closedCount++
		}
	}

	fmt.Printf("  Open ports: %v\n", openPorts)
	fmt.Printf("  Closed ports: %d total\n", closedCount)

	return nil
}

// handleInstall 处理 install 命令
func handleInstall(cli *Cli) error {
	serviceName := cli.ServiceName
	if serviceName == "" {
		serviceName = "network_probe"
	}

	// 确定主机和端口
	host := cli.Host
	if host == "" {
		host = "0.0.0.0"
	}

	port := cli.Port
	if port == 0 {
		// 尝试从配置文件中读取端口
		cfg, err := config.LoadConfig(config.GetConfigPath())
		if err == nil && cfg.Port > 0 {
			port = cfg.Port
		} else {
			port = 8080
		}
	}

	fmt.Printf("Installing systemd service '%s'...\n", serviceName)

	// 获取可执行文件路径
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %v", err)
	}

	// 创建 systemd 服务文件内容
	serviceFileContent := fmt.Sprintf(`[Unit]
Description=Network Probe Service
After=network.target

[Service]
Type=simple
ExecStart=%s server -p %d
Restart=always
RestartSec=5
User=root

[Install]
WantedBy=multi-user.target
`, execPath, port)

	// 服务文件路径
	serviceFilePath := fmt.Sprintf("/etc/systemd/system/%s.service", serviceName)

	// 写入服务文件
	err = os.WriteFile(serviceFilePath, []byte(serviceFileContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to write service file: %v", err)
	}

	// 重新加载 systemd 配置
	err = exec.Command("systemctl", "daemon-reload").Run()
	if err != nil {
		return fmt.Errorf("failed to reload systemd: %v", err)
	}

	// 启用服务
	err = exec.Command("systemctl", "enable", serviceName).Run()
	if err != nil {
		return fmt.Errorf("failed to enable service: %v", err)
	}

	// 启动服务
	err = exec.Command("systemctl", "start", serviceName).Run()
	if err != nil {
		return fmt.Errorf("failed to start service: %v", err)
	}

	fmt.Printf("Service '%s' installed and started successfully on %s:%d\n", serviceName, host, port)
	return nil
}

// handleUninstall 处理 uninstall 命令
func handleUninstall(cli *Cli) error {
	serviceName := cli.ServiceName
	if serviceName == "" {
		serviceName = "network_probe"
	}

	fmt.Printf("uninstalling systemd service '%s'...\n", serviceName)

	// 停止服务
	err := exec.Command("systemctl", "stop", serviceName).Run()
	if err != nil {
		// 服务可能已经停止，继续执行
		fmt.Printf("warning: failed to stop service (it may already be stopped): %v\n", err)
	}

	// 禁用服务
	err = exec.Command("systemctl", "disable", serviceName).Run()
	if err != nil {
		// 服务可能已经禁用，继续执行
		fmt.Printf("warning: failed to disable service (it may already be disabled): %v\n", err)
	}

	// 删除服务文件
	serviceFilePath := fmt.Sprintf("/etc/systemd/system/%s.service", serviceName)
	err = os.Remove(serviceFilePath)
	if err != nil {
		// 服务文件可能不存在，继续执行
		fmt.Printf("warning: failed to remove service file (it may not exist): %v\n", err)
	}

	// 重新加载 systemd 配置
	err = exec.Command("systemctl", "daemon-reload").Run()
	if err != nil {
		return fmt.Errorf("failed to reload systemd: %v", err)
	}

	fmt.Printf("service '%s' uninstalled successfully\n", serviceName)
	return nil
}

// handleServer 处理 server 命令
func handleServer(cli *Cli) error {
	// 创建 API 服务器
	server := api.NewServer()

	// 确定端口：如果命令行指定了端口，使用命令行端口；否则使用配置文件中的端口
	port := cli.Port
	if port == 0 {
		port = server.GetConfig().Port
	}

	// 确定主机：如果命令行指定了主机，使用命令行主机；否则使用默认值
	host := cli.Host
	if host == "" {
		host = "0.0.0.0"
	}

	// 打印服务信息
	fmt.Printf("starting network probe server on %s:%d\n", host, port)
	// fmt.Println("Available endpoints:")
	// fmt.Println("  POST /api/ping - ICMP ping test")
	// fmt.Println("  POST /api/tcping - TCP connection test")
	// fmt.Println("  POST /api/website - Website test")
	// fmt.Println("  POST /api/traceroute - Traceroute")
	// fmt.Println("  POST /api/dns - DNS query")
	// fmt.Println("  POST /api/mtr - MTR test")
	// fmt.Println("  GET  /api/health - Health check")
	// fmt.Println("  GET  /api/status - Service status")
	return server.Run(fmt.Sprintf("%s:%d", host, port))
}

// parsePortRange 解析端口范围
func parsePortRange(rangeStr string) ([]int, error) {
	var ports []int

	if strings.Contains(rangeStr, "-") {
		parts := strings.Split(rangeStr, "-")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid port range format")
		}

		start, err := strconv.Atoi(parts[0])
		if err != nil {
			return nil, fmt.Errorf("invalid start port: %v", err)
		}

		end, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, fmt.Errorf("invalid end port: %v", err)
		}

		for i := start; i <= end; i++ {
			ports = append(ports, i)
		}
	} else {
		port, err := strconv.Atoi(rangeStr)
		if err != nil {
			return nil, fmt.Errorf("invalid port: %v", err)
		}
		ports = append(ports, port)
	}

	return ports, nil
}

package cli

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"network-probe/internal/api"
	"network-probe/internal/modules"
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
	Protocol        string
	Domain          string
	QueryType       string
	Nameserver      string
	Range           string
	ServiceName     string
}

// NewCli 创建一个新的命令行参数结构
func NewCli() *Cli {
	return &Cli{
		LogLevel: "info",
	}
}

// Run 运行命令行工具
func Run() error {
	cli := NewCli()

	// 创建根命令
	rootCmd := &cobra.Command{
		Use:   "network-probe",
		Short: "A comprehensive network testing tool",
		Long:  "A comprehensive network testing tool that supports ping, tcping, traceroute, DNS queries, website testing, and port scanning.",
	}

	// 添加日志级别参数
	rootCmd.PersistentFlags().StringVarP(&cli.LogLevel, "log-level", "l", "info", "Set log level")

	// 添加 ping 命令
	pingCmd := &cobra.Command{
		Use:   "ping",
		Short: "Perform ICMP ping test",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("ping command requires a host argument")
			}
			cli.Command = "ping"
			cli.Host = args[0]
			return handlePing(cli)
		},
	}
	pingCmd.Flags().IntVarP(&cli.Count, "count", "c", 4, "Number of ping packets to send")
	pingCmd.Flags().IntVarP(&cli.Timeout, "timeout", "t", 2, "Timeout in seconds")
	rootCmd.AddCommand(pingCmd)

	// 添加 tcping 命令
	tcpingCmd := &cobra.Command{
		Use:   "tcping",
		Short: "Perform TCP connection test",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("tcping command requires a host argument")
			}
			cli.Command = "tcping"
			cli.Host = args[0]
			return handleTcping(cli)
		},
	}
	tcpingCmd.Flags().IntVarP(&cli.Port, "port", "p", 80, "Target port")
	tcpingCmd.Flags().IntVarP(&cli.Count, "count", "c", 4, "Number of connection attempts")
	tcpingCmd.Flags().IntVarP(&cli.Timeout, "timeout", "t", 3, "Timeout in seconds")
	rootCmd.AddCommand(tcpingCmd)

	// 添加 website 命令
	websiteCmd := &cobra.Command{
		Use:   "website",
		Short: "Test website availability",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("website command requires a URL argument")
			}
			cli.Command = "website"
			cli.URL = args[0]
			return handleWebsite(cli)
		},
	}
	websiteCmd.Flags().StringVarP(&cli.Method, "method", "m", "GET", "HTTP method")
	websiteCmd.Flags().BoolVarP(&cli.FollowRedirects, "follow-redirects", "f", false, "Follow redirects")
	websiteCmd.Flags().IntVarP(&cli.Timeout, "timeout", "t", 30, "Timeout in seconds")
	rootCmd.AddCommand(websiteCmd)

	// 添加 traceroute 命令
	tracerouteCmd := &cobra.Command{
		Use:   "traceroute",
		Short: "Perform traceroute",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("traceroute command requires a host argument")
			}
			cli.Command = "traceroute"
			cli.Host = args[0]
			return handleTraceroute(cli)
		},
	}
	tracerouteCmd.Flags().IntVarP(&cli.MaxHops, "max-hops", "m", 30, "Maximum number of hops")
	tracerouteCmd.Flags().StringVarP(&cli.Protocol, "protocol", "p", "icmp", "Protocol (icmp, udp, tcp)")
	rootCmd.AddCommand(tracerouteCmd)

	// 添加 dns 命令
	dnsCmd := &cobra.Command{
		Use:   "dns",
		Short: "Perform DNS query",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("dns command requires a domain argument")
			}
			cli.Command = "dns"
			cli.Domain = args[0]
			return handleDns(cli)
		},
	}
	dnsCmd.Flags().StringVarP(&cli.QueryType, "query-type", "t", "A", "Query type (A, AAAA, CNAME, MX, TXT, NS, SOA, PTR, ALL)")
	dnsCmd.Flags().StringVarP(&cli.Nameserver, "nameserver", "n", "", "Custom nameserver")
	rootCmd.AddCommand(dnsCmd)

	// 添加 server 命令
	serverCmd := &cobra.Command{
		Use:   "server",
		Short: "Start API server",
		RunE: func(cmd *cobra.Command, args []string) error {
			cli.Command = "server"
			return handleServer(cli)
		},
	}
	serverCmd.Flags().StringVarP(&cli.Host, "host", "H", "127.0.0.1", "Server host")
	serverCmd.Flags().IntVarP(&cli.Port, "port", "p", 8080, "Server port")
	rootCmd.AddCommand(serverCmd)

	// 添加 port-scan 命令
	portScanCmd := &cobra.Command{
		Use:   "port-scan",
		Short: "Scan ports on target host",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("port-scan command requires a host argument")
			}
			cli.Command = "port-scan"
			cli.Host = args[0]
			return handlePortScan(cli)
		},
	}
	portScanCmd.Flags().StringVarP(&cli.Range, "range", "r", "1-1000", "Port range (e.g., 1-1000)")
	portScanCmd.Flags().IntVarP(&cli.Timeout, "timeout", "t", 1000, "Timeout in milliseconds")
	rootCmd.AddCommand(portScanCmd)

	// 添加 install 命令
	installCmd := &cobra.Command{
		Use:   "install",
		Short: "Install as a systemd service",
		RunE: func(cmd *cobra.Command, args []string) error {
			cli.Command = "install"
			return handleInstall(cli)
		},
	}
	installCmd.Flags().StringVarP(&cli.Host, "host", "H", "127.0.0.1", "Server host")
	installCmd.Flags().IntVarP(&cli.Port, "port", "p", 8080, "Server port")
	installCmd.Flags().StringVarP(&cli.ServiceName, "service-name", "s", "network-probe", "Service name")
	rootCmd.AddCommand(installCmd)

	// 添加 uninstall 命令
	uninstallCmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall systemd service",
		RunE: func(cmd *cobra.Command, args []string) error {
			cli.Command = "uninstall"
			return handleUninstall(cli)
		},
	}
	uninstallCmd.Flags().StringVarP(&cli.ServiceName, "service-name", "s", "network-probe", "Service name")
	rootCmd.AddCommand(uninstallCmd)

	// 执行命令
	return rootCmd.Execute()
}

// handlePing 处理 ping 命令
func handlePing(cli *Cli) error {
	fmt.Printf("Pinging %s...\n", cli.Host)

	service := modules.NewPingService()
	config := modules.NewPingConfig()
	config.Host = cli.Host
	config.Count = cli.Count
	config.Timeout = time.Duration(cli.Timeout) * time.Second

	result, err := service.Ping(config)
	if err != nil {
		fmt.Printf("Ping failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Ping results for %s (%s):\n", result.Host, result.IP)
	fmt.Printf("  Packets: sent = %d, received = %d, loss = %.1f%%\n",
		result.Attempts, result.SuccessfulAttempts, result.PacketLoss)
	fmt.Printf("  RTT (ms): min = %.2f, max = %.2f, avg = %.2f\n",
		result.MinRTT, result.MaxRTT, result.AvgRTT)

	return nil
}

// handleTcping 处理 tcping 命令
func handleTcping(cli *Cli) error {
	fmt.Printf("TCPing %s:%d...\n", cli.Host, cli.Port)

	service := modules.NewTcpingService()
	config := modules.NewTcpingConfig()
	config.Host = cli.Host
	config.Port = cli.Port
	config.Count = cli.Count
	config.Timeout = time.Duration(cli.Timeout) * time.Second

	result, err := service.Tcping(config)
	if err != nil {
		fmt.Printf("TCPing failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("TCPing results for %s:%d (%s):\n", result.Host, result.Port, result.IP)
	fmt.Printf("  Attempts: %d, successful = %d, loss = %.1f%%\n",
		result.Attempts, result.SuccessfulAttempts, result.PacketLoss)
	fmt.Printf("  RTT (ms): min = %.2f, max = %.2f, avg = %.2f\n",
		result.MinRTT, result.MaxRTT, result.AvgRTT)

	return nil
}

// handleWebsite 处理 website 命令
func handleWebsite(cli *Cli) error {
	fmt.Printf("Testing website %s...\n", cli.URL)

	service := modules.NewWebsiteTestService()
	config := modules.NewWebsiteTestConfig()
	config.URL = cli.URL
	config.Method = cli.Method
	config.Timeout = time.Duration(cli.Timeout) * time.Second
	config.FollowRedirects = cli.FollowRedirects

	result, err := service.TestWebsite(config)
	if err != nil {
		fmt.Printf("Website test failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Website test results for %s:\n", result.URL)
	if result.Status != 0 {
		fmt.Printf("  Status: %s\n", result.StatusText)
		fmt.Printf("  Status Code: %d\n", result.Status)
	}
	fmt.Printf("  Response Time: %.2fms\n", result.ResponseTime)
	if result.ContentLength != 0 {
		fmt.Printf("  Content Length: %d bytes\n", result.ContentLength)
	}
	if result.Error != "" {
		fmt.Printf("  Error: %s\n", result.Error)
	}

	return nil
}

// handleTraceroute 处理 traceroute 命令
func handleTraceroute(cli *Cli) error {
	fmt.Printf("Tracerouting to %s using %s...\n", cli.Host, cli.Protocol)

	service := modules.NewTracerouteService()
	config := modules.NewTracerouteConfig()
	config.Host = cli.Host
	config.MaxHops = cli.MaxHops
	config.Protocol = cli.Protocol

	result, err := service.Traceroute(config)
	if err != nil {
		fmt.Printf("Traceroute failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Traceroute results for %s:\n", result.Host)
	fmt.Println("  Hops:")
	for _, hop := range result.Hops {
		hostname := hop.Hostname
		if hostname == "" {
			hostname = "*"
		}
		ip := hop.IP
		if ip == "" {
			ip = "*"
		}
		rtt := fmt.Sprintf("%.2fms", hop.RTT)
		if hop.Error != "" {
			fmt.Printf("    %2d: * * * (%s)\n", hop.Hop, hop.Error)
		} else {
			fmt.Printf("    %2d: %s (%s) %s\n", hop.Hop, hostname, ip, rtt)
		}
	}

	return nil
}

// handleDns 处理 dns 命令
func handleDns(cli *Cli) error {
	fmt.Printf("DNS query for %s (type: %s)...\n", cli.Domain, cli.QueryType)

	var service *modules.DnsService
	var err error

	if cli.Nameserver != "" {
		service, err = modules.NewDnsServiceWithNameserver(cli.Nameserver)
	} else {
		service = modules.NewDnsService()
	}

	if err != nil {
		fmt.Printf("DNS service creation failed: %v\n", err)
		os.Exit(1)
	}

	config := modules.NewDnsConfig()
	config.Domain = cli.Domain
	config.QueryType = modules.DnsQueryType(cli.QueryType)
	config.Nameserver = cli.Nameserver

	result, err := service.Query(config)
	if err != nil {
		fmt.Printf("DNS query failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("DNS query results for %s (type: %s):\n", result.Domain, config.QueryType)
	fmt.Println("  Records:")
	for _, record := range result.Records {
		fmt.Printf("    %s: %s\n", record.Type, record.Value)
	}
	if result.Error != "" {
		fmt.Printf("  Error: %s\n", result.Error)
	}

	return nil
}

// handleServer 处理 server 命令
func handleServer(cli *Cli) error {
	fmt.Printf("Starting API server on %s:%d...\n", cli.Host, cli.Port)

	server := api.NewServer()

	fmt.Printf("Server listening on http://%s:%d\n", cli.Host, cli.Port)
	fmt.Println("API endpoints:")
	fmt.Println("  POST /api/ping - ICMP ping test")
	fmt.Println("  POST /api/tcping - TCP connection test")
	fmt.Println("  POST /api/website - Website test")
	fmt.Println("  POST /api/traceroute - Traceroute")
	fmt.Println("  POST /api/dns - DNS query")
	fmt.Println("  GET  /api/health - Health check")
	fmt.Println("  GET  /api/status - Service status")

	return server.Run(fmt.Sprintf("%s:%d", cli.Host, cli.Port))
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
	fmt.Printf("Installing systemd service '%s'...\n", cli.ServiceName)

	// 这里应该实现系统服务的安装逻辑
	// 由于这是一个示例，这里暂时只打印信息
	fmt.Printf("Service '%s' would be installed on %s:%d\n", cli.ServiceName, cli.Host, cli.Port)

	return nil
}

// handleUninstall 处理 uninstall 命令
func handleUninstall(cli *Cli) error {
	fmt.Printf("Uninstalling systemd service '%s'...\n", cli.ServiceName)

	// 这里应该实现系统服务的卸载逻辑
	// 由于这是一个示例，这里暂时只打印信息
	fmt.Printf("Service '%s' would be uninstalled\n", cli.ServiceName)

	return nil
}

// parsePortRange 解析端口范围
func parsePortRange(rangeStr string) ([]int, error) {
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

		if start > end {
			return nil, fmt.Errorf("invalid port range: start > end")
		}

		ports := make([]int, 0, end-start+1)
		for i := start; i <= end; i++ {
			ports = append(ports, i)
		}

		return ports, nil
	} else {
		port, err := strconv.Atoi(rangeStr)
		if err != nil {
			return nil, fmt.Errorf("invalid port: %v", err)
		}

		return []int{port}, nil
	}
}

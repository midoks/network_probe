package report

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"network-probe/internal/config"
	"network-probe/internal/version"
)

// ReportType 表示上报类型
type ReportType string

// SubType 表示上报子类型
type SubType string

// 上报类型常量
const (
	// 系统信息上报类型
	ReportTypeSystem  ReportType = "sys"
	ReportTypeNode    ReportType = "node"
	ReportTypeRequest ReportType = "request"
)

const (
	SystemInfo SubType = "system_info"

	NodeInfo                SubType = "info"
	NodeWebSocketConnect    SubType = "websocket_connect"
	NodeWebSocketDisconnect SubType = "websocket_disconnect"

	// 功能请求上报类型
	RequestPing                 SubType = "ping"
	RequestTcping               SubType = "tcping"
	RequestWebsite              SubType = "website"
	RequestTraceroute           SubType = "traceroute"
	RequestDns                  SubType = "dns"
	RequestMtr                  SubType = "mtr"
	RequestWebSocketPing        SubType = "websocket_ping"
	RequestWebSocketTcping      SubType = "websocket_tcping"
	RequestWebSocketWebsite     SubType = "websocket_website"
	RequestWebSocketTraceroute  SubType = "websocket_traceroute"
	RequestWebSocketDns         SubType = "websocket_dns"
	RequestWebSocketMtr         SubType = "websocket_mtr"
	RequestWebSocketMtrStart    SubType = "websocket_mtr_start"
	RequestWebSocketMtrComplete SubType = "websocket_mtr_complete"
	RequestCliPing              SubType = "cli_ping"
	RequestCliTcping            SubType = "cli_tcping"
)

// ReportData 表示上报数据结构
type ReportData struct {
	Type      ReportType  `json:"type"`
	SubType   SubType     `json:"sub_type"`
	Timestamp int64       `json:"timestamp"`
	Data      interface{} `json:"data,omitempty"`
	Version   string      `json:"version,omitempty"`
}

// PingRequest 表示 ping 请求
type PingRequest struct {
	Host    string `json:"host"`
	Count   int    `json:"count"`
	Timeout int    `json:"timeout"`
}

// PingReportData 表示 ping 上报数据
type PingReportData struct {
	Request PingRequest `json:"request"`
	Result  interface{} `json:"result"`
}

// TcpingRequest 表示 tcping 请求
type TcpingRequest struct {
	Host    string `json:"host"`
	Port    int    `json:"port"`
	Count   int    `json:"count"`
	Timeout int    `json:"timeout"`
}

// TcpingReportData 表示 tcping 上报数据
type TcpingReportData struct {
	Request TcpingRequest `json:"request"`
	Result  interface{}   `json:"result"`
}

// WebsiteRequest 表示 website 请求
type WebsiteRequest struct {
	URL             string `json:"url"`
	Method          string `json:"method"`
	Timeout         int    `json:"timeout"`
	FollowRedirects bool   `json:"follow_redirects"`
}

// WebsiteReportData 表示 website 上报数据
type WebsiteReportData struct {
	Request WebsiteRequest `json:"request"`
	Result  interface{}    `json:"result"`
}

// TracerouteRequest 表示 traceroute 请求
type TracerouteRequest struct {
	Host     string `json:"host"`
	MaxHops  int    `json:"max_hops"`
	Protocol string `json:"protocol"`
}

// TracerouteReportData 表示 traceroute 上报数据
type TracerouteReportData struct {
	Request TracerouteRequest `json:"request"`
	Result  interface{}       `json:"result"`
}

// DnsRequest 表示 dns 请求
type DnsRequest struct {
	Domain     string `json:"domain"`
	QueryType  string `json:"query_type"`
	Nameserver string `json:"nameserver"`
}

// DnsReportData 表示 dns 上报数据
type DnsReportData struct {
	Request DnsRequest  `json:"request"`
	Result  interface{} `json:"result"`
}

// MtrRequest 表示 mtr 请求
type MtrRequest struct {
	Host     string `json:"host"`
	MaxHops  int    `json:"max_hops"`
	Count    int    `json:"count"`
	Interval int    `json:"interval"`
}

// MtrReportData 表示 mtr 上报数据
type MtrReportData struct {
	Request MtrRequest  `json:"request"`
	Result  interface{} `json:"result"`
}

// StartupRequest 表示启动请求
type StartupRequest struct {
	Address string `json:"address"`
	Version string `json:"version"`
	Status  string `json:"status"`
}

// StartupReportData 表示启动上报数据
type StartupReportData struct {
	Msg string `json:"msg"`
}

// WebSocketConnectReportData 表示 WebSocket 连接上报数据
type WebSocketConnectReportData struct {
	Status string `json:"status"`
}

// SystemInfoReportData 表示系统信息上报数据
type SystemInfoReportData struct {
	Result interface{} `json:"result"`
}

// Report 上报数据
func Report(reportType ReportType, subType SubType, data interface{}) error {
	cfg, err := config.LoadConfig(config.GetConfigPath())
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}

	if len(cfg.ReportEndpoints) == 0 {
		return fmt.Errorf("no report endpoints configured")
	}

	// 准备上报数据
	reportData := ReportData{
		Type:      reportType,
		SubType:   subType,
		Timestamp: time.Now().Unix(),
	}

	reportData.Version = version.Version

	fmt.Println(data)
	reportData.Data = data

	// 序列化数据
	jsonData, err := json.Marshal(reportData)
	if err != nil {
		return fmt.Errorf("failed to marshal report data: %v", err)
	}

	// 上报到每个端点（同步）
	var lastError error
	for _, endpoint := range cfg.ReportEndpoints {
		post_url := endpoint + "/api/logs"
		req, err := http.NewRequest("POST", post_url, bytes.NewBuffer(jsonData))
		if err != nil {
			lastError = fmt.Errorf("failed to create request to %s: %v", post_url, err)
			fmt.Printf("Failed to create request to %s: %v\n", post_url, err)
			continue
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Node-ID", cfg.NodeID)
		req.Header.Set("X-Secret", cfg.Secret)

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			lastError = fmt.Errorf("failed to report to %s: %v", post_url, err)
			fmt.Printf("Failed to report to %s: %v\n", post_url, err)
			continue
		}
		defer resp.Body.Close()
		fmt.Printf("Reported to %s successfully: %v\n", post_url, resp.Status)
	}

	return lastError
}

// ReportStartup 上报启动日志
func ReportNodeInfo(msg string) error {
	return Report(ReportTypeNode, NodeInfo, map[string]interface{}{
		"description": msg,
	})
}

// ReportPing 上报 ping 结果
func ReportPing(request, result interface{}) error {
	return Report(ReportTypeRequest, RequestPing, map[string]interface{}{
		"request": request,
		"result":  result,
	})
}

// ReportTcping 上报 tcping 结果
func ReportTcping(request, result interface{}) error {
	return Report(ReportTypeRequest, RequestTcping, map[string]interface{}{
		"request": request,
		"result":  result,
	})
}

// ReportWebsite 上报 website 结果
func ReportWebsite(request, result interface{}) error {
	return Report(ReportTypeRequest, RequestWebsite, map[string]interface{}{
		"request": request,
		"result":  result,
	})
}

// ReportTraceroute 上报 traceroute 结果
func ReportTraceroute(request, result interface{}) error {
	return Report(ReportTypeRequest, RequestTraceroute, map[string]interface{}{
		"request": request,
		"result":  result,
	})
}

// ReportDns 上报 dns 结果
func ReportDns(request, result interface{}) error {
	return Report(ReportTypeRequest, RequestDns, map[string]interface{}{
		"request": request,
		"result":  result,
	})
}

// ReportMtr 上报 mtr 结果
func ReportMtr(request, result interface{}) error {
	return Report(ReportTypeRequest, RequestMtr, map[string]interface{}{
		"request": request,
		"result":  result,
	})
}

// ReportWebSocketConnect 上报 WebSocket 连接
func ReportWebSocketConnect() error {
	return Report(ReportTypeNode, NodeWebSocketConnect, map[string]interface{}{
		"status": "connected",
	})
}

// ReportWebSocketDisconnect 上报 WebSocket 断开连接
func ReportWebSocketDisconnect() error {
	return Report(ReportTypeNode, NodeWebSocketDisconnect, map[string]interface{}{
		"status": "disconnected",
	})
}

// ReportWebSocketPing 上报 WebSocket ping 结果
func ReportWebSocketPing(request, result interface{}) error {
	return Report(ReportTypeRequest, RequestWebSocketPing, map[string]interface{}{
		"request": request,
		"result":  result,
	})
}

// ReportWebSocketTcping 上报 WebSocket tcping 结果
func ReportWebSocketTcping(request, result interface{}) error {
	return Report(ReportTypeRequest, RequestWebSocketTcping, map[string]interface{}{
		"request": request,
		"result":  result,
	})
}

// ReportWebSocketWebsite 上报 WebSocket website 结果
func ReportWebSocketWebsite(request, result interface{}) error {
	return Report(ReportTypeRequest, RequestWebSocketWebsite, map[string]interface{}{
		"request": request,
		"result":  result,
	})
}

// ReportWebSocketTraceroute 上报 WebSocket traceroute 结果
func ReportWebSocketTraceroute(request, result interface{}) error {
	return Report(ReportTypeRequest, RequestWebSocketTraceroute, map[string]interface{}{
		"request": request,
		"result":  result,
	})
}

// ReportWebSocketDns 上报 WebSocket dns 结果
func ReportWebSocketDns(request, result interface{}) error {
	return Report(ReportTypeRequest, RequestWebSocketDns, map[string]interface{}{
		"request": request,
		"result":  result,
	})
}

// ReportWebSocketMtr 上报 WebSocket mtr 结果
func ReportWebSocketMtr(request, result interface{}) error {
	return Report(ReportTypeRequest, RequestWebSocketMtr, map[string]interface{}{
		"request": request,
		"result":  result,
	})
}

// ReportWebSocketMtrStart 上报 WebSocket mtr 开始
func ReportWebSocketMtrStart(request interface{}) error {
	return Report(ReportTypeRequest, RequestWebSocketMtrStart, map[string]interface{}{
		"request": request,
	})
}

// ReportWebSocketMtrComplete 上报 WebSocket mtr 完成
func ReportWebSocketMtrComplete(request, result interface{}) error {
	return Report(ReportTypeRequest, RequestWebSocketMtrComplete, map[string]interface{}{
		"request": request,
		"result":  result,
	})
}

// ReportCliPing 上报 CLI ping 结果
func ReportCliPing(request, result interface{}) error {
	return Report(ReportTypeRequest, RequestCliPing, map[string]interface{}{
		"request": request,
		"result":  result,
	})
}

// ReportCliTcping 上报 CLI tcping 结果
func ReportCliTcping(request, result interface{}) error {
	return Report(ReportTypeRequest, RequestCliTcping, map[string]interface{}{
		"request": request,
		"result":  result,
	})
}

// ReportSystemInfo 上报节点运行信息
func ReportSystemInfo(data interface{}) error {
	return Report(ReportTypeSystem, SystemInfo, data)
}

// ReportErrorLog 上报错误日志
func ReportErrorLog(entry interface{}) error {
	return Report(ReportTypeSystem, "error", map[string]interface{}{
		"data": entry,
	})
}

// ReportCrashLog 上报崩溃日志
func ReportCrashLog(entry interface{}) error {
	return Report(ReportTypeSystem, "crash", map[string]interface{}{
		"data": entry,
	})
}

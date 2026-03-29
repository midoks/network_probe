package report

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"network-probe/internal/config"
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

	NodeStartup             SubType = "startup"
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
type ReportDataStartup struct {
	Msg string `json:"msg"`
}

// ReportData 表示上报数据结构
type ReportData struct {
	Type      ReportType  `json:"type"`
	SubType   SubType     `json:"sub_type"`
	Timestamp string      `json:"timestamp"`
	Data      interface{} `json:"Data,omitempty"`
	Version   string      `json:"version,omitempty"`
}

// Report 上报数据
func Report(reportType ReportType, subType SubType, data map[string]interface{}) error {
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
		Timestamp: time.Now().Format(time.RFC3339),
	}

	// 填充其他字段
	if data, ok := data["data"]; ok {
		reportData.Data = data
	}
	if version, ok := data["version"].(string); ok {
		reportData.Version = version
	}

	// 序列化数据
	jsonData, err := json.Marshal(reportData)
	if err != nil {
		return fmt.Errorf("failed to marshal report data: %v", err)
	}

	// 上报到每个端点（同步）
	var lastError error
	for _, endpoint := range cfg.ReportEndpoints {
		req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonData))
		if err != nil {
			lastError = fmt.Errorf("failed to create request to %s: %v", endpoint, err)
			fmt.Printf("Failed to create request to %s: %v\n", endpoint, err)
			continue
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Node-ID", cfg.NodeID)
		req.Header.Set("X-Secret", cfg.Secret)

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			lastError = fmt.Errorf("failed to report to %s: %v", endpoint, err)
			fmt.Printf("Failed to report to %s: %v\n", endpoint, err)
			continue
		}
		defer resp.Body.Close()
		fmt.Printf("Reported to %s successfully: %v\n", endpoint, resp.Status)
	}

	return lastError
}

// ReportStartup 上报启动日志
func ReportStartup(msg string) error {
	return Report(ReportTypeNode, NodeStartup, map[string]interface{}{
		"address": address,
		"version": version,
		"status":  "started",
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
	return Report(ReportTypeRequestTcping, map[string]interface{}{
		"request": request,
		"result":  result,
	})
}

// ReportWebsite 上报 website 结果
func ReportWebsite(request, result interface{}) error {
	return Report(ReportTypeRequestWebsite, map[string]interface{}{
		"request": request,
		"result":  result,
	})
}

// ReportTraceroute 上报 traceroute 结果
func ReportTraceroute(request, result interface{}) error {
	return Report(ReportTypeRequestTraceroute, map[string]interface{}{
		"request": request,
		"result":  result,
	})
}

// ReportDns 上报 dns 结果
func ReportDns(request, result interface{}) error {
	return Report(ReportTypeRequestDns, map[string]interface{}{
		"request": request,
		"result":  result,
	})
}

// ReportMtr 上报 mtr 结果
func ReportMtr(request, result interface{}) error {
	return Report(ReportTypeRequestMtr, map[string]interface{}{
		"request": request,
		"result":  result,
	})
}

// ReportWebSocketConnect 上报 WebSocket 连接
func ReportWebSocketConnect() error {
	return Report(ReportTypeNodeWebSocketConnect, map[string]interface{}{
		"status": "connected",
	})
}

// ReportWebSocketDisconnect 上报 WebSocket 断开连接
func ReportWebSocketDisconnect() error {
	return Report(ReportTypeNodeWebSocketDisconnect, map[string]interface{}{
		"status": "disconnected",
	})
}

// ReportWebSocketPing 上报 WebSocket ping 结果
func ReportWebSocketPing(request, result interface{}) error {
	return Report(ReportTypeRequestWebSocketPing, map[string]interface{}{
		"request": request,
		"result":  result,
	})
}

// ReportWebSocketTcping 上报 WebSocket tcping 结果
func ReportWebSocketTcping(request, result interface{}) error {
	return Report(ReportTypeRequestWebSocketTcping, map[string]interface{}{
		"request": request,
		"result":  result,
	})
}

// ReportWebSocketWebsite 上报 WebSocket website 结果
func ReportWebSocketWebsite(request, result interface{}) error {
	return Report(ReportTypeRequestWebSocketWebsite, map[string]interface{}{
		"request": request,
		"result":  result,
	})
}

// ReportWebSocketTraceroute 上报 WebSocket traceroute 结果
func ReportWebSocketTraceroute(request, result interface{}) error {
	return Report(ReportTypeRequestWebSocketTraceroute, map[string]interface{}{
		"request": request,
		"result":  result,
	})
}

// ReportWebSocketDns 上报 WebSocket dns 结果
func ReportWebSocketDns(request, result interface{}) error {
	return Report(ReportTypeRequestWebSocketDns, map[string]interface{}{
		"request": request,
		"result":  result,
	})
}

// ReportWebSocketMtr 上报 WebSocket mtr 结果
func ReportWebSocketMtr(request, result interface{}) error {
	return Report(ReportTypeRequestWebSocketMtr, map[string]interface{}{
		"request": request,
		"result":  result,
	})
}

// ReportWebSocketMtrStart 上报 WebSocket mtr 开始
func ReportWebSocketMtrStart(request interface{}) error {
	return Report(ReportTypeRequestWebSocketMtrStart, map[string]interface{}{
		"request": request,
	})
}

// ReportWebSocketMtrComplete 上报 WebSocket mtr 完成
func ReportWebSocketMtrComplete(request, result interface{}) error {
	return Report(ReportTypeRequestWebSocketMtrComplete, map[string]interface{}{
		"request": request,
		"result":  result,
	})
}

// ReportCliPing 上报 CLI ping 结果
func ReportCliPing(request, result interface{}) error {
	return Report(ReportTypeRequestCliPing, map[string]interface{}{
		"request": request,
		"result":  result,
	})
}

// ReportCliTcping 上报 CLI tcping 结果
func ReportCliTcping(request, result interface{}) error {
	return Report(ReportTypeRequestCliTcping, map[string]interface{}{
		"request": request,
		"result":  result,
	})
}

// ReportSystemInfo 上报系统信息
func ReportSystemInfo(systemInfo interface{}) error {
	return Report(ReportTypeSystemInfo, map[string]interface{}{
		"result": systemInfo,
	})
}

// ReportSystem 上报系统信息
func ReportSystem(data interface{}) error {
	return Report(ReportTypeSystem, map[string]interface{}{
		"result": data,
	})
}

// ReportNode 上报节点信息
func ReportNode(data interface{}) error {
	return Report(ReportTypeNode, map[string]interface{}{
		"result": data,
	})
}

// ReportRequest 上报请求信息
func ReportRequest(data interface{}) error {
	return Report(ReportTypeRequest, map[string]interface{}{
		"result": data,
	})
}

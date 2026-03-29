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

// 上报类型常量
const (
	// 系统信息上报类型
	ReportTypeSystem  ReportType = "sys"
	ReportTypeNode    ReportType = "node"
	ReportTypeRequest ReportType = "request"

	// 系统信息上报类型
	ReportTypeSystemInfo ReportType = "sys.system_info"

	// 节点信息上报类型
	ReportTypeNodeStartup             ReportType = "node.startup"
	ReportTypeNodeWebSocketConnect    ReportType = "node.websocket_connect"
	ReportTypeNodeWebSocketDisconnect ReportType = "node.websocket_disconnect"

	// 功能请求上报类型
	ReportTypeRequestPing                 ReportType = "request.ping"
	ReportTypeRequestTcping               ReportType = "request.tcping"
	ReportTypeRequestWebsite              ReportType = "request.website"
	ReportTypeRequestTraceroute           ReportType = "request.traceroute"
	ReportTypeRequestDns                  ReportType = "request.dns"
	ReportTypeRequestMtr                  ReportType = "request.mtr"
	ReportTypeRequestWebSocketPing        ReportType = "request.websocket_ping"
	ReportTypeRequestWebSocketTcping      ReportType = "request.websocket_tcping"
	ReportTypeRequestWebSocketWebsite     ReportType = "request.websocket_website"
	ReportTypeRequestWebSocketTraceroute  ReportType = "request.websocket_traceroute"
	ReportTypeRequestWebSocketDns         ReportType = "request.websocket_dns"
	ReportTypeRequestWebSocketMtr         ReportType = "request.websocket_mtr"
	ReportTypeRequestWebSocketMtrStart    ReportType = "request.websocket_mtr_start"
	ReportTypeRequestWebSocketMtrComplete ReportType = "request.websocket_mtr_complete"
	ReportTypeRequestCliPing              ReportType = "request.cli_ping"
	ReportTypeRequestCliTcping            ReportType = "request.cli_tcping"
)

// ReportData 表示上报数据结构
type ReportData struct {
	Type      ReportType  `json:"type"`
	Timestamp string      `json:"timestamp"`
	NodeID    string      `json:"node_id"`
	Secret    string      `json:"select"`
	Request   interface{} `json:"request,omitempty"`
	Result    interface{} `json:"result,omitempty"`
	Status    string      `json:"status,omitempty"`
	Address   string      `json:"address,omitempty"`
	Version   string      `json:"version,omitempty"`
}

// Report 上报数据
func Report(reportType ReportType, data map[string]interface{}) {
	// 加载配置
	cfg, err := config.LoadConfig(config.GetConfigPath())
	if err != nil {
		// 配置加载失败，跳过上报
		return
	}

	if len(cfg.ReportEndpoints) == 0 {
		// 没有配置上报端点，跳过上报
		return
	}

	// 准备上报数据
	reportData := ReportData{
		Type:      reportType,
		Timestamp: time.Now().Format(time.RFC3339),
		NodeID:    cfg.NodeID,
	}

	// 填充其他字段
	if request, ok := data["request"]; ok {
		reportData.Request = request
	}
	if result, ok := data["result"]; ok {
		reportData.Result = result
	}
	if status, ok := data["status"].(string); ok {
		reportData.Status = status
	}
	if address, ok := data["address"].(string); ok {
		reportData.Address = address
	}
	if version, ok := data["version"].(string); ok {
		reportData.Version = version
	}

	// 序列化数据
	jsonData, err := json.Marshal(reportData)
	if err != nil {
		fmt.Printf("Failed to marshal report data: %v\n", err)
		return
	}

	// 上报到每个端点
	for _, endpoint := range cfg.ReportEndpoints {
		go func(ep string) {
			req, err := http.NewRequest("POST", ep, bytes.NewBuffer(jsonData))
			if err != nil {
				fmt.Printf("Failed to create request to %s: %v\n", ep, err)
				return
			}

			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Node-ID", cfg.NodeID)
			req.Header.Set("X-Secret", cfg.Secret)

			client := &http.Client{Timeout: 10 * time.Second}
			resp, err := client.Do(req)
			if err != nil {
				fmt.Printf("Failed to report to %s: %v\n", ep, err)
				return
			}
			fmt.Println(resp)
			defer resp.Body.Close()
		}(endpoint)
	}
}

// ReportStartup 上报启动日志
func ReportStartup(address, version string) {
	Report(ReportTypeNodeStartup, map[string]interface{}{
		"address": address,
		"version": version,
		"status":  "started",
	})
}

// ReportPing 上报 ping 结果
func ReportPing(request, result interface{}) {
	Report(ReportTypeRequestPing, map[string]interface{}{
		"request": request,
		"result":  result,
	})
}

// ReportTcping 上报 tcping 结果
func ReportTcping(request, result interface{}) {
	Report(ReportTypeRequestTcping, map[string]interface{}{
		"request": request,
		"result":  result,
	})
}

// ReportWebsite 上报 website 结果
func ReportWebsite(request, result interface{}) {
	Report(ReportTypeRequestWebsite, map[string]interface{}{
		"request": request,
		"result":  result,
	})
}

// ReportTraceroute 上报 traceroute 结果
func ReportTraceroute(request, result interface{}) {
	Report(ReportTypeRequestTraceroute, map[string]interface{}{
		"request": request,
		"result":  result,
	})
}

// ReportDns 上报 dns 结果
func ReportDns(request, result interface{}) {
	Report(ReportTypeRequestDns, map[string]interface{}{
		"request": request,
		"result":  result,
	})
}

// ReportMtr 上报 mtr 结果
func ReportMtr(request, result interface{}) {
	Report(ReportTypeRequestMtr, map[string]interface{}{
		"request": request,
		"result":  result,
	})
}

// ReportWebSocketConnect 上报 WebSocket 连接
func ReportWebSocketConnect() {
	Report(ReportTypeNodeWebSocketConnect, map[string]interface{}{
		"status": "connected",
	})
}

// ReportWebSocketDisconnect 上报 WebSocket 断开连接
func ReportWebSocketDisconnect() {
	Report(ReportTypeNodeWebSocketDisconnect, map[string]interface{}{
		"status": "disconnected",
	})
}

// ReportWebSocketPing 上报 WebSocket ping 结果
func ReportWebSocketPing(request, result interface{}) {
	Report(ReportTypeRequestWebSocketPing, map[string]interface{}{
		"request": request,
		"result":  result,
	})
}

// ReportWebSocketTcping 上报 WebSocket tcping 结果
func ReportWebSocketTcping(request, result interface{}) {
	Report(ReportTypeRequestWebSocketTcping, map[string]interface{}{
		"request": request,
		"result":  result,
	})
}

// ReportWebSocketWebsite 上报 WebSocket website 结果
func ReportWebSocketWebsite(request, result interface{}) {
	Report(ReportTypeRequestWebSocketWebsite, map[string]interface{}{
		"request": request,
		"result":  result,
	})
}

// ReportWebSocketTraceroute 上报 WebSocket traceroute 结果
func ReportWebSocketTraceroute(request, result interface{}) {
	Report(ReportTypeRequestWebSocketTraceroute, map[string]interface{}{
		"request": request,
		"result":  result,
	})
}

// ReportWebSocketDns 上报 WebSocket dns 结果
func ReportWebSocketDns(request, result interface{}) {
	Report(ReportTypeRequestWebSocketDns, map[string]interface{}{
		"request": request,
		"result":  result,
	})
}

// ReportWebSocketMtr 上报 WebSocket mtr 结果
func ReportWebSocketMtr(request, result interface{}) {
	Report(ReportTypeRequestWebSocketMtr, map[string]interface{}{
		"request": request,
		"result":  result,
	})
}

// ReportWebSocketMtrStart 上报 WebSocket mtr 开始
func ReportWebSocketMtrStart(request interface{}) {
	Report(ReportTypeRequestWebSocketMtrStart, map[string]interface{}{
		"request": request,
	})
}

// ReportWebSocketMtrComplete 上报 WebSocket mtr 完成
func ReportWebSocketMtrComplete(request, result interface{}) {
	Report(ReportTypeRequestWebSocketMtrComplete, map[string]interface{}{
		"request": request,
		"result":  result,
	})
}

// ReportCliPing 上报 CLI ping 结果
func ReportCliPing(request, result interface{}) {
	Report(ReportTypeRequestCliPing, map[string]interface{}{
		"request": request,
		"result":  result,
	})
}

// ReportCliTcping 上报 CLI tcping 结果
func ReportCliTcping(request, result interface{}) {
	Report(ReportTypeRequestCliTcping, map[string]interface{}{
		"request": request,
		"result":  result,
	})
}

// ReportSystemInfo 上报系统信息
func ReportSystemInfo(systemInfo interface{}) {
	Report(ReportTypeSystemInfo, map[string]interface{}{
		"result": systemInfo,
	})
}

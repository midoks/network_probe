package report

import (
	"encoding/json"
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
	SystemInfo SubType = "info"

	NodeWebSocketConnect    SubType = "websocket_connect"
	NodeWebSocketDisconnect SubType = "websocket_disconnect"

	// 功能请求上报类型
	RequestPing                 SubType = "ping"
	RequestTcping               SubType = "tcping"
	RequestWebsite              SubType = "website"
	RequestTraceroute           SubType = "traceroute"
	RequestDns                  SubType = "dns"
	RequestMtr                  SubType = "mtr"
	RequestStartup              SubType = "startup"
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
	Type      ReportType `json:"type"`
	Timestamp int64      `json:"timestamp"`
	Version   string     `json:"version,omitempty"`
	Data      string     `json:"data,omitempty"`
}

type ReportNodeLogs struct {
	Tag         string `json:"tag,omitempty"`
	Description string `json:"description,omitempty"`
	Level       string `json:"level,omitempty"`
	CreateTime  int64  `json:"create_time,omitempty"`
	Count       int32  `json:"count,omitempty"`
	IsFixed     bool   `json:"is_fixed,omitempty"`
	IsRead      bool   `json:"is_read,omitempty"`
	ParamsJSON  []byte `json:"params_json,omitempty"`
}

type ReportRequestLogs struct {
	Description string `json:"description,omitempty"`
	Level       string `json:"level,omitempty"`
	Count       int32  `json:"count,omitempty"`
	IsFixed     bool   `json:"is_fixed,omitempty"`
	IsRead      bool   `json:"is_read,omitempty"`
	ParamsJSON  []byte `json:"params_json,omitempty"`
	CreateTime  int64  `json:"create_time,omitempty"`
}

// ReportSysInfo 表示系统状态信息
type ReportSysInfo struct {
	DownloadBandwidth string `json:"download_bandwidth"`
	UploadBandwidth   string `json:"upload_bandwidth"`
	Connections       string `json:"connections"`
	AccessRate        string `json:"access_rate"`
	AttackRate        string `json:"attack_rate"`
	MaxDiskWriteSpeed string `json:"max_disk_write_speed"`
	MemoryCacheUsage  string `json:"memory_cache_usage"`
	CPUUsage          string `json:"cpu_usage"`
	MemoryUsage       string `json:"memory_usage"`
	TotalMemory       string `json:"total_memory"`
	Load              string `json:"load"`
	CreateTime        int64  `json:"create_time,omitempty"`
}

func (a *ReportData) SetNodeLogsData(p ReportNodeLogs) error {
	a.Type = ReportTypeNode
	b, err := json.Marshal(p)
	if err != nil {
		return err
	}
	a.Data = string(b)
	return nil
}

func (a *ReportData) SetRequestLogsData(data interface{}) error {
	a.Type = ReportTypeRequest
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}
	a.Data = string(b)
	return nil
}

func (a *ReportData) SetSysInfoData(p ReportSysInfo) error {
	a.Type = ReportTypeSystem
	b, err := json.Marshal(p)
	if err != nil {
		return err
	}
	a.Data = string(b)
	return nil
}

package report

import (
	"encoding/json"
)

// ReportType 表示上报类型
type ReportType string

// 上报类型常量
const (
	ReportTypeSystem   ReportType = "sys"
	ReportTypeNode     ReportType = "node"
	ReportTypeNodeItem ReportType = "node_item"
	ReportTypeRequest  ReportType = "request"
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

type ReportNodeItem struct {
	Item  string `json:"item,omitempty"`
	Value string `json:"value,omitempty"`
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

func (a *ReportData) SetNodeItemData(p ReportNodeItem) error {
	a.Type = ReportTypeNodeItem
	b, err := json.Marshal(p)
	if err != nil {
		return err
	}
	a.Data = string(b)
	return nil
}

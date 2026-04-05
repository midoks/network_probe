package report

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
	"time"

	"network-probe/internal/config"
	"network-probe/internal/version"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
)

type NodeStatusExecutor struct {
	isFirstTime     bool
	lastUpdatedTime time.Time

	cpuLogicalCount  int
	cpuPhysicalCount int

	// 流量统计
	lastUDPInDatagrams  int64
	lastUDPOutDatagrams int64

	ticker *time.Ticker
}

type NodeStatus struct {
	BuildVersion string `json:"build_version"`
	OS           string `json:"os"`
	Arch         string `json:"arch"`
	ExePath      string `json:"exe_path"`
	Hostname     string `json:"hostname"`
	UpdatedAt    int64  `json:"updated_at"`
	Timestamp    int64  `json:"timestamp"`

	//cpu
	CPUUsage         float64 `json:"cpu_usage"`
	CPULogicalCount  int     `json:"cpu_logical_count"`
	CPUPhysicalCount int     `json:"cpu_physical_count"`

	// mem
	MemoryUsage float64 `json:"memory_usage"`
	MemoryTotal uint64  `json:"memory_total"`

	// load
	Load1m  float64 `json:"load1m"`
	Load5m  float64 `json:"load5m"`
	Load15m float64 `json:"load15m"`

	// disk
	DiskUsage             float64 `json:"disk_usage"`
	DiskMaxUsage          float64 `json:"disk_max_usage"`
	DiskMaxUsagePartition string  `json:"disk_max_usage_partition"`
	DiskTotal             uint64  `json:"disk_total"`
	DiskWritingSpeedMB    int     `json:"disk_writing_speed_mb"` // 硬盘写入速度

	IsActive bool   `json:"is_active"`
	Error    string `json:"error"`
}

func NewNodeStatusExecutor() *NodeStatusExecutor {
	return &NodeStatusExecutor{
		ticker: time.NewTicker(30 * time.Second),

		lastUDPInDatagrams:  -1,
		lastUDPOutDatagrams: -1,
	}
}

func (this *NodeStatusExecutor) Run() {
	this.isFirstTime = true
	this.lastUpdatedTime = time.Now()
	this.update()

	for range this.ticker.C {
		this.isFirstTime = false
		this.update()
	}
}

func (this *NodeStatusExecutor) update() {
	// 加载配置
	_, err := config.LoadConfig(config.GetConfigPath())
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		return
	}

	// 构建节点状态
	status := &NodeStatus{
		BuildVersion: version.Version,
		OS:           runtime.GOOS,
		Arch:         runtime.GOARCH,
		IsActive:     true,
		UpdatedAt:    time.Now().Unix(),
		Timestamp:    time.Now().Unix(),
	}

	// 获取可执行文件路径
	status.ExePath, _ = os.Executable()

	// 获取主机名
	hostname, _ := os.Hostname()
	status.Hostname = hostname

	// cpu
	this.updateCPU(status)

	// mem
	this.updateMem(status)

	// load
	this.updateLoad(status)

	// disk
	this.updateDisk(status)

	// 序列化数据
	nodeData, err := json.Marshal(status)
	if err != nil {
		fmt.Printf("Failed to marshal node status: %v\n", err)
		return
	}

	fmt.Println("nodeData:", string(nodeData))

	// 上报数据
	if err := ReportBytes(nodeData); err != nil {
		fmt.Printf("failed to report node status: %v\n", err)
		return
	}

	// 修改更新时间
	this.lastUpdatedTime = time.Now()

	fmt.Printf("node status reported successfully\n")
}

// 更新CPU
func (this *NodeStatusExecutor) updateCPU(status *NodeStatus) {
	var duration = time.Duration(0)
	if this.isFirstTime {
		duration = 100 * time.Millisecond
	}

	percents, err := cpu.Percent(duration, false)
	if err != nil {
		status.Error = "cpu.Percent(): " + err.Error()
		return
	}
	if len(percents) == 0 {
		return
	}

	status.CPUUsage = percents[0] / 100

	NodeItem("cpu", map[string]interface{}{
		"usage": status.CPUUsage,
		"cores": runtime.NumCPU(),
	})

	if this.cpuLogicalCount == 0 && this.cpuPhysicalCount == 0 {
		status.CPULogicalCount, err = cpu.Counts(true)
		if err != nil {
			status.Error = "cpu.Counts(): " + err.Error()
			return
		}
		status.CPUPhysicalCount, err = cpu.Counts(false)
		if err != nil {
			status.Error = "cpu.Counts(): " + err.Error()
			return
		}
		this.cpuLogicalCount = status.CPULogicalCount
		this.cpuPhysicalCount = status.CPUPhysicalCount
	} else {
		status.CPULogicalCount = this.cpuLogicalCount
		status.CPUPhysicalCount = this.cpuPhysicalCount
	}
}

// 更新MEM
func (this *NodeStatusExecutor) updateMem(status *NodeStatus) {
	stat, err := mem.VirtualMemory()
	if err != nil {
		return
	}

	// 重新计算内存
	if stat.Total > 0 {
		stat.Used = stat.Total - stat.Free - stat.Buffers - stat.Cached
		status.MemoryUsage = float64(stat.Used) / float64(stat.Total)
	}

	NodeItem("mem", map[string]interface{}{
		"usage": status.MemoryUsage,
		"total": status.MemoryTotal,
		"used":  stat.Used,
	})

	// 内存严重不足时自动释放内存
	if stat.Total > 0 {
		var minFreeMemory = stat.Total / 8
		if minFreeMemory > 1<<30 {
			minFreeMemory = 1 << 30
		}
		if stat.Available > 0 && stat.Available < minFreeMemory {
			runtime.GC()
			debug.FreeOSMemory()
		}
	}
}

// 更新负载
func (this *NodeStatusExecutor) updateLoad(status *NodeStatus) {
	stat, err := load.Avg()
	if err != nil {
		status.Error = err.Error()
		return
	}
	if stat == nil {
		status.Error = "load is nil"
		return
	}
	status.Load1m = stat.Load1
	status.Load5m = stat.Load5
	status.Load15m = stat.Load15

	// 记录监控数据
	NodeItem("load", map[string]interface{}{
		"load1m":  status.Load1m,
		"load5m":  status.Load5m,
		"load15m": status.Load15m,
	})
}

// 更新硬盘
func (this *NodeStatusExecutor) updateDisk(status *NodeStatus) {
	partitions, err := disk.Partitions(false)
	if err != nil {
		NodeError("NODE_STATUS", err.Error())
		return
	}

	var rootTotal = uint64(0)
	var totalUsed = uint64(0)
	var maxUsage = float64(0)
	// 检查当前系统是否为支持的系统类型
	supportedOS := false
	for _, os := range []string{"darwin", "linux", "freebsd"} {
		if os == runtime.GOOS {
			supportedOS = true
			break
		}
	}

	if supportedOS {
		for _, p := range partitions {
			if p.Mountpoint == "/" {
				usage, _ := disk.Usage(p.Mountpoint)

				fmt.Println(usage, p.Mountpoint)
				if usage != nil {
					rootTotal = usage.Total
					totalUsed = usage.Used
				}
				break
			}
		}
	}
	var total = rootTotal
	status.DiskTotal = total
	if total > 0 {
		status.DiskUsage = float64(totalUsed) / float64(total)
	}
	status.DiskMaxUsage = maxUsage / 100

	NodeItem("disk", map[string]interface{}{
		"total":    status.DiskTotal,
		"usage":    status.DiskUsage,
		"maxUsage": status.DiskMaxUsage,
	})
}

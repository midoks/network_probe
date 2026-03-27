package api

import (
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"network-probe/internal/modules"
)

// parseDuration 将秒数转换为 time.Duration
func parseDuration(seconds int) time.Duration {
	return time.Duration(seconds) * time.Second
}

// Server 表示 API 服务器
type Server struct {
	router *gin.Engine
}

// NewServer 创建一个新的 API 服务器
func NewServer() *Server {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	// 配置 CORS
	router.Use(cors.Default())

	server := &Server{
		router: router,
	}

	// 设置路由
	server.setupRoutes()

	return server
}

// setupRoutes 设置路由
func (s *Server) setupRoutes() {
	// 健康检查
	s.router.GET("/api/health", s.healthCheck)
	s.router.GET("/api/status", s.status)

	// API 路由组
	api := s.router.Group("/api")
	{
		api.POST("/ping", s.handlePing)
		api.POST("/tcping", s.handleTcping)
		api.POST("/website", s.handleWebsite)
		api.POST("/traceroute", s.handleTraceroute)
		api.POST("/dns", s.handleDns)
	}

	// 根路径
	s.router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Network Probe API Server")
	})
}

// healthCheck 健康检查
func (s *Server) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

// status 服务状态
func (s *Server) status(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "running",
		"version": "1.0.0",
	})
}

// handlePing 处理 ping 请求
func (s *Server) handlePing(c *gin.Context) {
	var req struct {
		Host    string `json:"host" binding:"required"`
		Count   int    `json:"count"`
		Timeout int    `json:"timeout"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 设置默认值
	if req.Count == 0 {
		req.Count = 4
	}
	if req.Timeout == 0 {
		req.Timeout = 2
	}

	service := modules.NewPingService()
	config := modules.NewPingConfig()
	config.Host = req.Host
	config.Count = req.Count
	config.Timeout = parseDuration(req.Timeout)

	result, err := service.Ping(config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// handleTcping 处理 tcping 请求
func (s *Server) handleTcping(c *gin.Context) {
	var req struct {
		Host    string `json:"host" binding:"required"`
		Port    int    `json:"port" binding:"required"`
		Count   int    `json:"count"`
		Timeout int    `json:"timeout"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 设置默认值
	if req.Count == 0 {
		req.Count = 4
	}
	if req.Timeout == 0 {
		req.Timeout = 3
	}

	service := modules.NewTcpingService()
	config := modules.NewTcpingConfig()
	config.Host = req.Host
	config.Port = req.Port
	config.Count = req.Count
	config.Timeout = parseDuration(req.Timeout)

	result, err := service.Tcping(config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// handleWebsite 处理 website 请求
func (s *Server) handleWebsite(c *gin.Context) {
	var req struct {
		URL             string `json:"url" binding:"required"`
		Method          string `json:"method"`
		Timeout         int    `json:"timeout"`
		FollowRedirects bool   `json:"follow_redirects"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 设置默认值
	if req.Method == "" {
		req.Method = "GET"
	}
	if req.Timeout == 0 {
		req.Timeout = 30
	}

	service := modules.NewWebsiteTestService()
	config := modules.NewWebsiteTestConfig()
	config.URL = req.URL
	config.Method = req.Method
	config.Timeout = parseDuration(req.Timeout)
	config.FollowRedirects = req.FollowRedirects

	result, err := service.TestWebsite(config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// handleTraceroute 处理 traceroute 请求
func (s *Server) handleTraceroute(c *gin.Context) {
	var req struct {
		Host     string `json:"host" binding:"required"`
		MaxHops  int    `json:"max_hops"`
		Protocol string `json:"protocol"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 设置默认值
	if req.MaxHops == 0 {
		req.MaxHops = 30
	}
	if req.Protocol == "" {
		req.Protocol = "icmp"
	}

	service := modules.NewTracerouteService()
	config := modules.NewTracerouteConfig()
	config.Host = req.Host
	config.MaxHops = req.MaxHops
	config.Protocol = req.Protocol

	result, err := service.Traceroute(config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// handleDns 处理 DNS 请求
func (s *Server) handleDns(c *gin.Context) {
	var req struct {
		Domain     string `json:"domain" binding:"required"`
		QueryType  string `json:"query_type"`
		Nameserver string `json:"nameserver"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 设置默认值
	if req.QueryType == "" {
		req.QueryType = "A"
	}

	var service *modules.DnsService
	var err error

	if req.Nameserver != "" {
		service, err = modules.NewDnsServiceWithNameserver(req.Nameserver)
	} else {
		service = modules.NewDnsService()
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	config := modules.NewDnsConfig()
	config.Domain = req.Domain
	config.QueryType = modules.DnsQueryType(req.QueryType)
	config.Nameserver = req.Nameserver

	result, err := service.Query(config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// Run 启动服务器
func (s *Server) Run(addr string) error {
	return s.router.Run(addr)
}

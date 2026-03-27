# Network Probe

一个功能全面的网络测试工具，使用 Go 语言编写，支持 ping、tcping、traceroute、DNS 查询、网站测试和端口扫描等功能。

## 功能特性

- **ICMP Ping**: 使用 ICMP 回显请求测试网络连通性（需要 root 权限）
- **TCP Ping**: 测试到特定端口的 TCP 连接
- **网站测试**: 测试网站可用性和响应时间
- **Traceroute**: 追踪数据包到达目标的路径
- **DNS 查询**: 执行各种 DNS 记录查询（A、AAAA、CNAME、MX、NS、TXT 等）
- **端口扫描**: 扫描目标主机的开放端口
- **API 服务器**: 启动 Web API 服务器，提供 HTTP 接口进行网络测试

## 项目结构

```
network_probe/
├── main.go                    # 程序入口
├── go.mod                     # Go 模块定义
├── go.sum                     # Go 依赖校验
├── internal/
│   ├── api/
│   │   └── api.go            # API 服务器实现
│   ├── cli/
│   │   └── cli.go            # 命令行接口实现
│   ├── modules/
│   │   ├── ping.go           # ICMP ping 功能
│   │   ├── tcping.go         # TCP ping 功能
│   │   ├── website.go        # 网站测试功能
│   │   ├── traceroute.go     # Traceroute 功能
│   │   └── dns.go            # DNS 查询功能
│   └── utils/
│       └── error.go          # 错误处理工具
├── config/
│   └── api_node.yaml         # 配置文件
└── .github/
    └── workflows/
        └── release.yml       # GitHub Actions 发布工作流
```

## 环境要求

- Go 1.20 或更高版本
- 对于 ICMP ping 和 traceroute：需要 root 权限（sudo）

## 安装

### 从源码构建

1. 克隆仓库：

```bash
git clone https://github.com/your-username/network_probe.git
cd network_probe
```

2. 构建项目：

```bash
go build -o network_probe
```

### 直接运行

```bash
go run .
```

## 使用方法

### 查看帮助

```bash
./network_probe --help
```

### Ping 测试

使用 ICMP 协议测试网络连通性（需要 root 权限）：

```bash
sudo ./network_probe ping www.baidu.com
sudo ./network_probe ping www.baidu.com -c 10 -t 3
```

参数：
- `-c, --count`: 发送的数据包数量（默认：4）
- `-t, --timeout`: 超时时间，单位秒（默认：2）

### TCP Ping 测试

测试到特定端口的 TCP 连接：

```bash
./network_probe tcping example.com -p 80
./network_probe tcping example.com -p 443 -c 10 -t 5
```

参数：
- `-p, --port`: 目标端口（默认：80）
- `-c, --count`: 连接尝试次数（默认：4）
- `-t, --timeout`: 超时时间，单位秒（默认：3）

### 网站测试

测试网站可用性和响应时间：

```bash
./network_probe website https://example.com
./network_probe website https://example.com -m POST -t 60 -f
```

参数：
- `-m, --method`: HTTP 方法（默认：GET）
- `-t, --timeout`: 超时时间，单位秒（默认：30）
- `-f, --follow-redirects`: 跟随重定向

### Traceroute

追踪数据包到达目标的路径（需要 root 权限）：

```bash
sudo ./network_probe traceroute example.com
sudo ./network_probe traceroute example.com -m 20 -p icmp
```

参数：
- `-m, --max-hops`: 最大跳数（默认：30）
- `-p, --protocol`: 协议类型（默认：icmp，可选：icmp, udp, tcp）

### DNS 查询

执行 DNS 记录查询：

```bash
./network_probe dns example.com
./network_probe dns example.com -t MX
./network_probe dns example.com -t A -n 8.8.8.8
```

参数：
- `-t, --query-type`: 查询类型（默认：A，可选：A, AAAA, CNAME, MX, TXT, NS, SOA）
- `-n, --nameserver`: 自定义 DNS 服务器

### 端口扫描

扫描目标主机的开放端口：

```bash
./network_probe port-scan example.com
./network_probe port-scan example.com -r 1-1000 -t 500
```

参数：
- `-r, --range`: 端口范围（默认：1-1000）
- `-t, --timeout`: 超时时间，单位毫秒（默认：1000）

### 启动 API 服务器

启动 Web API 服务器：

```bash
./network_probe server
./network_probe server -H 0.0.0.0 -p 8080
```

参数：
- `-H, --host`: 服务器主机（默认：127.0.0.1）
- `-p, --port`: 服务器端口（默认：8080）

### 安装为系统服务

```bash
sudo ./network_probe install -s network_probe -H 127.0.0.1 -p 8080
```

### 卸载系统服务

```bash
sudo ./network_probe uninstall -s network_probe
```

## API 接口

启动 API 服务器后，可以使用以下 HTTP 接口：

### 健康检查

```bash
curl http://localhost:8080/api/health
```

### 服务状态

```bash
curl http://localhost:8080/api/status
```

### Ping 测试

```bash
curl -X POST http://localhost:8080/api/ping \
  -H "Content-Type: application/json" \
  -d '{"host": "www.baidu.com", "count": 4, "timeout": 2}'
```

### TCP Ping 测试

```bash
curl -X POST http://localhost:8080/api/tcping \
  -H "Content-Type: application/json" \
  -d '{"host": "www.baidu.com", "port": 80, "count": 4, "timeout": 3}'
```

### 网站测试

```bash
curl -X POST http://localhost:8080/api/website \
  -H "Content-Type: application/json" \
  -d '{"url": "https://www.baidu.com", "method": "GET", "timeout": 30, "follow_redirects": true}'
```

### Traceroute

```bash
curl -X POST http://localhost:8080/api/traceroute \
  -H "Content-Type: application/json" \
  -d '{"host": "www.baidu.com", "max_hops": 30, "protocol": "icmp"}'
```

### DNS 查询

```bash
curl -X POST http://localhost:8080/api/dns \
  -H "Content-Type: application/json" \
  -d '{"domain": "www.baidu.com", "query_type": "A", "nameserver": ""}'
```

## 技术栈

- **Go 1.20+**: 编程语言
- **Cobra**: 命令行框架
- **Gin**: Web 框架
- **CORS**: 跨域资源共享中间件

## 权限说明

某些功能需要 root 权限才能正常运行：

- **ICMP Ping**: 需要 root 权限发送 ICMP 数据包
- **Traceroute**: 需要 root 权限发送 ICMP 数据包

如果未使用 root 权限运行这些命令，程序会返回明确的错误提示：

```
Ping failed: ping requires root privileges. Please run with sudo or as root
```

## 开发

### 运行测试

```bash
go test ./...
```

### 代码格式化

```bash
go fmt ./...
```

### 依赖管理

```bash
go mod tidy
```

## 许可证

MIT License

## 贡献

欢迎提交 Issue 和 Pull Request！

## 作者

- [Your Name](https://github.com/your-username)

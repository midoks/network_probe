# Network Probe

一个高性能的 Rust 网络工具集，支持 ping、tcping、网站测试、路由跟踪、DNS 查询，并提供 API、WebSocket 和命令行接口。

## 功能特性

- **ICMP Ping**: 传统 ping 测试，支持自定义包数量和超时时间
- **TCPing**: TCP 连接测试，可测试端口连通性和延迟
- **网站测试**: HTTP/HTTPS 网站可用性测试，支持多种方法
- **路由跟踪**: 追踪数据包到达目标的路由路径
- **DNS 查询**: 支持多种记录类型的 DNS 查询
- **API 接口**: RESTful API 提供所有功能
- **WebSocket**: 实时网络测试接口
- **命令行**: 完整的 CLI 工具支持
- **高性能**: 异步 I/O，支持并发操作

## 安装和构建

```bash
# 克隆项目
cd network_probe

# 构建项目
cargo build --release

# 运行测试
cargo test

# 运行性能基准测试
cargo bench
```

## 命令行使用

### 基本用法

```bash
# 查看帮助
./target/release/network_probe --help

# Ping 测试
./target/release/network_probe ping google.com --count 4

# TCP 端口测试
./target/release/network_probe tcping google.com --port 443 --count 3

# 网站测试
./target/release/network_probe website https://google.com

# DNS 查询
./target/release/network_probe dns google.com --query-type A

# 路由跟踪
./target/release/network_probe traceroute google.com

# 端口扫描
./target/release/network_probe port-scan 127.0.0.1 --range 1-1000
```

### 高级用法

```bash
# 设置日志级别为调试模式
./target/release/network_probe --log-level debug ping localhost

# 自定义超时时间
./target/release/network_probe tcping example.com --port 80 --timeout 5

# 使用特定 DNS 服务器
./target/release/network_probe dns google.com --query-type A --nameserver 8.8.8.8

# 测试多个网站
./target/release/network_probe website https://httpbin.org/status/200 --method GET
```

## API 服务器使用

### 启动服务器

```bash
# 启动 API 服务器（默认端口 8080）
./target/release/network_probe server

# 指定主机和端口
./target/release/network_probe server --host 0.0.0.0 --port 8080
```

### API 端点

#### 健康检查
```bash
curl http://localhost:8080/api/health
```

#### Ping 测试
```bash
curl -X POST http://localhost:8080/api/ping \
  -H "Content-Type: application/json" \
  -d '{"host": "google.com", "count": 4}'
```

#### TCP 连接测试
```bash
curl -X POST http://localhost:8080/api/tcping \
  -H "Content-Type: application/json" \
  -d '{"host": "google.com", "port": 443, "count": 3}'
```

#### 网站测试
```bash
curl -X POST http://localhost:8080/api/website \
  -H "Content-Type: application/json" \
  -d '{"url": "https://google.com", "method": "GET"}'
```

#### DNS 查询
```bash
curl -X POST http://localhost:8080/api/dns \
  -H "Content-Type: application/json" \
  -d '{"domain": "google.com", "query_type": "A"}'
```

#### 路由跟踪
```bash
curl -X POST http://localhost:8080/api/traceroute \
  -H "Content-Type: application/json" \
  -d '{"host": "google.com", "max_hops": 30}'
```

## WebSocket 使用

连接到 `ws://localhost:8080/ws`，可以发送以下格式的消息：

### Ping 测试
```json
{
  "type": "Ping",
  "data": {
    "host": "google.com",
    "count": 4
  }
}
```

### 网站测试
```json
{
  "type": "Website",
  "data": {
    "url": "https://google.com",
    "method": "GET"
  }
}
```

### DNS 查询
```json
{
  "type": "Dns",
  "data": {
    "domain": "google.com",
    "query_type": "A"
  }
}
```

## 性能测试结果

运行 `cargo bench` 的性能测试结果：

- **Ping 测试**: ~470ns（本地回环）
- **TCPing 测试**: ~440μs
- **网站测试**: ~2-8ms
- **DNS 查询**: ~30-70ms
- **并发操作**: ~500μs

## 项目结构

```
network_probe/
├── src/
│   ├── modules/          # 核心功能模块
│   │   ├── ping.rs      # ICMP ping 实现
│   │   ├── tcping.rs    # TCP 连接测试
│   │   ├── website.rs   # 网站测试
│   │   ├── traceroute.rs # 路由跟踪
│   │   └── dns.rs       # DNS 查询
│   ├── api/             # RESTful API 接口
│   ├── websocket/       # WebSocket 接口
│   ├── cli/             # 命令行接口
│   ├── utils/           # 工具函数和错误处理
│   ├── main.rs          # 主程序入口
│   └── lib.rs           # 库入口
├── tests/               # 集成测试
├── benches/             # 性能基准测试
├── Cargo.toml           # 项目配置
└── README.md            # 项目文档
```

## 技术栈

- **异步运行时**: Tokio
- **HTTP 框架**: Axum
- **DNS 解析**: trust-dns-resolver
- **ICMP ping**: surge-ping
- **HTTP 客户端**: reqwest
- **命令行解析**: clap
- **序列化**: serde
- **日志**: env_logger
- **基准测试**: criterion

## 使用场景

- **网络监控**: 持续监控网络服务可用性
- **性能测试**: 测试网络延迟和吞吐量
- **故障排查**: 诊断网络连接问题
- **安全扫描**: 端口扫描和服务发现
- **DNS 验证**: 验证 DNS 配置和解析
- **API 服务**: 提供网络测试能力的微服务

## 开发指南

### 添加新功能

1. 在 `src/modules/` 中创建新的模块
2. 实现相应的配置结构和结果结构
3. 添加到 API 和 CLI 接口
4. 编写单元测试和集成测试
5. 添加性能基准测试

### 运行测试

```bash
# 运行所有测试
cargo test

# 运行特定模块测试
cargo test modules::ping

# 运行集成测试
cargo test --test integration_tests
```

### 性能优化

- 使用异步 I/O 避免阻塞
- 实现连接池复用
- 合理使用并发
- 优化内存分配
- 添加缓存机制

## 许可证

MIT License

## 贡献

欢迎提交 Issue 和 Pull Request 来改进这个项目！
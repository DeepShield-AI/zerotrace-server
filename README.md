# ZeroTrace Server

ZeroTrace 可观测性平台的服务端，负责采集数据接收、处理和存储。

## 系统架构

```
┌─────────────────┐     gRPC + TCP      ┌──────────────────────┐
│  zerotrace-agent │ ──────────────────→ │    zt-server        │
│  (每台宿主机)    │                     │  (控制器+写入+查询)     │
└─────────────────┘                     └──────┬───────────────┘
                                               │
                                     ┌─────────┴─────────┐
                                     │                   │
                              ┌──────┴──────┐    ┌───────┴──────┐
                              │    MySQL    │    │  ClickHouse  │
                              │  (元数据)    │    │  (观测数据)   │
                              └─────────────┘    └──────────────┘
```

## 环境要求

### 软件

| 依赖 | 版本 | 用途 |
|------|------|------|
| Go | 1.26+（见 `go.mod`） | 编译 server |
| protoc | 3.21+ | 编译 protobuf |
| protoc-gen-gofast | 配套 protoc | agent 消息协议编译 |
| protoc-gen-gogo | 配套 protoc | 内部消息协议编译 |
| tmpl | 1.1.0 | 生成代码模板编译（watcher、hmap、fields 等） |
| ujson | 5.x（Python 模块） | IP 地理信息生成 |
| MySQL | 8.0+ | 元数据存储 |
| ClickHouse | 23.8+ | 时序数据存储 |
| make | - | 编译调度 |
| git | - | 版本管理 |

## 本地部署指南

### 1. 安装基础依赖

#### Go

```bash
# 安装 Go（以 1.26 为例，版本以 go.mod 中为准）
wget https://go.dev/dl/go1.26.2.linux-amd64.tar.gz
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go1.26.2.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
echo 'export GOPROXY=https://goproxy.cn,direct' >> ~/.bashrc  # Go 代理
source ~/.bashrc
go version
```

#### protoc 及插件

```bash
# 安装 protoc
sudo apt install -y protobuf-compiler
protoc --version  # 需 ≥ 3.21

# 设置 Go 代理（官方源可能无法访问）
export GOPROXY=https://goproxy.cn,direct
# 可写入 ~/.bashrc 永久生效：
# echo 'export GOPROXY=https://goproxy.cn,direct' >> ~/.bashrc

# 安装 protoc-gen-gofast（用于 agent 消息）
go install github.com/gogo/protobuf/protoc-gen-gofast@latest

# 安装 protoc-gen-gogo（用于内部消息）
go install github.com/gogo/protobuf/protoc-gen-gogo@latest

# 确认 PATH 包含 $GOPATH/bin
export PATH=$PATH:$(go env GOPATH)/bin
```

#### 其他构建工具

```bash
# 安装 tmpl（生成代码模板工具，watcher、hmap、fields 等生成依赖）
go install github.com/benbjohnson/tmpl@v1.1.0

# 安装 ujson（IP 地理信息生成依赖）
pip install ujson
# 若 pip 提示 externally-managed-environment（PEP 668），改用：
# pip install --break-system-packages ujson
```

### 2. 克隆代码

```bash
git clone --recurse-submodules https://github.com/DeepShield-AI/zerotrace-server.git
cd zerotrace-server
```

### 3. 启动 MySQL 和 ClickHouse

数据库通过 Docker Compose 启动，无需本地安装：

```bash
# 创建数据目录
sudo mkdir -p /opt/zt/{mysql,clickhouse,clickhouse_storage}

# 添加数据库初始化信息的路径
- mkdir -p /etc/metadb/schema/rawsql/mysql/
- sudo cp -r controller/db/metadb/migrator/schema/rawsql/mysql/* /etc/metadb/schema/rawsql/mysql/
- sudo cp -r querier/db_descriptions /etc/  
- sudo chmod -R 755 /etc/db_descriptions

# 启动数据库服务（仅 mysql 和 clickhouse）
cd manifests/docker-compose
docker compose up -d mysql clickhouse algorithms
cd ../..

# 验证
docker ps | grep -E "zt-mysql|zt-clickhouse"
curl http://localhost:8123/ping  # ClickHouse HTTP 接口
```

> 端口映射：MySQL `3306:30130`（宿主机 3306 → 容器 30130）、ClickHouse `9000:9000` / `8123:8123`。

### 4. 编译

```bash
export PATH=$PATH:$(go env GOPATH)/bin && make server
```

该命令会自动执行：

1. `go mod tidy && go mod download && go mod vendor` — 下载依赖
2. 应用 vendor 补丁（优化 ClickHouse 写入性能、hashmap 等）
3. 编译 protobuf（agent/controller/k8s 等消息协议）
4. 编译生成代码（watcher、hmap 等）
5. `go build` 输出到 `bin/deepflow-server`

### 5. 配置

编辑 `server.yaml`，关键配置如下：

```yaml
# 日志
log-file: /var/log/deepflow/server.log
log-level: info

controller:
  # agent 控制面 gRPC 端口
  grpc-port: 30035
  # HTTP 管理端口
  listen-port: 20417

  # 本机 IP（必须设置，否则 agent 无法获取控制器 IP）
  node-ip: x.x.x.x            # ← 替换为实际宿主机 IP

  # MySQL 配置
  mysql:
    enabled: true
    database: deepflow
    user-name: root
    user-password: deepflow
    host: 127.0.0.1
    port: 3306

  # ClickHouse 配置
  clickhouse:
    database: flow_tag
    user-name: default
    host: 127.0.0.1
    port: 9000

  # 数据节点地址（必须设置，agent 通过此地址连接 ingester）
  trisolaris:
    tsdb-ip: x.x.x.x           # ← 替换为实际宿主机 IP

  # Redis（可选，用于缓存，可关闭）
  redis:
    enabled: false

ingester:
  # 数据接收端口
  listen-port: 30033
  # 本机 IP（必须设置）
  node-ip: x.x.x.x            # ← 替换为实际宿主机 IP
  controller-ips:
    - 127.0.0.1
  controller-port: 30035

  ckdb:
    host: 127.0.0.1
    port: 9000

querier:
  # 查询 API 端口
  listen-port: 20416
  clickhouse:
    host: 127.0.0.1
    port: 9000
```

### 6. 初始化数据库

Server 首次启动时会自动在 MySQL 中创建 `deepflow` 数据库和所有表结构。

### 7. 启动

```bash
# 确保数据库已启动（如未启动）
cd manifests/docker-compose && docker compose up -d mysql clickhouse && cd ../..

# 启动 server
K8S_NODE_NAME_FOR_DEEPFLOW=$(hostname) K8S_NODE_IP_FOR_DEEPFLOW=$(hostname -I | awk '{print $1}') DEEPFLOW_SERVER_RUNNING_MODE=STANDALONE ./bin/deepflow-server -f ./server.yaml

# 后台启动
K8S_NODE_NAME_FOR_DEEPFLOW=$(hostname) K8S_NODE_IP_FOR_DEEPFLOW=$(hostname -I | awk '{print $1}') DEEPFLOW_SERVER_RUNNING_MODE=STANDALONE nohup ./bin/deepflow-server -f ./server.yaml > deepflow-server.log 2>&1 &
```

### 8. 验证

```bash
# 检查 server 进程
ps aux | grep deepflow-server

# 健康检查
curl http://localhost:20417/v1/health/

# ClickHouse 可用性
curl http://localhost:8123/ping

## 编译说明

### Makefile 目标

| 目标 | 说明 |
|------|------|
| `make all` | 编译 server（默认） |
| `make server` | 编译 server 二进制 |
| `make vendor` | 下载依赖 + 打 vendor 补丁 |
| `make test` | 运行测试 |
| `make clean` | 清理编译产物 |
| `make proto` | 编译所有 .proto 文件 |

### Vendor 补丁

`make vendor` 会应用以下补丁，优化运行时性能：

| 补丁 | 位置 | 作用 |
|------|------|------|
| clickhouse-go/write_improve.patch | `patch/clickhouse-go/` | 优化 Array(string) 写入性能 |
| clickhouse-go/nullable.patch | `patch/clickhouse-go/` | 修复 nullable 字段指针重置 |
| clickhouse-go/datetime_uint32.patch | `patch/clickhouse-go/` | 优化 datetime 写入 |
| clickhouse-go/reduce_connection_memory.patch | `patch/clickhouse-go/` | 减少连接内存占用 |
| cornelk_hashmap/perf.patch | `patch/cornelk_hashmap/` | 优化大批量写入时 hashmap 性能 |
| cornelk_hashmap/complex128.patch | `patch/cornelk_hashmap/` | hashmap 支持 [2]uint64 key |

### 代码生成

以下代码由 `go generate` 自动生成，修改模板后需重新生成：

- `libs/hmap/idmap/ubig_id_map.go` — ID 映射表
- `libs/hmap/lru/ubig_lru.go` — LRU 缓存
- `libs/flow-metrics/pooled_meters.go` — 指标池
- `libs/kubernetes/watcher.gen.go` — K8s Watcher
- `libs/geo/ip_info.go` — IP 地理信息

```bash
# 重新生成所有生成代码
go generate ./...
```

> 注：proto 生成（`flow_log.pb.go`、`metric.pb.go`、`stats.pb.go`）的 include 路径指向 `vendor/`（`-I=../../../vendor`），依赖 `gogo.proto` 从 vendor 目录解析，无需 GOPATH/src 布局。

## 端口参考

| 端口 | 模块 | 协议 | 用途 |
|------|------|------|------|
| 30033 | ingester | TCP | agent 数据写入 |
| 30035 | controller | gRPC | agent 控制面同步 |
| 20416 | querier | HTTP | 数据查询 API |
| 20417 | controller | HTTP | 管理 API |
| 20419 | profile | HTTP | 性能分析 |
| 9000 | ClickHouse | TCP | 数据库原生协议 |
| 8123 | ClickHouse | HTTP | 数据库 HTTP 接口 |
| 3306 | MySQL | TCP | 数据库连接 |

## 项目结构

```
├── cmd/server/           # 入口 main
├── controller/           # 控制器（gRPC、HTTP API、agent 管理）
│   ├── grpc/             #   gRPC 服务
│   ├── monitor/          #   健康检查、VTap 生命周期管理
│   └── trisolaris/       #   核心调度逻辑
├── ingester/             # 数据写入（ClickHouse）
├── querier/              # 数据查询
├── libs/                 # 公共库
│   ├── datatype/pb/      #   数据类型 proto
│   ├── flow-metrics/pb/  #   指标 proto
│   └── stats/pb/         #   统计 proto
├── message/              # protobuf 消息定义
├── patch/                # vendor 补丁
├── server.yaml           # 服务端配置
└── Makefile              # 编译
```

## 故障排查

| 现象 | 原因 | 解决 |
|------|------|------|
| agent 日志 `no ingester_ip` | server.yaml 缺少 `tsdb-ip` | 设置 `trisolaris.tsdb-ip` 为宿主机 IP |
| agent 日志 `http sync: 404` | 未设置 `ZT_DATA_VIA_HTTP=false` | 启动 agent 时加上环境变量 |
| agent 连接超时 | 防火墙/安全组拦截 | 确保 30033、30035 端口可达 |
| ClickHouse 表为空 | agent 未启动或配置错误 | 检查 agent 日志和 server 日志 |
| server 日志 `controller ip() is invalid` | `node-ip` 未设置 | 在 `controller` 和 `ingester` 段设置 `node-ip` |
| server 日志 `delete vtap: worker1-F0` | host_device 表缺少本机记录 | 更新代码重新编译，fallback 已支持自动创建 |
| agent 启动失败 `CAP_SYS_ADMIN` | 缺少 root 权限 | 使用 `sudo` 启动 agent |
| `go mod vendor` 报错 | 网络或版本问题 | 检查 `go.mod` 中版本号，确保 Go 版本匹配 |
| `make server` 报 `exec: "tmpl": executable file not found` | 未安装 tmpl | `go install github.com/benbjohnson/tmpl@v1.1.0` |
| `make server` 报 `ModuleNotFoundError: No module named 'ujson'` | 未安装 ujson | `pip install ujson`（PEP 668 环境加 `--break-system-packages`） |

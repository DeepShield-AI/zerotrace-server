# ZeroTrace Server

ZeroTrace 可观测性平台的服务端，负责采集数据接收、处理和存储。

## 系统架构

```
┌─────────────────┐     gRPC + TCP      ┌──────────────────────┐
│  zerotrace-agent │ ──────────────────→ │    zt-server         │
│  (每台宿主机)    │                     │  (控制器+写入+查询)  │
└─────────────────┘                     └──────┬───────────────┘
                                               │
                                     ┌─────────┴─────────┐
                                     │                   │
                              ┌──────┴──────┐    ┌───────┴──────┐
                              │  zt-mysql   │    │ zt-clickhouse │
                              │  (元数据)   │    │  (时序数据)   │
                              └─────────────┘    └──────────────┘
```

## 环境要求

| 依赖 | 最低版本 | 说明 |
|------|---------|------|
| Docker | 24.0+ | 容器运行时 |
| Docker Compose | v2.20+ | 服务编排 |
| Linux | x86_64 / aarch64 | 部署宿主机 |
| 内存 | 8 GB+ | 建议 16 GB |
| 磁盘 | 50 GB+ | 数据持久化 |

## 快速部署

### 1. 准备宿主机

```bash
# 创建数据目录
sudo mkdir -p /opt/zt/{mysql,clickhouse,clickhouse_storage}

# 确认 Docker 可用
docker --version && docker compose version
```

### 2. 配置环境变量

编辑 `manifests/docker-compose/.env`：

```bash
# DeepFlow 版本（目前固定 v7.0）
DEEPFLOW_VERSION=v7.0

# 宿主机 IP——agent 通过此 IP 连接 server
NODE_IP_FOR_DEEPFLOW=<替换为宿主机IP>
```

### 3. 配置 server.yaml

编辑 `manifests/docker-compose/common/config/server/server.yaml`：

```yaml
controller:
  grpc-port: 20035            # agent 控制面 gRPC 端口
  listen-port: 20417           # HTTP 管理端口
  mysql:
    host: mysql                # Docker 内用 service 名
    port: 30130                # MySQL 端口（容器内）
    user-name: root
    user-password: deepflow
  clickhouse:
    host: clickhouse
    port: 9000
  trisolaris:
    tsdb-ip: <宿主机IP>         # ★ 必须设为宿主机 IP
ingester:
  ckdb:
    host: clickhouse
    port: 9000
querier:
  listen-port: 20416           # 查询 API 端口
```

> ⚠️ `trisolaris.tsdb-ip` 必须设为宿主机 IP，否则 agent 无法获取 ingester 地址。

### 4. 启动

```bash
cd manifests/docker-compose
docker compose up -d
```

### 5. 验证

```bash
# 容器状态
docker ps | grep zt-

# 预期输出：
# zt-server      Up   0.0.0.0:20416->20416/tcp, 0.0.0.0:30033->20033/tcp, ...
# zt-clickhouse  Up   0.0.0.0:8123->8123/tcp, 0.0.0.0:9000->9000/tcp
# zt-mysql       Up   0.0.0.0:3306->30130/tcp

# 服务健康检查
curl http://localhost:30417/v1/health/
# → {"OPT_STATUS":"SUCCESS","DESCRIPTION":"","DATA":null}

# ClickHouse 可用性
curl http://localhost:8123/ping
# → Ok.
```

## 端口参考

| 容器端口 | 宿主机端口 | 模块 | 协议 | 用途 |
|---------|-----------|------|------|------|
| 20033 | 30033 | ingester | TCP | agent 数据写入 |
| 20035 | 30035 | controller | gRPC | agent 控制面同步 |
| 20416 | 20416 | querier | HTTP | 数据查询 API |
| 20417 | 30417 | controller | HTTP | 管理 API |
| 20419 | 20419 | profile | HTTP | 性能分析 |
| 9000 | 9000 | ClickHouse | TCP | 数据库原生协议 |
| 8123 | 8123 | ClickHouse | HTTP | 数据库 HTTP 接口 |
| 30130 | 3306 | MySQL | TCP | 数据库连接 |

## 常用运维

```bash
# 查看日志
docker logs zt-server -f --tail 100

# 重启
cd manifests/docker-compose && docker compose restart

# 停止
cd manifests/docker-compose && docker compose down

# 升级（修改 .env 版本号后）
cd manifests/docker-compose && docker compose pull && docker compose up -d

# 数据检查
docker exec zt-clickhouse clickhouse-client \
  --query "SELECT count() FROM flow_log.l7_flow_log"

docker exec zt-mysql mysql -uroot -pdeepflow \
  -e "SELECT id,name,ip,hostname FROM deepflow.vtap"
```

## 配置文件一览

```
manifests/docker-compose/
├── .env                          # 版本号 + 宿主机 IP
├── docker-compose.yaml           # 容器编排
└── common/config/
    ├── server/server.yaml        # zt-server 核心配置
    ├── clickhouse/config.xml     # ClickHouse 引擎配置
    ├── clickhouse/users.xml      # 数据库用户权限
    └── mysql/
        ├── my.cnf                # MySQL 引擎配置
        └── init.sql              # 首次初始化 SQL
```

## 故障排查

| 现象 | 原因 | 解决 |
|------|------|------|
| agent 日志 `no ingester_ip` | server.yaml 缺少 `tsdb-ip` | 设置 `trisolaris.tsdb-ip` 为宿主机 IP |
| agent 日志 `http sync: 404` | 未设置 `ZT_DATA_VIA_HTTP=false` | 启动 agent 时加上环境变量 |
| agent 连接超时 | 防火墙/安全组拦截 | 确保 30033、30035 端口可达 |
| ClickHouse 表为空 | agent 未启动或配置错误 | 检查 agent 日志和 server 日志 |
| server 日志 `no proxy_controller_ip` | NAT IP 配置问题 | 检查 `server.yaml` 中 controller IP 相关配置 |

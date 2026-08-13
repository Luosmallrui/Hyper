# Hyper

一个基于 Go 的统一后端服务，为同一平台的 **四个客户端** 提供 HTTP API：

| 客户端 | 说明 |
|---|---|
| 小程序端 | C 端用户（票务、订单、笔记、关注、消息等） |
| PC 网页端 | Web 客户端 |
| 商家端 | 主办方/组织者门户（票务、提现、退款、门店、核销） |
| 管理端 | 运营后台（`admin_pc`） |

四个端共用一套代码和 API 面，通过路由分组 + 中间件隔离权限（管理端路由与鉴权单独拆分在 `handler/admin.go` 与 `middleware/admin.go`）。

## 技术栈

- **Web 框架**：Gin
- **数据访问**：GORM + MySQL（`dao/`，含 Redis 缓存层 `dao/cache`）
- **缓存 / 锁**：Redis
- **消息队列**：RocketMQ
- **服务发现**：Nacos
- **RPC**：CloudWeGo Kitex（thrift `push` 服务，IM 推送）
- **长连接**：独立 WebSocket / TCP 连接服务（`socket/`）
- **鉴权**：JWT
- **支付**：微信支付 API v3
- **对象存储**：阿里云 OSS
- **AI**：OpenAI SDK（`pkg/llm`，默认走 DashScope 兼容端点）
- **依赖注入**：Google Wire

## 目录结构

```
Hyper/
├── cmd/
│   ├── api-server/        # HTTP API 服务入口 (main + wire 注入)
│   └── conn-server/       # 长连接服务入口 (WebSocket/TCP + Kitex RPC)
├── handler/               # Gin HTTP handler，按域一文件，实现 RegisterRouter(r gin.IRouter)
├── service/               # 业务逻辑，按域一文件，通过接口 + wire.Bind 注入
├── dao/                   # GORM 数据访问
│   └── cache/             # Redis 缓存封装
├── models/                # GORM 模型
├── types/                 # 请求/响应 DTO
├── config/                # 配置结构体 (app/db/redis/jwt/oss/nacos/rocketmq/pay/llm) + table.sql
├── configs/               # 配置文件模板 config.{env}.yaml
├── middleware/            # auth / admin auth / 日志 / prometheus
├── pkg/                   # 共享库: jwt, oss, snowflake, encrypt, timewheel, socket,
│                          #        llm, rocketmq, nacos, response, database, log ...
├── rpc/                   # Kitex RPC (push.thrift + kitex_gen 生成代码)
├── socket/                # WebSocket/TCP 连接服务 (handler/process/router)
├── docs/                  # 按日期记录的接口变更说明（前后端契约，改动接口前先查这里）
└── script/                # 脚本
```

> 说明：代码采用**扁平经典分层**（handler → service → dao → models），并非 README 曾描述的 DDD `internal/domain` + `internal/module` 结构。

## 快速开始

### 环境要求

- Go 1.24+
- 已配置的 MySQL、Redis、RocketMQ、Nacos（见 `configs/config.{env}.yaml`）

### 运行

```shell
# HTTP API 服务
APP_ENV=dev go run ./cmd/api-server/. serve

# 长连接服务（WebSocket/TCP + Kitex RPC）
APP_ENV=dev go run ./cmd/conn-server/.
```

`APP_ENV` 未设置时默认 `dev`，对应加载 `configs/config.dev.yaml`。

### 构建

```shell
make build        # 构建 api-server 与 conn-server 到 bin/
make gen          # 重新生成 Wire 依赖注入代码
```

### 测试

```shell
make test         # go test ./... -race
```

## Makefile 常用命令

| 命令 | 说明 |
|---|---|
| `make gen` / `make gen-api` / `make gen-conn` | 生成 Wire 注入代码 |
| `make build` / `make build-api` / `make build-conn` | 交叉编译到 `bin/` |
| `make run-api` / `make run-conn` | 本地运行 |
| `make test` | 跑测试 |
| `make clean` | 清理 `bin/` |

## 接口文档

对外接口契约（含四个端的 API 变更）以 `docs/` 下的日期命名 Markdown 为准，另见：

- `square_API_DOC.md` — 广场相关接口
- `ticketing_api.md` / `im_api.md` 等 — 票务、IM 等域接口

所有 REST 接口统一挂在 `/api/v1` 下，健康检查为 `GET /`，Prometheus 指标为 `GET /metrics`。

## 配置

配置结构体见 `config/config.go`，主要配置项：

```yaml
app:        # 应用、微信小程序 app_id/secret
jwt:        # JWT 密钥与过期时间
server:     # http / websocket / rpc 端口
mysql:      # 数据库连接
redis:      # Redis 连接
oss:        # 阿里云 OSS
nacos:      # 服务发现
rocketmq:   # 消息队列
wechat_pay: # 微信支付
llm:        # 大模型 API
```

数据库表结构见 `config/table.sql`。

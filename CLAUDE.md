# CLAUDE.md — Hyper Backend

This file provides guidance for AI assistants working in this repository.

---

## Project Overview

**Hyper** is a production-grade Go backend for a social/e-commerce platform. It includes:

- REST API server (HTTP)
- Real-time WebSocket connection server
- RPC push service (Thrift via Kitex)
- Async message processing (RocketMQ)

**Go version:** 1.24.4
**Module name:** `Hyper`

---

## Repository Structure

```
/cmd/
  api-server/       # HTTP REST API entry point
  conn-server/      # WebSocket connection server entry point
  fanout-server/    # Message fanout server (currently disabled)

/config/            # Configuration types, YAML loading, DB/Redis/OSS/etc. setup
  config.go         # Config struct and loader
  table.sql         # Database schema reference

/handler/           # Gin HTTP handlers (one file per domain)
/service/           # Business logic layer (interfaces + implementations)
/dao/               # Data access objects; /dao/cache/ for Redis-backed DAOs
/models/            # GORM model definitions
/middleware/        # Gin middleware: auth.go, log.go, prometheus.go
/pkg/               # Reusable utility packages (see below)
/rpc/               # Kitex RPC server and Thrift-generated code
/socket/            # WebSocket server logic (router/, process/, handler/)
/types/             # Shared custom type definitions
/docs/              # API documentation (Markdown)
```

### `/pkg` packages

| Package      | Purpose                                  |
|--------------|------------------------------------------|
| `jwt`        | JWT generation and parsing               |
| `response`   | Standardized Gin response helpers        |
| `log`        | Zap-based structured logger              |
| `context`    | Gin context helpers (`GetUserID`, etc.)  |
| `database`   | GORM DB initialization                   |
| `encrypt`    | Hashing/encryption utilities             |
| `jsonutil`   | JSON helpers (uses `gjson`)              |
| `llm`        | OpenAI API client wrapper                |
| `nacos`      | Nacos service discovery/config client    |
| `oss`        | Alibaba OSS file upload/download         |
| `rocketmq`   | RocketMQ producer/consumer wrappers      |
| `snowflake`  | Distributed ID generation                |
| `socket`     | WebSocket client abstraction             |
| `strutil`    | String utilities                         |
| `timeutil`   | Time formatting helpers                  |
| `timewheel`  | Timer wheel implementation               |
| `utils`      | Miscellaneous helpers                    |

---

## Common Commands

```bash
# Generate Wire DI code (required before building)
make gen          # all servers
make gen-api      # API server only
make gen-conn     # connection server only

# Build binaries (output: /bin/)
make build        # both servers
make build-api    # API server only
make build-conn   # connection server only

# Run in development
make run-api      # starts HTTP API server
make run-conn     # starts WebSocket server

# Test
make test         # go test ./... -race

# Clean
make clean        # removes /bin/
```

---

## Architecture Patterns

### Layered Architecture

```
HTTP Handler (handler/)
    ↓
Service Interface + Implementation (service/)
    ↓
DAO / Cache (dao/, dao/cache/)
    ↓
Database (MySQL via GORM) / Cache (Redis)
```

### Dependency Injection (Google Wire)

Each server has a `wire.go` (provider declarations) and a `wire_gen.go` (generated bootstrap). Run `make gen` after modifying providers. Never hand-edit `wire_gen.go`.

### Response Format

All HTTP responses use the `pkg/response` helpers:

```go
response.Success(c, data)           // 200 with payload
response.Fail(c, code, msg)         // HTTP 200, business error code
response.Abort(c, httpStatus, msg)  // HTTP error, aborts middleware chain
```

The `Response` struct:
```go
type Response struct {
    Code int         `json:"code"`
    Msg  string      `json:"msg"`
    Data interface{} `json:"data,omitempty"`
}
```

### Authentication Middleware

`middleware.Auth(secret)` validates JWT Bearer tokens and sets `user_id` (int) and `openid` (string) in the Gin context.

**Debug mode:** Sending `Authorization: Bearer debug-mode` bypasses JWT validation and sets `user_id=1`. Only valid when the server is in debug mode.

Token auto-refresh: if the access token expires in fewer than 20 seconds, a new token is returned in the `X-New-Access-Token` response header.

### Error Handling

- Services return `(result, error)`.
- Use `*response.BizError` for domain errors with specific codes.
- The `response.ErrorMiddleware()` Gin middleware catches `c.Error()`-appended errors and converts them to JSON responses.
- Do not use `panic` for control flow; only for initialization failures.

---

## Configuration

Config files live at `configs/config.{env}.yaml` (excluded from git). The `Config` struct in `config/config.go` covers:

| Section       | Purpose                         |
|---------------|---------------------------------|
| `app`         | Name, debug flag, environment   |
| `server`      | HTTP, WebSocket, TCP, RPC ports |
| `mysql`       | Database DSN and pool settings  |
| `redis`       | Redis address and credentials   |
| `jwt`         | Token secrets and expiry        |
| `oss`         | Alibaba OSS bucket settings     |
| `nacos`       | Service discovery config        |
| `rocketmq`    | Message queue endpoints         |
| `wechat_pay`  | WeChat Pay API credentials      |

Config is loaded at startup via `config.New(filename)`. Never commit real config files.

---

## Key Dependencies

| Library                        | Role                              |
|--------------------------------|-----------------------------------|
| `github.com/gin-gonic/gin`     | HTTP router and middleware        |
| `github.com/gorilla/websocket` | WebSocket protocol                |
| `github.com/cloudwego/kitex`   | RPC framework (Thrift)            |
| `gorm.io/gorm`                 | ORM for MySQL                     |
| `github.com/redis/go-redis/v9` | Redis client                      |
| `github.com/golang-jwt/jwt/v5` | JWT auth                          |
| `github.com/google/wire`       | Compile-time dependency injection |
| `go.uber.org/zap`              | Structured logging                |
| `github.com/openai/openai-go`  | LLM integration                   |
| `github.com/apache/rocketmq-*` | Message queue (v2 and v5 clients) |
| `github.com/bwmarrin/snowflake`| Distributed ID generation         |
| `gopkg.in/yaml.v3`             | YAML config parsing               |

---

## Code Conventions

### Naming

- Exported identifiers: `PascalCase`
- Unexported identifiers: `camelCase`
- Service interfaces: prefixed with `I` (e.g., `IUserService`, `INoteService`)
- Handler structs: suffixed with `Handler` (e.g., `UserHandler`)

### Service Layer

Define an interface and implement it in the same file or a paired `_impl.go` file. Register the implementation with Wire via a provider function.

### Logging

Use the global logger at `pkg/log`:
```go
import "Hyper/pkg/log"
log.L.Info("message", zap.String("key", "value"))
log.L.Error("failure", zap.Error(err))
```

### Context Helpers

Extract request-scoped values from `*gin.Context` via `pkg/context`:
```go
userID := ctx.GetUserID(c)   // returns int
```

### Models

GORM models in `models/`. Prefer soft deletes where applicable (`gorm.Model` embeds `DeletedAt`).

### Comments

Much of the existing codebase uses Chinese comments. New code may use either English or Chinese; be consistent within a file.

---

## Testing

- Test files follow Go conventions: `*_test.go` alongside source files.
- Use `github.com/stretchr/testify` for assertions.
- Use `github.com/golang/mock` or `go.uber.org/mock` for mocking service interfaces.
- Run all tests with race detection: `make test`.

---

## What NOT to Do

- Do not hand-edit `wire_gen.go`; always regenerate via `make gen`.
- Do not commit `configs/config.*.yaml` (gitignored).
- Do not commit `bin/` artifacts (gitignored).
- Do not use `panic` in handler or service code; return errors instead.
- Do not bypass the service layer by calling DAO methods directly from handlers.
- Do not add unnecessary abstractions; follow the existing flat-file-per-domain pattern.

---

## External Services

The platform integrates with:

- **WeChat** — OAuth login (OpenID) and WeChat Pay
- **Alibaba OSS** — File/media storage
- **Nacos** — Service discovery and dynamic config
- **RocketMQ** — Async event processing
- **OpenAI** — LLM-powered features
- **Google APIs** — Various integrations

# go-clean-arch_demo

**v2 企业级架构演示**（feat/v2-enterprise-stack 分支）—— Gin + GORM + 轻量 DDD + JWT + Prometheus + Swagger + gRPC

> **从 [bxcodec/go-clean-arch](https://github.com/bxcodec/go-clean-arch) fork 并全面重构**

## 技术栈

| 层 | 选型 |
|---|---|
| 语言 | Go 1.23+ |
| Web 框架 | Gin v1.10 |
| ORM | GORM v2 + go-sql-driver/mysql |
| 配置 | Viper v1.19（YAML + 环境变量） |
| 日志 | uber-go/zap v1.27 |
| 鉴权 | golang-jwt/jwt v5 + bcrypt |
| gRPC | google.golang.org/grpc v1.66 |
| API 文档 | swaggo/swag v1.16 + Swagger UI |
| Metrics | prometheus/client_golang v1.20 |
| 校验 | go-playground/validator/v10 |
| 限流 | golang.org/x/time/rate（token bucket） |
| 熔断 | sony/gobreaker |
| 迁移 | golang-migrate v4.17 |
| 热重载 | cosmtrek/air |
| 静态检查 | golangci-lint v1.61 |
| 容器 | distroless/static-debian12（多阶段构建） |

## 快速开始

> **第一次使用？** 复制 example 文件作为本地配置起点：
> ```bash
> cp .env.example .env            # 被 godotenv 自动加载，覆盖 config.yaml
> cp configs/config.local.yaml.example configs/config.local.yaml
> # 编辑 .env 和 config.local.yaml 填入你的本地值
> ```
>
> 加载优先级：`config.yaml` < `.env` < `config.<APP_ENV>.yaml` < shell `APP_*` env
```bash
# 1. 安装工具链
make install-tools

# 2. 启动 MySQL
make dev-env

# 3. 应用数据库迁移
make migrate-up

# 4. 生成 Swagger 文档（首次运行）
make swagger

# 5. 启动服务（开发模式热重载）
make dev-air

# 或直接构建运行
make build
./bin/rest    # REST API on :9090
./bin/grpc    # gRPC API on :9091
```

## 📑 文档目录

- [docs/v2-design.md](./docs/v2-design.md) — **完整设计文档**（1854 行，含 17 章）
- [docs/architecture.md](./docs/architecture.md) — 分层架构 + 依赖方向 + 关键决策
- [docs/api.md](./docs/api.md) — REST + gRPC API 完整文档
- [docs/infrastructure.md](./docs/infrastructure.md) — Makefile / Docker / air / golangci 等详解

## 架构概览

```
cmd/           REST + gRPC 入口
internal/
  interfaces/  HTTP (Gin) + gRPC handler（适配层）
  application/ 业务用例（article / author / auth / user）
  domain/      实体 + Repository 接口（最内层）
  infrastructure/  GORM 实现 + Viper + migrations
pkg/           公共工具（logger / errcode / response / jwt / server / ratelimit / circuitbreaker）
configs/       多环境 YAML 配置
deployments/   Dockerfile + docker-compose.yaml
api/proto/     .proto 源文件
```

依赖规则：**外圈依赖内圈，内圈不知道外圈存在**。详细分层和依赖图见 `docs/architecture.md`。

## REST API 概览

```
POST   /api/v1/auth/login              登录获取 token
POST   /api/v1/auth/refresh            刷新 access token
POST   /api/v1/auth/register           注册新用户

POST   /api/v1/authors                 创建作者
GET    /api/v1/authors                 列出作者
GET    /api/v1/authors/:id             获取作者
PUT    /api/v1/authors/:id             更新作者
DELETE /api/v1/authors/:id             删除作者

POST   /api/v1/articles                创建文章
GET    /api/v1/articles                列表（cursor 分页）
GET    /api/v1/articles/:id            获取文章
PUT    /api/v1/articles/:id            更新文章
DELETE /api/v1/articles/:id            删除文章

GET    /healthz                        存活探针
GET    /readyz                         就绪探针（DB ping）
GET    /metrics                        Prometheus 指标
GET    /swagger/index.html             Swagger UI
GET    /debug/pprof/*                  pprof（仅 debug 模式）
```

完整 curl 示例见 `docs/api.md`。

## v3 TODO（暂未实现）

- API Gateway（Nginx / Kong / BFF）
- Redis 缓存 / 分布式限流
- 审计日志
- 软删除（gorm.DeletedAt）
- 集成测试（testcontainers）
- OAuth2 / OIDC
- OpenTelemetry 全栈
- 消息队列
- K8s / Helm
- 详细见 `docs/v2-design.md` 第十六章

---

Forked from [bxcodec/go-clean-arch](https://github.com/bxcodec/go-clean-arch).

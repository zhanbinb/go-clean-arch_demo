# V2 企业级架构升级方案

> **目标**：在 `feat/v2-enterprise-stack` 分支上，把当前 `feat` 分支的 Clean Architecture 演示项目升级为 **Gin + GORM + 轻量 DDD + REST/gRPC + 完整企业级基础设施** 的现代化 Go 项目骨架。

---

## 一、目标与非目标

### 1.1 目标

1. 保留原有业务领域（Article / Author CRUD）的所有功能
2. 引入企业级 Go 技术栈（Gin、GORM、Viper、Zap、JWT、Prometheus 等）
3. 用轻量 DDD 思想重新组织代码（聚合根、仓储接口、应用服务）
4. 同时提供 REST（Gin）和 gRPC 双协议接入
5. 包含完整的可观测性、健康检查、优雅停机、配置多环境管理
6. 文档齐全、目录清晰、可以拿来作为新项目的脚手架

### 1.2 非目标

- 不引入 CQRS / Event Sourcing（保持轻量 DDD）
- 不引入分布式追踪（OpenTelemetry 暂作预留接口，不强制启用）
- 不引入 Kubernetes 部署文件（保留 docker-compose）
- 不引入 API 网关、注册中心等微服务组件
- 暂不做分布式事务、消息队列

---

## 二、技术栈与版本

| 类别 | 选型 | 版本 | 替代方案说明 |
|---|---|---|---|
| 语言 | Go | 1.23+ | 1.22+ 均可，最低 1.21（slices/maps 包需要） |
| Web 框架 | `gin-gonic/gin` | v1.10+ | 也可 echo/fiber，选 gin 因社区最大 |
| ORM | `gorm.io/gorm` | v1.25+ | 也可 sqlx + 手动，选 GORM 因企业项目最常见 |
| MySQL 驱动 | `gorm.io/driver/mysql` | v1.5+ | GORM 官方驱动 |
| 配置 | `spf13/viper` | v1.18+ | 也可 env-only |
| 日志 | `uber-go/zap` | v1.27+ | 也可 zerolog/slog |
| JWT | `golang-jwt/jwt/v5` | v5+ | 也可 jwx |
| 密码哈希 | `golang.org/x/crypto/bcrypt` | latest | 标准库扩展 |
| 数据库迁移 | `golang-migrate/migrate/v4` | v4.17+ | Makefile 已预留 |
| 参数校验 | `go-playground/validator/v10` | v10+ | Gin 默认集成 |
| Mock | `vektra/mockery` | v2.42+ | 沿用 |
| 测试 | `stretchr/testify` | v1.9+ | 沿用 |
| Metrics | `prometheus/client_golang` | v1.19+ | 中间件暴露 |
| 链路追踪 | `otel` 预留 | — | 暂不启用，留好接口 |
| 热重载 | `cosmtrek/air` | v1.52+ | 沿用 |
| 静态检查 | `golangci-lint` | v1.60+ | 配置升级到 v2 schema |
| Protobuf | `bufbuild/buf` | v1.34+ | 现代 protoc 替代 |
| Protobuf 生成器 | `protoc-gen-go` / `protoc-gen-go-grpc` | latest | 配套 buf |

---

## 三、目录结构

```
go-clean-arch_demo/
├── cmd/                                # 入口
│   ├── rest/main.go                    #   REST 入口（Gin）
│   └── grpc/main.go                    #   gRPC 入口
│
├── api/                                # API 协议层
│   ├── proto/                          #   .proto 源文件
│   │   ├── article/v1/article.proto
│   │   ├── author/v1/author.proto
│   │   └── auth/v1/auth.proto
│   └── gen/                            #   buf 生成代码（gitignored 或 vendor）
│       ├── go/
│       └── grpc/
│
├── internal/
│   ├── domain/                         # 领域层（最内层，零外部依赖）
│   │   ├── article/
│   │   │   ├── entity.go               #   Article 实体 + 值对象
│   │   │   ├── repository.go           #   Repository 接口（依赖反转）
│   │   │   └── errors.go               #   领域错误
│   │   ├── author/
│   │   │   ├── entity.go
│   │   │   ├── repository.go
│   │   │   └── errors.go
│   │   └── user/                       #   新增：JWT 登录需要用户聚合
│   │       ├── entity.go
│   │       ├── repository.go
│   │       └── errors.go
│   │
│   ├── application/                    # 应用层（用例编排）
│   │   ├── article/
│   │   │   ├── service.go              #   ArticleService（依赖注入 repo）
│   │   │   └── dto.go                  #   输入输出 DTO（隔离传输层）
│   │   ├── author/
│   │   │   ├── service.go
│   │   │   └── dto.go
│   │   ├── auth/
│   │   │   ├── service.go              #   AuthService（登录、刷新）
│   │   │   └── dto.go
│   │   └── user/
│   │       ├── service.go
│   │       └── dto.go
│   │
│   ├── infrastructure/                 # 基础设施层
│   │   ├── config/
│   │   │   └── config.go               #   Viper 配置加载
│   │   └── persistence/
│   │       ├── gorm/
│   │       │   ├── db.go               #   GORM 连接工厂
│   │       │   ├── base_model.go       #   公共 model（id/created_at/updated_at/deleted_at）
│   │       │   ├── article_model.go    #   GORM model struct（与 domain entity 解耦）
│   │       │   ├── article_repository.go  #   实现 domain.Repository 接口
│   │       │   ├── author_model.go
│   │       │   ├── author_repository.go
│   │       │   ├── user_model.go
│   │       │   └── user_repository.go
│   │       └── migrations/             #   golang-migrate 迁移文件
│   │           ├── 0001_init.up.sql
│   │           ├── 0001_init.down.sql
│   │           └── ...
│   │
│   └── interfaces/                     # 接口适配层
│       ├── http/                       #   REST 适配器
│       │   ├── handler/
│       │   │   ├── article.go          #   ArticleHandler（Gin）
│       │   │   ├── author.go
│       │   │   └── auth.go             #   Login/Refresh
│       │   ├── middleware/
│       │   │   ├── cors.go             #   跨域
│       │   │   ├── logger.go           #   请求日志（基于 Zap）
│       │   │   ├── recovery.go         #   panic 恢复
│       │   │   ├── metrics.go          #   Prometheus
│       │   │   ├── jwt.go              #   ★ JWT 鉴权（新增）
│       │   │   └── tracing.go          #   占位（预留 OTel 接口）
│       │   └── router.go               #   路由注册
│       └── grpc/                       #   gRPC 适配器
│           ├── handler/
│           │   └── article.go          #   GetArticle 示例
│           └── server.go               #   gRPC server 注册
│
├── pkg/                                # 可被外部项目引用的公共工具
│   ├── logger/                         #   Zap 封装
│   ├── errcode/                        #   统一错误码
│   ├── jwt/                            #   JWT 签发/解析
│   ├── response/                       #   HTTP 统一响应
│   └── server/                         #   优雅启停 + 信号处理
│       ├── http.go
│       └── grpc.go
│
├── configs/                            # 配置（多环境 YAML）
│   ├── config.yaml                     #   默认配置
│   ├── config.local.yaml               #   本地覆盖（gitignore）
│   └── config.prod.yaml                #   生产模板
│
├── deployments/                        # 部署相关
│   ├── Dockerfile                      #   多阶段 + distroless
│   └── docker-compose.yaml             #   本地编排
│
├── docs/                               # 文档
│   ├── README.md                       #   文档目录
│   ├── architecture.md                 #   架构图 + 依赖方向
│   ├── api.md                          #   REST + gRPC 接口说明
│   ├── infrastructure.md               #   基础设施详解
│   └── v2-design.md                    #   本文档
│
├── scripts/                            # 辅助脚本
│   └── proto.sh                        #   buf generate 一键脚本
│
├── buf.gen.yaml                        # Buf 代码生成配置
├── buf.yaml                            # Buf workspace
├── Makefile                            # 命令入口（升级版）
├── .air.toml                           # 热重载
├── .golangci.yaml                      # 静态检查（v2 schema）
├── .dockerignore
├── .env.example
├── go.mod
└── README.md
```

---

## 四、领域层设计（DDD）

### 4.1 聚合根划分

采用**两个独立聚合**（决策 2A）：

```
Article 聚合根
├── ID (int64)
├── Title (string)
├── Content (string)
├── AuthorID (int64)         ← 仅引用 Author 聚合的 ID，不持有 Author 实体
├── AuthorName (string)      ← 冗余字段：写入时的快照
├── CreatedAt/UpdatedAt
└── 行为：Create/Update/Delete/Publis h（示例）
```

```
Author 聚合根
├── ID (int64)
├── Name (string)
├── Email (string, value object 候选)
├── CreatedAt/UpdatedAt
└── 行为：Rename/ChangeEmail
```

```
User 聚合根（新增）
├── ID (int64)
├── Username (string)
├── PasswordHash (string)    ← bcrypt
├── CreatedAt/UpdatedAt
└── 行为：SetPassword/VerifyPassword
```

**关键 DDD 原则**：
- `Article` 不持有 `Author` 实体，只持有 `AuthorID` + 冗余 `AuthorName`（这是性能优化，避免每次列表查询 N+1）
- `Author` 完全独立，可单独管理
- `User` 是认证专用聚合，与业务 Article/Author 解耦

### 4.2 仓储接口设计

接口**定义在使用方包内**（沿用 current 项目约定）：

```go
// internal/domain/article/repository.go
package article

import "context"

type Repository interface {
    Save(ctx context.Context, a *Article) error
    GetByID(ctx context.Context, id int64) (*Article, error)
    List(ctx context.Context, cursor string, limit int) ([]*Article, string, error)
    Update(ctx context.Context, a *Article) error
    Delete(ctx context.Context, id int64) error
}
```

GORM 实现在 `internal/infrastructure/persistence/gorm/` 下，实现这个接口。GORM model struct 与 domain entity 解耦（防止 ORM 字段污染领域）。

### 4.3 领域错误

```go
// internal/domain/article/errors.go
package article

import "errors"

var (
    ErrNotFound      = errors.New("article not found")
    ErrAlreadyExists = errors.New("article already exists")
    ErrInvalidInput  = errors.New("invalid article input")
)
```

`internal/domain/author/errors.go` 同理。`User` 增加 `ErrInvalidCredentials`、`ErrUserExists` 等。

---

## 五、应用层设计

### 5.1 Service 依赖注入

```go
// internal/application/article/service.go
package article

import (
    "context"

    "github.com/zhanbinb/go-clean-arch_demo/internal/domain/article"
)

type Service struct {
    articles  article.Repository
    authors   author.Repository       // 跨聚合查询
    logger    *logger.Logger
}

func NewService(articles article.Repository, authors author.Repository, log *logger.Logger) *Service {
    return &Service{articles: articles, authors: authors, logger: log}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (*ArticleDTO, error) {
    // 1. 校验作者存在
    // 2. 创建 Article 实体
    // 3. 持久化
    // 4. 返回 DTO
}
```

### 5.2 DTO 隔离

```go
// internal/application/article/dto.go
type CreateInput struct {
    Title    string `json:"title" validate:"required,min=1,max=200"`
    Content  string `json:"content" validate:"required,min=1"`
    AuthorID int64  `json:"author_id" validate:"required,gt=0"`
}

type ArticleDTO struct {
    ID         int64     `json:"id"`
    Title      string    `json:"title"`
    Content    string    `json:"content"`
    AuthorID   int64     `json:"author_id"`
    AuthorName string    `json:"author_name"`
    CreatedAt  time.Time `json:"created_at"`
    UpdatedAt  time.Time `json:"updated_at"`
}
```

**关键**：DTO 不在领域层定义，**在 application 层定义**。HTTP/gRPC handler 把它转成 JSON 或 protobuf。这样：
- domain entity 不被 JSON tag / validate tag 污染
- 不同接口层（HTTP/gRPC）共用同一组 DTO
- 协议升级时只改 DTO，不改 domain

---

## 六、基础设施层设计

### 6.1 GORM 连接

```go
// internal/infrastructure/persistence/gorm/db.go
package gorm

import (
    "fmt"
    "time"

    "gorm.io/driver/mysql"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"

    "github.com/zhanbinb/go-clean-arch_demo/internal/infrastructure/config"
)

func NewDB(cfg config.Database) (*gorm.DB, error) {
    dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=true&loc=Local",
        cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name)

    db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Warn),
    })
    if err != nil {
        return nil, fmt.Errorf("open mysql: %w", err)
    }

    sqlDB, _ := db.DB()
    sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
    sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
    sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)
    return db, nil
}
```

### 6.2 GORM Model 与 Domain Entity 解耦

```go
// internal/infrastructure/persistence/gorm/article_model.go
package gorm

import "time"

// ArticleModel GORM 用的数据库模型
// 注意：与 domain/article/entity.go 中的 Article 不同！
type ArticleModel struct {
    ID         uint64    `gorm:"primaryKey;column:id"`
    Title      string    `gorm:"size:200;not null;column:title"`
    Content    string    `gorm:"type:text;not null;column:content"`
    AuthorID   uint64    `gorm:"not null;index;column:author_id"`
    AuthorName string    `gorm:"size:100;column:author_name"`
    CreatedAt  time.Time `gorm:"column:created_at"`
    UpdatedAt  time.Time `gorm:"column:updated_at"`
}

func (ArticleModel) TableName() string { return "articles" }
```

```go
// internal/infrastructure/persistence/gorm/article_repository.go
type articleRepository struct {
    db *gorm.DB
}

func NewArticleRepository(db *gorm.DB) article.Repository {
    return &articleRepository{db: db}
}

func (r *articleRepository) GetByID(ctx context.Context, id int64) (*article.Article, error) {
    var m ArticleModel
    err := r.db.WithContext(ctx).First(&m, id).Error
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, article.ErrNotFound
        }
        return nil, fmt.Errorf("query article: %w", err)
    }
    return toArticleEntity(&m), nil  // 转换为 domain entity
}
```

**为什么 Model 与 Entity 解耦？**
- ORM 字段（gorm tag）不该污染领域实体
- 数据库字段调整不需要改 domain code
- 多个聚合共享同一张表的特殊情况容易处理

### 6.3 数据库迁移

```
internal/infrastructure/persistence/migrations/
├── 0001_init.up.sql      # 创建 articles, authors, users 表
├── 0001_init.down.sql
├── 0002_add_user.up.sql  # 如果后续需要扩展
└── 0002_add_user.down.sql
```

`Makefile` 启用之前注释掉的 migrate 命令：
```makefile
migrate-up:
	migrate -database $(MYSQL_DSN) -path internal/infrastructure/persistence/migrations up

migrate-down:
	migrate -database $(MYSQL_DSN) -path internal/infrastructure/persistence/migrations down 1
```

`article.sql` 被替代为迁移文件。

---

## 七、接口层设计

### 7.1 REST（Gin）

```go
// internal/interfaces/http/handler/article.go
package handler

import (
    "net/http"
    "strconv"

    "github.com/gin-gonic/gin"

    "github.com/zhanbinb/go-clean-arch_demo/internal/application/article"
    "github.com/zhanbinb/go-clean-arch_demo/pkg/errcode"
    "github.com/zhanbinb/go-clean-arch_demo/pkg/response"
)

type ArticleHandler struct {
    svc *article.Service
}

func NewArticleHandler(svc *article.Service) *ArticleHandler {
    return &ArticleHandler{svc: svc}
}

// POST /api/v1/articles
func (h *ArticleHandler) Create(c *gin.Context) {
    var input article.CreateInput
    if err := c.ShouldBindJSON(&input); err != nil {
        response.Error(c, errcode.ErrBadRequest.WithMsg(err.Error()))
        return
    }
    dto, err := h.svc.Create(c.Request.Context(), input)
    if err != nil {
        response.Error(c, errcode.FromError(err))
        return
    }
    response.OK(c, dto)
}
```

**统一响应格式**：
```go
// pkg/response/response.go
type Resp struct {
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
}
```

**错误码体系**：
```go
// pkg/errcode/errcode.go
var (
    ErrOK              = New(0, "OK")
    ErrBadRequest      = New(40000, "bad request")
    ErrUnauthorized    = New(40100, "unauthorized")
    ErrForbidden       = New(40300, "forbidden")
    ErrNotFound        = New(40400, "not found")
    ErrConflict        = New(40900, "conflict")
    ErrInternal        = New(50000, "internal server error")
)
```

### 7.2 路由分组（带版本 + JWT）

```go
// internal/interfaces/http/router.go
func NewRouter(cfg *config.Config, h *Handlers, mw *Middlewares) *gin.Engine {
    r := gin.New()
    r.Use(mw.Logger(), mw.Recovery(), mw.CORS(), mw.Metrics())

    // 健康检查（无需认证）
    r.GET("/healthz", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
    r.GET("/readyz", h.Health.Ready)

    // 公开 API
    v1 := r.Group("/api/v1")
    v1.POST("/auth/login", h.Auth.Login)
    v1.POST("/auth/refresh", h.Auth.Refresh)

    // 需要鉴权的 API
    auth := v1.Group("")
    auth.Use(mw.JWT())
    {
        auth.POST("/articles", h.Article.Create)
        auth.GET("/articles", h.Article.List)
        auth.GET("/articles/:id", h.Article.Get)
        auth.PUT("/articles/:id", h.Article.Update)
        auth.DELETE("/articles/:id", h.Article.Delete)

        auth.POST("/authors", h.Author.Create)
        auth.GET("/authors/:id", h.Author.Get)
    }

    return r
}
```

### 7.3 gRPC（示例）

```protobuf
// api/proto/article/v1/article.proto
syntax = "proto3";

package article.v1;

option go_package = "github.com/zhanbinb/go-clean-arch_demo/api/gen/go/article/v1;articlev1";

service ArticleService {
    rpc GetArticle(GetArticleRequest) returns (GetArticleResponse);
    // 预留其他接口
    rpc CreateArticle(CreateArticleRequest) returns (CreateArticleResponse);
    rpc ListArticles(ListArticlesRequest) returns (ListArticlesResponse);
}

message GetArticleRequest {
    int64 id = 1;
}

message GetArticleResponse {
    Article article = 1;
}

message Article {
    int64 id = 1;
    string title = 2;
    string content = 3;
    int64 author_id = 4;
    string author_name = 5;
    int64 created_at = 6;  // unix timestamp
    int64 updated_at = 7;
}
```

```go
// internal/interfaces/grpc/handler/article.go
type articleServer struct {
    articlev1.UnimplementedArticleServiceServer
    svc *article.Service
}

func (s *articleServer) GetArticle(ctx context.Context, req *articlev1.GetArticleRequest) (*articlev1.GetArticleResponse, error) {
    dto, err := s.svc.GetByID(ctx, req.GetId())
    if err != nil {
        return nil, status.Error(codes.NotFound, "article not found")
    }
    return &articlev1.GetArticleResponse{
        Article: toProtoArticle(dto),
    }, nil
}
```

**REST 和 gRPC 共享同一个 application.Service**：不需要重复实现业务逻辑。

---

## 八、横切关注点

### 8.1 配置（Viper）

```go
// internal/infrastructure/config/config.go
type Config struct {
    Server   ServerConfig   `mapstructure:"server"`
    Database DatabaseConfig `mapstructure:"database"`
    JWT      JWTConfig      `mapstructure:"jwt"`
    Log      LogConfig      `mapstructure:"log"`
}

func Load(env string) (*Config, error) {
    viper.SetConfigName("config")
    viper.AddConfigPath("./configs")
    if err := viper.ReadInConfig(); err != nil {
        return nil, err
    }
    if env != "" {
        viper.SetConfigName("config." + env)
        viper.MergeInConfig()
    }
    viper.SetEnvPrefix("APP")
    viper.AutomaticEnv()
    // ...
}
```

```yaml
# configs/config.yaml
server:
  http_port: 9090
  grpc_port: 9091
  mode: release  # debug/release/test
database:
  host: 127.0.0.1
  port: 3306
  user: app
  password: app
  name: article
  max_open_conns: 50
  max_idle_conns: 10
  conn_max_lifetime: 3600
jwt:
  secret: "change-me-in-production"
  ttl: 3600
  refresh_ttl: 86400
log:
  level: info
  format: json
```

### 8.2 日志（Zap）

```go
// pkg/logger/logger.go
type Logger struct {
    z *zap.Logger
}

func New(level, format string) (*Logger, error) {
    var cfg zap.Config
    if format == "json" {
        cfg = zap.NewProductionConfig()
    } else {
        cfg = zap.NewDevelopmentConfig()
    }
    if err := cfg.Level.UnmarshalText([]byte(level)); err != nil {
        return nil, err
    }
    z, err := cfg.Build(zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel))
    if err != nil {
        return nil, err
    }
    return &Logger{z: z}, nil
}

func (l *Logger) With(fields ...zap.Field) *Logger {
    return &Logger{z: l.z.With(fields...)}
}

func (l *Logger) Info(msg string, fields ...zap.Field)  { l.z.Info(msg, fields...) }
func (l *Logger) Warn(msg string, fields ...zap.Field)  { l.z.Warn(msg, fields...) }
func (l *Logger) Error(msg string, fields ...zap.Field) { l.z.Error(msg, fields...) }
```

### 8.3 JWT 鉴权

```go
// pkg/jwt/jwt.go
type Claims struct {
    UserID   int64  `json:"uid"`
    Username string `json:"usr"`
    jwt.RegisteredClaims
}

type Manager struct {
    secret []byte
    ttl    time.Duration
}

func (m *Manager) Sign(userID int64, username string) (string, error) {
    claims := Claims{
        UserID: userID, Username: username,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.ttl)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            Issuer:    "go-clean-arch-demo",
        },
    }
    return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
}
```

中间件：
```go
// internal/interfaces/http/middleware/jwt.go
func JWT(secret string) gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("Authorization")
        if !strings.HasPrefix(token, "Bearer ") {
            response.Error(c, errcode.ErrUnauthorized)
            c.Abort()
            return
        }
        claims := &jwt.Claims{}
        if _, err := jwt.ParseWithClaims(token[7:], claims, func(t *jwt.Token) (interface{}, error) {
            return []byte(secret), nil
        }); err != nil {
            response.Error(c, errcode.ErrUnauthorized)
            c.Abort()
            return
        }
        c.Set("user_id", claims.UserID)
        c.Set("username", claims.Username)
        c.Next()
    }
}
```

### 8.4 优雅停机 + 信号处理

```go
// pkg/server/http.go
func RunHTTPServer(ctx context.Context, addr string, h http.Handler, logger *logger.Logger) error {
    srv := &http.Server{Addr: addr, Handler: h}

    errCh := make(chan error, 1)
    go func() {
        logger.Info("http server starting", zap.String("addr", addr))
        if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
            errCh <- err
        }
    }()

    select {
    case err := <-errCh:
        return err
    case <-ctx.Done():
        logger.Info("http server shutting down")
        shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        defer cancel()
        return srv.Shutdown(shutdownCtx)
    }
}
```

```go
// cmd/rest/main.go (摘要)
func main() {
    cfg := config.Load()
    log, _ := logger.New(cfg.Log.Level, cfg.Log.Format)
    defer log.Sync()

    db, _ := gorm.NewDB(cfg.Database)

    ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
    defer cancel()

    // ... 装配 services, handlers ...
    router := interfaces.NewRouter(cfg, h, mw)

    if err := server.RunHTTPServer(ctx, addr, router, log); err != nil {
        log.Fatal("http server", zap.Error(err))
    }
}
```

### 8.5 健康检查 + Metrics

```go
// /healthz  进程存活
// /readyz   检查 DB 连接、依赖等

// Prometheus 中间件：暴露 /metrics
r.GET("/metrics", gin.WrapH(promhttp.Handler()))
```

---

## 九、Makefile 升级

```makefile
# 新增 targets
make proto              # buf generate 重新生成 protobuf 代码
make migrate-up         # 应用所有迁移
make migrate-down       # 回滚一个迁移
make migrate-create N   # 创建新迁移文件
make run-rest           # 启动 REST 服务
make run-grpc           # 启动 gRPC 服务

# 保留
make build / build-race
make tests / tests-complete
make lint
make install-deps       # 加入 buf、protoc-gen-go、golangci-lint v2
make dev-air
make image-build
make clean
```

---

## 十、Dockerfile 升级

```dockerfile
# Stage 1: builder
FROM golang:1.23-alpine3.20 AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 make build-rest && CGO_ENABLED=0 make build-grpc

# Stage 2: runtime (distroless，无 shell，更小更安全)
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /src/bin/rest /app/rest
COPY --from=builder /src/bin/grpc /app/grpc
COPY --from=builder /src/configs /app/configs

EXPOSE 9090 9091
USER nonroot:nonroot
ENTRYPOINT ["/app/rest"]
```

distroless 镜像大小 ~10MB，比 alpine 还小，且无 shell 无法被攻击。

---

## 十一、改造前后文件映射

| Old (current main) | New (v2) |
|---|---|
| `app/main.go` | `cmd/rest/main.go` + `cmd/grpc/main.go` |
| `domain/article.go` | `internal/domain/article/entity.go` |
| `domain/author.go` | `internal/domain/author/entity.go` |
| `domain/errors.go` | 拆到 `internal/domain/{article,author}/errors.go` + `pkg/errcode/` |
| `article/service.go` | `internal/application/article/service.go` |
| `article/mocks/` | `internal/application/article/mocks/` |
| `internal/repository/mysql/article.go` | `internal/infrastructure/persistence/gorm/article_repository.go`（+ `article_model.go`） |
| `internal/repository/mysql/author.go` | `internal/infrastructure/persistence/gorm/author_repository.go`（+ `author_model.go`） |
| `internal/rest/article.go` | `internal/interfaces/http/handler/article.go` |
| `internal/rest/middleware/` | `internal/interfaces/http/middleware/`（扩展 + jwt + logger + recovery + metrics） |
| `internal/rest/mocks/` | `internal/interfaces/http/mocks/` |
| `article.sql` | `internal/infrastructure/persistence/migrations/*.sql` |
| `example.env` | `configs/config.yaml` + `configs/config.local.yaml` |
| `Dockerfile` | `deployments/Dockerfile`（多阶段 + distroless） |
| `compose.yaml` | `deployments/docker-compose.yaml` |
| `.golangci.yaml` | v2 schema 升级 |
| `Makefile` | 加入 proto / migrate / run-rest / run-grpc targets |

**新增文件**：
- `cmd/rest/main.go`、`cmd/grpc/main.go`
- `internal/domain/user/{entity,repository,errors}.go`
- `internal/application/{auth,user}/{service,dto}.go`
- `internal/infrastructure/config/config.go`
- `internal/infrastructure/persistence/gorm/{db,base_model,user_model,user_repository}.go`
- `internal/infrastructure/persistence/migrations/0001_init.{up,down}.sql`
- `internal/interfaces/http/{router,handler/auth}.go`
- `internal/interfaces/http/middleware/{jwt,logger,recovery,metrics,tracing}.go`
- `internal/interfaces/grpc/{server,handler/article}.go`
- `pkg/{logger,errcode,jwt,response,server/{http,grpc}}.go`
- `configs/{config.yaml,config.local.yaml,config.prod.yaml}`
- `api/proto/{article,author,auth}/v1/*.proto`
- `buf.{yaml,gen.yaml}`
- `scripts/proto.sh`
- `docs/{README,architecture,api}.md`

---

## 十二、实施阶段（按 commit 拆分）

| 阶段 | 内容 | 提交粒度 |
|---|---|---|
| 1 | 创建骨架：`go.mod` 升级 + 目录创建 + 占位文件 | 1 commit |
| 2 | `pkg/` 公共工具：logger / errcode / response / jwt / server | 1-2 commits |
| 3 | 领域层：`internal/domain/{article,author,user}` | 1 commit |
| 4 | 应用层：`internal/application/{article,author,user,auth}` + DTO | 1-2 commits |
| 5 | 基础设施：`config/` + `persistence/gorm/` + 迁移文件 | 2 commits |
| 6 | 接口层 REST：`interfaces/http/` 全套 | 2 commits |
| 7 | 接口层 gRPC：proto 定义 + 生成代码 + handler | 1 commit |
| 8 | 入口 `cmd/{rest,grpc}/main.go` | 1 commit |
| 9 | 基础设施：Makefile + Dockerfile + docker-compose + .golangci.yaml + .air.toml | 1-2 commits |
| 10 | 文档：`docs/{architecture,api}.md` + 更新 README.md | 1 commit |
| 11 | 跑通所有：构建 + 测试 + lint + docker build | 1 commit（fix） |

预计 15-20 个 commits。

---

## 十三、风险与回退方案

### 13.1 风险

1. **buf / protoc 工具链安装**：新机器需要装 buf、protoc-gen-go、protoc-gen-go-grpc。需要保证 `make install-deps` 能装好（或文档说明）
2. **distroless 镜像调试困难**：无 shell。需要本地调试时用 alpine 版本
3. **GORM v2 性能**：相比 sqlx，GORM 有反射开销。高并发场景需要 benchmark 验证
4. **JWT secret 管理**：配置里的 secret 不能入库。需要确认 `.gitignore` 排除 `config.local.yaml`

### 13.2 回退

`main` 分支保持原样不变。新分支 `feat/v2-enterprise-stack` 完全独立。如果改造失败，`git checkout main` 即可回到当前状态。

---

## 十四、待确认事项（提交 review 后可能的迭代）

1. **API 路径前缀**：当前方案是 `/api/v1/...`，是否需要更细的版本号（如 `/api/v1/articles/...` vs `/api/articles/...`）？
2. **Swagger 文档**：是否需要集成 swag 自动生成 Swagger UI？
3. **OpenTelemetry**：是否在本次就启用（需要选 exporter：Jaeger / Tempo / OTLP）？
4. **部署脚本**：是否需要 k8s manifests / Helm chart / 简单的 systemd unit？

---

## 十五、确认与下一步

请 review 本文档。如果有要调整的地方，告诉我。

**通过后**，我会按"十二、实施阶段"的顺序逐步提交，每个阶段后给你 diff review。中间任何一步你觉得不对，可以 `git revert` 单个 commit，不影响其他进度。

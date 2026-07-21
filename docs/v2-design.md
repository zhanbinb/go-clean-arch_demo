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

## 十四、确认决策（review 后定稿）

| 决策点 | 决定 |
|---|---|
| API 路径前缀 | `/api/v1/...`（URL 路径版本化） |
| Swagger 文档 | **纳入**：使用 `swaggo/swag` 生成 OpenAPI + Swagger UI |
| OpenTelemetry | **暂不启用**：在代码里预留接口，下版本再做 |
| K8s 部署 | **TODO**：v3 再做 |
| **新增组件**（强烈推荐 9 项） | **全部纳入**：Swagger、gRPC reflection、Request ID、pprof、限流、熔断、CI、CORS、优雅停机 |
| **可选组件**（C 类） | **TODO**：v3 再做（详见第十七章） |
| API Gateway | **TODO**：v3 再做（部署层 Nginx/Kong 文档 + 进程内 BFF 二选一） |

---

## 十五、补充组件设计（新增 9 项）

> 以下 9 项是企业级 Go 后端服务的"标配"，按统一标准在 v2 中实现。

### 15.1 Swagger / OpenAPI 文档

**栈选型**：

| 库 | 用途 |
|---|---|
| `swaggo/swag` | 从代码注释生成 OpenAPI 2.0/3.0 规范 |
| `swaggo/files` | 嵌入 Swagger UI 静态资源 |
| `swaggo/gin-swagger` | Gin 适配器 |

**注解约定**（每个 handler 方法上方）：

```go
// CreateArticle godoc
// @Summary      创建文章
// @Description  需要登录；标题 1-200 字；内容必填
// @Tags         articles
// @Accept       json
// @Produce      json
// @Param        Authorization  header    string                    true  "Bearer {token}"
// @Param        input          body      article.CreateInput        true  "文章内容"
// @Success      201            {object}  response.Resp{data=article.ArticleDTO}
// @Failure      400            {object}  response.Resp
// @Failure      401            {object}  response.Resp
// @Router       /api/v1/articles [post]
// @Security     BearerAuth
func (h *ArticleHandler) Create(c *gin.Context) { ... }
```

**生成与挂载**：

```makefile
swagger: $(SWAG)
    swag init -g cmd/rest/main.go -o docs/swagger --parseDependency --parseInternal
```

```go
// router.go
import (
    swaggerFiles "github.com/swaggo/files"
    ginSwagger "github.com/swaggo/gin-swagger"
    _ "github.com/zhanbinb/go-clean-arch_demo/docs/swagger" // 嵌入生成物
)

r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
```

访问：`http://localhost:9090/swagger/index.html`

**生成产物目录**：`docs/swagger/{swagger.json, swagger.yaml, docs.go}` — `docs.go` 需要嵌入二进制，所以**不忽略**该目录。

---

### 15.2 gRPC 调试工具链

**启用 Server Reflection**（让 grpcurl/grpcui 能动态发现服务）：

```go
// internal/interfaces/grpc/server.go
import "google.golang.org/grpc/reflection"

func NewServer(cfg *config.Config, h *Handlers) *grpc.Server {
    s := grpc.NewServer(
        grpc.UnaryInterceptor(middleware.ChainUnaryServer(...)),
    )
    articlev1.RegisterArticleServiceServer(s, h.Article)
    authv1.RegisterAuthServiceServer(s, h.Auth)
    reflection.Register(s)  // ← 关键
    return s
}
```

**调试命令**（开发文档里给示例）：

```bash
# 列出所有服务
grpcurl -plaintext localhost:9091 list

# 列出某个服务的方法
grpcurl -plaintext localhost:9091 list article.v1.ArticleService

# 调用 GetArticle
grpcurl -plaintext -d '{"id": 1}' localhost:9091 article.v1.ArticleService/GetArticle

# 启动 GUI 调试器（类似 Postman 的 gRPC 版）
grpcui -plaintext localhost:9091
```

**Makefile 集成**：

```makefile
install-deps: ... grpc-ui-tools  # 自动安装 grpcurl 和 grpcui

grpc-ui:
    grpcui -plaintext localhost:$(GRPC_PORT)
```

---

### 15.3 Request ID / Trace ID 中间件

**目的**：让一个 HTTP 请求的所有日志可关联（同一 request_id）。

**规则**：

1. 入口：取请求头 `X-Request-ID`
   - 如果有 → 使用
   - 如果没有 → 生成 UUID v4
2. 注入：
   - `gin.Context`（key = `"request_id"`）
   - `context.Context`（value，方便传到 service / repo）
   - 响应头 `X-Request-ID`（echo 给客户端）
3. Zap 日志字段自动带上 `request_id`

**实现要点**：

```go
// internal/interfaces/http/middleware/requestid.go
package middleware

import (
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "github.com/zhanbinb/go-clean-arch_demo/pkg/logger"
    "go.uber.org/zap"
)

const HeaderXRequestID = "X-Request-ID"
const ContextKeyRequestID = "request_id"

func RequestID(base *logger.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        rid := c.GetHeader(HeaderXRequestID)
        if rid == "" {
            rid = uuid.NewString()
        }
        c.Set(ContextKeyRequestID, rid)
        c.Writer.Header().Set(HeaderXRequestID, rid)

        // 把 request_id 注入 ctx 并把 logger 包成带字段的版本
        ctx := logger.WithRequestID(c.Request.Context(), rid)
        c.Request = c.Request.WithContext(ctx)
        c.Next()
    }
}

// 调用方式（service 层）：
func (s *Service) Create(ctx context.Context, input CreateInput) (*ArticleDTO, error) {
    rid := logger.RequestIDFromContext(ctx)  // 取出 request_id
    s.logger.Info("creating article", zap.String("request_id", rid), zap.String("title", input.Title))
    ...
}
```

`pkg/logger` 里加 context 传递工具：

```go
type ctxKey int
const ctxKeyRequestID ctxKey = iota

func WithRequestID(ctx context.Context, rid string) context.Context {
    return context.WithValue(ctx, ctxKeyRequestID, rid)
}
func RequestIDFromContext(ctx context.Context) string {
    v, _ := ctx.Value(ctxKeyRequestID).(string)
    return v
}
```

---

### 15.4 pprof 性能分析端点

**栈选型**：

- `net/http/pprof`（标准库）
- `github.com/gin-contrib/pprof`（Gin 适配器，避免手写路由）

**挂载位置**：`/debug/pprof/*`（默认端口即可，但生产环境应该做访问控制）

```go
import "github.com/gin-contrib/pprof"

// 仅在 debug 模式启用
if cfg.Server.Mode == "debug" {
    pprof.Register(r)
}
```

**生产环境访问控制建议**：

| 方案 | 说明 |
|---|---|
| 独立端口 | `pprof` 监听 `:6060`，不暴露到外网，由 ops 通过 SSH 隧道访问 |
| IP 白名单 | 中间件检查来源 IP（只允许内网） |
| Basic Auth | 中间件加 Basic Auth |
| 不挂载 | 生产完全不挂载，临时调试时改配置重启 |

v2 默认采用 **"debug 模式才挂载 + 独立端口 `:6060`"** 双保险。

```go
go func() {
    if err := http.ListenAndServe("localhost:6060", nil); err != nil {
        log.Error("pprof server", zap.Error(err))
    }
}()
```

**常用调试命令**：

```bash
# CPU profile 30 秒
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# 内存 heap
go tool pprof http://localhost:6060/debug/pprof/heap

# goroutine 死锁排查
curl http://localhost:6060/debug/pprof/goroutine?debug=2
```

---

### 15.5 限流（Rate Limiting）

**算法**：Token Bucket（允许突发，更友好）

**库**：`golang.org/x/time/rate`

**维度**：默认按 IP，可选按 UserID（JWT 中间件之后）

**存储**：内存 `map[string]*rate.Limiter`，定期清理

**配置**：

```yaml
ratelimit:
  enabled: true
  rps: 100            # 每秒补充令牌数
  burst: 200          # 桶容量（允许突发大小）
  cleanup_interval: 5m
  dimension: ip       # ip | user
```

**实现要点**：

```go
// pkg/ratelimit/ratelimit.go
type Limiter struct {
    mu       sync.Mutex
    visitors map[string]*visitor
    rps      rate.Limit
    burst    int
    ttl      time.Duration
}

type visitor struct {
    limiter  *rate.Limiter
    lastSeen time.Time
}

func New(cfg config.RateLimit) *Limiter {
    l := &Limiter{
        visitors: make(map[string]*visitor),
        rps:      rate.Limit(cfg.RPS),
        burst:    cfg.Burst,
        ttl:      cfg.CleanupInterval,
    }
    go l.cleanupLoop()
    return l
}

func (l *Limiter) Allow(key string) bool {
    l.mu.Lock()
    defer l.mu.Unlock()
    v, ok := l.visitors[key]
    if !ok {
        v = &visitor{limiter: rate.NewLimiter(l.rps, l.burst)}
        l.visitors[key] = v
    }
    v.lastSeen = time.Now()
    return v.limiter.Allow()
}
```

```go
// middleware/ratelimit.go
func RateLimit(l *ratelimit.Limiter) gin.HandlerFunc {
    return func(c *gin.Context) {
        key := c.ClientIP()
        if uid, ok := c.Get("user_id"); ok {
            key = fmt.Sprintf("user:%d", uid)
        }
        if !l.Allow(key) {
            c.Header("Retry-After", "1")
            response.Error(c, errcode.ErrTooManyRequests)
            c.Abort()
            return
        }
        c.Next()
    }
}
```

`errcode.ErrTooManyRequests = New(42900, "too many requests")` 加入错误码体系。

---

### 15.6 熔断（Circuit Breaker）

**库**：`github.com/sony/gobreaker`

**场景**：保护下游调用，主要是：
1. **数据库查询**（DB 短暂不可用时快速失败，避免请求堆积）
2. **未来可能引入的外部 API 调用**（HTTP / RPC）

**配置**：

```go
// pkg/circuitbreaker/breaker.go
type Breaker struct {
    cb *gobreaker.CircuitBreaker
}

func New(name string, settings gobreaker.Settings) *Breaker {
    settings.Name = name
    if settings.ReadyToTrip == nil {
        settings.ReadyToTrip = func(counts gobreaker.Counts) bool {
            return counts.ConsecutiveFailures > 5
        }
    }
    if settings.Timeout == 0 {
        settings.Timeout = 30 * time.Second
    }
    return &Breaker{cb: gobreaker.NewCircuitBreaker(settings)}
}

func (b *Breaker) Do(fn func() (interface{}, error)) (interface{}, error) {
    return b.cb.Execute(fn)
}
```

**在仓储层使用**（装饰器模式）：

```go
// internal/infrastructure/persistence/gorm/article_repository.go
type articleRepository struct {
    db      *gorm.DB
    breaker *circuitbreaker.Breaker
}

func (r *articleRepository) GetByID(ctx context.Context, id int64) (*article.Article, error) {
    result, err := r.breaker.Do(func() (interface{}, error) {
        var m ArticleModel
        if err := r.db.WithContext(ctx).First(&m, id).Error; err != nil {
            if errors.Is(err, gorm.ErrRecordNotFound) {
                return nil, article.ErrNotFound  // 业务错误不计入熔断
            }
            return nil, err
        }
        return toArticleEntity(&m), nil
    })
    if err != nil {
        return nil, err
    }
    return result.(*article.Article), nil
}
```

**注意**：业务错误（如 ErrNotFound）不应该触发熔断——只有"基础设施错误"（连接失败、超时）才算。所以 `Do` 内部需要区分。

---

### 15.7 CI 工作流

**栈选型**：GitHub Actions

**目录**：`.github/workflows/ci.yml`

**触发**：push 到 main / feat 分支 + 所有 PR

**Jobs**：

```yaml
name: CI

on:
  push:
    branches: [main, 'feat/**']
  pull_request:
    branches: [main]

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.23'
          cache: true
      - run: make install-deps
      - run: make lint
      - run: make swagger

  test:
    runs-on: ubuntu-latest
    services:
      mysql:
        image: mysql:8.3
        env:
          MYSQL_ROOT_PASSWORD: test
          MYSQL_DATABASE: test
          MYSQL_USER: test
          MYSQL_PASSWORD: test
        ports: ['3306:3306']
        options: >-
          --health-cmd="mysqladmin ping"
          --health-interval=5s
          --health-retries=10
    env:
      DATABASE_HOST: 127.0.0.1
      DATABASE_PORT: 3306
      DATABASE_USER: test
      DATABASE_PASSWORD: test
      DATABASE_NAME: test
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.23'
          cache: true
      - run: make install-deps
      - run: make migrate-up
      - run: make tests
      - name: Upload coverage
        if: always()
        uses: codecov/codecov-action@v4
        with:
          files: ./coverage.out

  build:
    runs-on: ubuntu-latest
    needs: [lint, test]
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.23'
          cache: true
      - run: make install-deps
      - run: make build-rest build-grpc
      - name: Upload artifacts
        uses: actions/upload-artifact@v4
        with:
          name: binaries
          path: bin/

  docker:
    runs-on: ubuntu-latest
    needs: [build]
    if: github.ref == 'refs/heads/main'
    steps:
      - uses: actions/checkout@v4
      - run: docker build -f deployments/Dockerfile -t ${{ github.repository }}:latest .
      - run: docker save ${{ github.repository }}:latest | gzip > image.tar.gz
      - uses: actions/upload-artifact@v4
        with:
          name: docker-image
          path: image.tar.gz
```

**缓存策略**：

- Go modules 缓存：`actions/setup-go@v5` 自带
- Docker layer 缓存：用 `docker/build-push-action@v5` 的 `cache-from: type=gha`

---

### 15.8 CORS 跨域

**库**：`github.com/gin-contrib/cors`

**配置（从 YAML 读取，避免硬编码）**：

```yaml
cors:
  allow_origins:
    - "*"            # 开发环境
    # 生产环境替换为具体域名：
    # - "https://app.example.com"
  allow_methods:
    - GET
    - POST
    - PUT
    - PATCH
    - DELETE
    - OPTIONS
  allow_headers:
    - Origin
    - Content-Type
    - Authorization
    - X-Request-ID
  expose_headers:
    - X-Request-ID
  allow_credentials: true
  max_age: 12h
```

**挂载**：

```go
// middleware/cors.go
import "github.com/gin-contrib/cors"

func CORS(cfg config.CORS) gin.HandlerFunc {
    return cors.New(cors.Config{
        AllowOrigins:     cfg.AllowOrigins,
        AllowMethods:     cfg.AllowMethods,
        AllowHeaders:     cfg.AllowHeaders,
        ExposeHeaders:    cfg.ExposeHeaders,
        AllowCredentials: cfg.AllowCredentials,
        MaxAge:           cfg.MaxAge,
    })
}
```

**生产环境注意**：不要用 `"*"` + `AllowCredentials: true`（浏览器会拒绝）。两者必须二选一。

---

### 15.9 优雅停机（详细设计）

**目标**：收到 `SIGINT` / `SIGTERM` 时：
1. 停止接收新请求
2. 等待正在处理的请求完成（最多 N 秒）
3. 关闭下游连接（DB、Redis、外部服务）
4. 进程退出

**信号处理**：

```go
// cmd/rest/main.go
ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
defer stop()
```

**HTTP 服务优雅停机**：

```go
// pkg/server/http.go
type HTTPServer struct {
    srv  *http.Server
    log  *logger.Logger
}

func NewHTTPServer(addr string, handler http.Handler, log *logger.Logger) *HTTPServer {
    return &HTTPServer{
        srv: &http.Server{
            Addr:              addr,
            Handler:           handler,
            ReadHeaderTimeout: 10 * time.Second,
            ReadTimeout:       30 * time.Second,
            WriteTimeout:      30 * time.Second,
            IdleTimeout:       120 * time.Second,
        },
        log: log,
    }
}

func (h *HTTPServer) Run(ctx context.Context) error {
    errCh := make(chan error, 1)
    go func() {
        h.log.Info("http server starting", zap.String("addr", h.srv.Addr))
        if err := h.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
            errCh <- err
        }
    }()

    select {
    case err := <-errCh:
        return err
    case <-ctx.Done():
        h.log.Info("http server shutting down")
        shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        defer cancel()
        if err := h.srv.Shutdown(shutdownCtx); err != nil {
            return fmt.Errorf("http shutdown: %w", err)
        }
        return nil
    }
}
```

**gRPC 服务优雅停机**：

```go
// pkg/server/grpc.go
func (g *GRPCServer) Run(ctx context.Context) error {
    errCh := make(chan error, 1)
    go func() {
        g.log.Info("grpc server starting", zap.String("addr", g.srv.Addr))
        errCh <- g.s.Serve(g.lis)
    }()

    select {
    case err := <-errCh:
        if errors.Is(err, grpc.ErrServerStopped) {
            return nil
        }
        return err
    case <-ctx.Done():
        g.log.Info("grpc server graceful stopping")
        stopped := make(chan struct{})
        go func() {
            g.s.GracefulStop()  // 等待进行中的 RPC 完成
            close(stopped)
        }()
        select {
        case <-stopped:
            return nil
        case <-time.After(30 * time.Second):
            g.log.Warn("grpc graceful stop timeout, forcing stop")
            g.s.Stop()
            return nil
        }
    }
}
```

**资源清理顺序**：

```go
func main() {
    cfg, log, db := init()  // 配置、日志、DB
    defer func() {
        sqlDB, _ := db.DB()
        sqlDB.Close()              // 1. 关闭 DB 连接池
        log.Sync()                 // 2. flush 日志缓冲
    }()

    ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
    defer stop()

    httpSrv := server.NewHTTPServer(cfg.Server.HTTPAddr, router, log)
    grpcSrv := server.NewGRPCServer(cfg.Server.GRPCAddr, grpcHandler, log)

    errCh := make(chan error, 2)
    go func() { errCh <- httpSrv.Run(ctx) }()
    go func() { errCh <- grpcSrv.Run(ctx) }()

    // 等待任意一个退出
    if err := <-errCh; err != nil {
        log.Error("server exited with error", zap.Error(err))
    }
    log.Info("all servers stopped, exiting")
}
```

---

## 十六、v3 版本规划（TODO）

> 以下功能本期（v2）**不做**，列入 TODO，在 v3 或后续版本评估引入。

### 16.1 API Gateway

| 维度 | 方案 |
|---|---|
| 部署层网关 | Nginx 反向代理 / Kong / APISIX / Envoy。文档给出 Nginx + 限流 + 鉴权转发的配置示例 |
| 进程内 BFF | `internal/gateway/` 聚合层，给前端做接口聚合和裁剪 |
| 服务网格 | Istio / Linkerd（k8s 场景） |

**评估时机**：v3 或微服务拆分时再选型

### 16.2 Redis 缓存

| 用途 | 价值 |
|---|---|
| 热点文章列表缓存 | 减少 DB 读压力 |
| 限流计数器（分布式） | 多实例部署时共享计数 |
| Session 存储 | 替换 JWT（如果改用 Session 认证） |
| 分布式锁 | 防止并发问题 |

**栈选型**：`redis/go-redis/v9`

### 16.3 审计日志

**场景**：合规要求（金融、医疗、政府）

**实现方向**：

- 数据库表：`audit_logs(id, actor_id, action, resource, before, after, created_at)`
- 中间件自动记录：谁在什么时候改了什么
- 不可篡改（可写专用存储如 WORM）

### 16.4 软删除（Soft Delete）

**方案**：GORM 内置 `gorm.DeletedAt`

```go
type ArticleModel struct {
    ID        uint64         `gorm:"primaryKey"`
    // ...
    DeletedAt gorm.DeletedAt `gorm:"index"`
}
```

**改动**：
- DELETE API 改为 UPDATE（设置 deleted_at）
- 列表查询默认过滤已删除
- 提供"包含已删除"参数供后台使用

### 16.5 集成测试（testcontainers）

**栈**：`testcontainers/testcontainers-go` + `mysql`

```go
func TestArticleRepository(t *testing.T) {
    ctx := context.Background()
    mysqlC, _ := mysql.RunContainer(ctx,
        testcontainers.WithImage("mysql:8.3"),
    )
    defer mysqlC.Terminate(ctx)
    
    db := connectTo(mysqlC.ConnectionString(ctx))
    repo := NewArticleRepository(db)
    
    // 真实 SQL 测试，不依赖 sqlmock
}
```

### 16.6 依赖自动更新

**方案 A：Dependabot**

`.github/dependabot.yml`：

```yaml
version: 2
updates:
  - package-ecosystem: "gomod"
    directory: "/"
    schedule:
      interval: "weekly"
  - package-ecosystem: "github-actions"
    directory: "/"
    schedule:
      interval: "weekly"
```

**方案 B：Renovate**（更灵活）

`renovate.json`：

```json
{
  "$schema": "https://docs.renovatebot.com/renovate-schema.json",
  "extends": ["config:recommended"],
  "schedule": ["before 6am on monday"]
}
```

### 16.7 代码覆盖率

**栈**：`codecov/codecov-action`

- CI 上传 coverage.out
- 设置覆盖率门槛（如 70%），不达标 PR 不允许合并

### 16.8 OAuth2 / OIDC

**库**：`ory/fosite`、`coreos/go-oidc`

**场景**：第三方登录（GitHub、Google、企业 SSO）

**改动**：
- 引入 OAuth2 Authorization Code 流程
- 多 Provider 支持
- 替换简单 JWT

### 16.9 OpenTelemetry 全栈

**栈选型**：

| 后端 | 适用 |
|---|---|
| Jaeger | 自建、传统 |
| Tempo | Grafana 生态 |
| 阿里云 ARMS | 阿里云 |
| 腾讯云 CAT | 腾讯云 |
| Datadog | SaaS |

**实施**：

```go
import "go.opentelemetry.io/otel"

func main() {
    otel.Init(...)  // 配置 OTLP exporter
    // HTTP / gRPC 中间件自动注入 trace
}
```

### 16.10 消息队列

**候选**：

| MQ | 适用 |
|---|---|
| Kafka | 大流量、事件溯源 |
| RabbitMQ | 任务队列、传统 |
| NATS | 轻量、实时 |
| Redis Streams | 简单队列 |

**典型场景**：领域事件发布（`ArticleCreated` → 触发通知 / 搜索索引更新）

### 16.11 多租户（Multi-tenancy）

**实现路径**：

```go
type TenantKey string
ctx := context.WithValue(parentCtx, TenantKey("tenant_id"), "tenant_abc")

// 每个查询自动加 WHERE tenant_id = ?
func (r *ArticleRepository) GetByID(ctx context.Context, id int64) (*Article, error) {
    tenant := ctx.Value(TenantKey("tenant_id")).(string)
    r.db.Where("tenant_id = ?", tenant).First(...)
}
```

### 16.12 Kubernetes / Helm

**交付物**：
- `deployments/k8s/deployment.yaml`
- `deployments/k8s/service.yaml`
- `deployments/k8s/ingress.yaml`
- `deployments/k8s/configmap.yaml`
- `deployments/helm/go-clean-arch/`

### 16.13 GraphQL

**栈**：`99designs/gqlgen`

**场景**：前端需要灵活查询字段时

### 16.14 特性开关（Feature Flag）

**栈**：`Unleash` / `LaunchDarkly` / 自研

**场景**：灰度发布、A/B 测试、紧急回滚

### 16.15 定时任务（Scheduler）

**栈**：`robfig/cron/v3`

**场景**：定期清理日志、定期汇总数据、定时备份

### 16.16 文件上传 / 对象存储

**栈**：`minio-go` / `aws-sdk-go-v2` / 阿里云 OSS

**场景**：用户头像、文章封面

### 16.17 WebSocket / SSE

**场景**：实时通知、进度推送

### 16.18 国际化（i18n）

**栈**：`nicksnyder/go-i18n/v2`

**场景**：多语言 API 错误消息

---

## 十七、确认与下一步

### 17.1 设计状态

✅ 设计文档已包含：
- 第三章：目录结构
- 第四章：领域层（DDD 聚合）
- 第五章：应用层
- 第六章：基础设施层（GORM）
- 第七章：接口层（REST + gRPC）
- 第八章：横切关注点
- 第十五章：补充 9 个组件（Swagger / gRPC reflection / Request ID / pprof / 限流 / 熔断 / CI / CORS / 优雅停机）
- 第十六章：v3 TODO 列表（17 项）

### 17.2 待用户最终确认

请确认以下三项：

1. **十五章 9 个组件的设计细节**是否符合预期？
   - 任何要调整的地方？
   - 任何要补充的地方？
2. **第十六章 TODO 列表**有没有要追加/移除的项？
3. **整体设计是否可以开始实施**？

### 17.3 实施计划（不变）

按"十二、实施阶段"的 15 阶段顺序，每个阶段 1-2 个 commit：

| 阶段 | 内容 | 状态 |
|---|---|---|
| 1 | 项目骨架：go.mod 升级 + 目录 + 占位文件 | 待开始 |
| 2 | `pkg/` 公共工具 | 待开始 |
| 3 | 领域层 | 待开始 |
| 4 | 应用层 + DTO | 待开始 |
| 5 | 基础设施：config + gorm + 迁移 | 待开始 |
| 6 | 接口层 REST + Swagger + 中间件 | 待开始 |
| 7 | 接口层 gRPC + reflection | 待开始 |
| 8 | `cmd/{rest,grpc}/main.go` | 待开始 |
| 9 | 基础设施：Makefile + Dockerfile + compose + golangci | 待开始 |
| 10 | `.github/workflows/ci.yml` | 待开始 |
| 11 | 文档：architecture.md + api.md + 更新 README | 待开始 |
| 12 | 跑通：build + test + lint + docker build | 待开始 |

预计 15-20 个 commits。

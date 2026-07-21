# 架构总览

本文档说明 v2 企业级架构的分层、依赖方向和关键设计决策。完整设计见 [v2-design.md](./v2-design.md)。

## 一、分层架构

```
┌─────────────────────────────────────────────────────────────┐
│                       cmd/                                  │
│              (REST + gRPC entry points)                     │
│         wire.New() → 装配所有依赖                            │
└────────────────────────┬────────────────────────────────────┘
                         │
┌────────────────────────▼────────────────────────────────────┐
│              internal/interfaces/                           │
│   ┌────────────────────────┬────────────────────────┐      │
│   │  http/  (Gin)          │  grpc/  (gRPC)         │      │
│   │  - handlers            │  - handlers            │      │
│   │  - middleware          │  - reflection          │      │
│   │  - router              │  - service desc        │      │
│   └────────────────────────┴────────────────────────┘      │
└────────────────────────┬────────────────────────────────────┘
                         │ 调用
┌────────────────────────▼────────────────────────────────────┐
│              internal/application/                          │
│   ┌──────────────┬──────────────┬──────────────┐            │
│   │ article/     │ author/      │ auth/        │ user/      │
│   │ Service      │ Service      │ Service      │ Service    │
│   │ + DTOs       │ + DTOs       │ + DTOs       │ + DTOs     │
│   └──────────────┴──────────────┴──────────────┘            │
└────────────────────────┬────────────────────────────────────┘
                         │ 依赖接口（依赖反转）
┌────────────────────────▼────────────────────────────────────┐
│              internal/domain/                                │
│   ┌──────────────┬──────────────┬──────────────┐            │
│   │ article/     │ author/      │ user/        │            │
│   │ Entity       │ Entity       │ Entity       │            │
│   │ Repository   │ Repository   │ Repository   │            │
│   │ (interface)  │ (interface)  │ (interface)  │            │
│   │ Errors       │ Errors       │ Errors       │            │
│   └──────────────┴──────────────┴──────────────┘            │
└────────────────────────▲────────────────────────────────────┘
                         │ 实现接口
┌────────────────────────┴────────────────────────────────────┐
│              internal/infrastructure/                       │
│   ┌─────────────────────────────────────────────────┐      │
│   │ persistence/gorm/                               │      │
│   │   - ArticleModel + ArticleRepository            │      │
│   │   - AuthorModel + AuthorRepository              │      │
│   │   - UserModel + UserRepository                  │      │
│   │   - db.go (NewDB)                               │      │
│   │ persistence/migrations/  (golang-migrate)       │      │
│   │ config/  (Viper)                                │      │
│   └─────────────────────────────────────────────────┘      │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│              pkg/  (公共工具，无业务依赖)                    │
│   logger · errcode · response · jwt · server                │
│   ratelimit · circuitbreaker                                │
└─────────────────────────────────────────────────────────────┘
```

## 二、依赖方向规则

**外圈依赖内圈，内圈不知道外圈存在**：

```
cmd/  →  internal/interfaces/  →  internal/application/  →  internal/domain/
                                ↑                                       ↑
                                └── internal/infrastructure/  ─────────┘
                                    (实现 domain 的 Repository 接口)
```

具体规则：

1. `domain/` 不 import 任何本项目包（最内层）
2. `application/` 只依赖 `domain/` + `pkg/`
3. `infrastructure/` 实现 `domain/` 的接口（依赖反转）
4. `interfaces/` 只依赖 `application/` + `pkg/`
5. `cmd/` 是组合根，知道所有具体类型
6. `pkg/` 不依赖任何本项目业务包（纯工具）

**关键**：`application/` 包里的 Repository 引用是**类型别名**：

```go
type ArticleRepo = article.Repository
```

这让 application 层不需要 import infrastructure 层。

## 三、关键设计决策

### 3.1 Article 聚合不持有 Author 实体

Article 聚合根只存 `AuthorID` + `AuthorName` 快照：

```go
type Article struct {
    ID         int64
    AuthorID   int64     // 引用，不是对象
    AuthorName string    // 写入时的快照
    // ...
}
```

**为什么**：避免 N+1 查询；列表查询不需要 join 也能展示作者名。

### 3.2 GORM Model 与 Domain Entity 解耦

```go
// 仓储层（基础设施）
type ArticleModel struct {  // GORM 用
    ID    uint64 `gorm:"primaryKey"`
    Title string `gorm:"size:200"`
    // ...
}

func (m *ArticleModel) toEntity() *article.Article { ... }  // 转 domain

// 领域层
type Article struct {  // 业务用
    ID    int64
    Title string
    // 不带任何 tag
}
```

**为什么**：ORM tag 污染业务实体；改表结构不需要改 domain。

### 3.3 DTO 在 application 层定义

```go
// application/article/dto.go
type ArticleDTO struct {
    ID      int64     `json:"id"`
    Title   string    `json:"title"`
    // JSON tag 在 DTO 上，不在 entity 上
}
```

**为什么**：domain entity 不被 JSON / validate tag 污染；改 API 协议不影响 domain。

### 3.4 REST + gRPC 共享 application.Service

```go
// HTTP handler
func (h *ArticleHandler) GetByID(c *gin.Context) {
    dto, err := h.svc.GetByID(c.Request.Context(), id)  // 同一个 svc
    // ...
}

// gRPC handler
func (s *ArticleServer) GetArticle(ctx context.Context, req *pb.GetArticleRequest) {
    dto, err := s.svc.GetByID(ctx, req.Id)  // 同一个 svc
    // ...
}
```

**为什么**：业务逻辑只写一次；HTTP 和 gRPC 是适配层。

## 四、错误处理流程

```
domain.ErrNotFound                    （领域层：纯语义）
   ↓ errors.Is(err, ErrNotFound)
application.FromError(err)             （应用层：转译）
   ↓
errcode.ErrNotFound  (HTTP 404)       （公共层：HTTP 状态码）
   ↓
response.Error(c, e)                  （接口层：JSON 序列化）
```

每一层只做**翻译**，不混入其他职责。

## 五、跨切关注点

通过 Gin middleware 实现，统一在 `router.go` 注册：

```
RequestID  →  Logger  →  Recovery  →  CORS  →  Metrics  →  Handler
```

| 中间件 | 职责 |
|---|---|
| RequestID | 提取/生成 X-Request-ID，注入 ctx + 响应头 |
| Logger | Zap 结构化日志，自动带 request_id |
| Recovery | panic 恢复，返回 500 |
| CORS | 跨域配置（gin-contrib/cors） |
| Metrics | Prometheus 请求计数 + 延迟直方图 |
| JWT | `/api/v1/*` 鉴权后挂载 |
| RateLimit | per-IP 或 per-user 限流 |

## 六、配置层级

```
configs/config.yaml           ← 默认配置（入库）
configs/config.local.yaml     ← 本地覆盖（gitignore）
APP_* 环境变量                ← 最高优先级
```

加载顺序：默认 → env-specific → 环境变量覆盖。

Viper 实现：

```go
v.SetConfigName("config")
v.MergeInConfig()  // 可选 config.<env>.yaml
v.SetEnvPrefix("APP")
v.AutomaticEnv()   // 最高优先级
```

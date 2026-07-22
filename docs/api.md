# API 文档

## REST API（HTTP + JSON）

**Base URL**：`http://localhost:9090/api/v1`
**鉴权**：除 `/auth/*` 外所有接口需 `Authorization: Bearer <token>` 头
**Content-Type**：`application/json`

完整可调试文档：`http://localhost:9090/swagger/index.html`（启动后）

### 统一响应格式

```json
{
  "code": 0,
  "message": "OK",
  "data": { ... }
}
```

错误响应：

```json
{
  "code": 40400,
  "message": "not found"
}
```

### 错误码表

| Code | HTTP | 含义 |
|---|---|---|
| 0 | 200 | OK |
| 40000 | 400 | 通用 bad request |
| 40001 | 400 | 无效参数 |
| 40002 | 400 | 校验失败 |
| 40100 | 401 | 未鉴权 |
| 40101 | 401 | token 过期 |
| 40102 | 401 | token 无效 |
| 40300 | 403 | 禁止访问 |
| 40400 | 404 | 资源不存在 |
| 40900 | 409 | 冲突 |
| 42900 | 429 | 限流 |
| 50000 | 500 | 服务器内部错误 |
| 50300 | 503 | 服务不可用 |

### 认证

#### POST /auth/login

```bash
curl -X POST http://localhost:9090/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"alice","password":"secret123"}'
```

响应：

```json
{
  "code": 0,
  "message": "OK",
  "data": {
    "access_token": "eyJhbGc...",
    "refresh_token": "eyJhbGc...",
    "expires_in": 3600
  }
}
```

#### POST /auth/refresh

```bash
curl -X POST http://localhost:9090/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{"refresh_token":"eyJhbGc..."}'
```

#### POST /auth/register

```bash
curl -X POST http://localhost:9090/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"alice","password":"secret123"}'
```

### Authors

#### POST /authors

```bash
curl -X POST http://localhost:9090/api/v1/authors \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice","email":"alice@example.com"}'
```

#### GET /authors

```bash
curl "http://localhost:9090/api/v1/authors?limit=20&offset=0" \
  -H "Authorization: Bearer $TOKEN"
```

#### GET /authors/{id}

```bash
curl http://localhost:9090/api/v1/authors/1 \
  -H "Authorization: Bearer $TOKEN"
```

#### PUT /authors/{id}

```bash
curl -X PUT http://localhost:9090/api/v1/authors/1 \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice Updated"}'
```

#### DELETE /authors/{id}

```bash
curl -X DELETE http://localhost:9090/api/v1/authors/1 \
  -H "Authorization: Bearer $TOKEN"
```

### Articles

#### POST /articles

```bash
curl -X POST http://localhost:9090/api/v1/articles \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "My first article",
    "content": "Hello world",
    "author_id": 1
  }'
```

#### GET /articles（cursor 分页）

```bash
# 第一页
curl "http://localhost:9090/api/v1/articles?limit=10" \
  -H "Authorization: Bearer $TOKEN"

# 后续页：传 cursor
curl "http://localhost:9090/api/v1/articles?limit=10&cursor=20" \
  -H "Authorization: Bearer $TOKEN"
```

响应：

```json
{
  "code": 0,
  "data": {
    "items": [
      {
        "id": 10,
        "title": "...",
        "content": "...",
        "author_id": 1,
        "author_name": "Alice",
        "created_at": "2026-07-21T10:00:00Z",
        "updated_at": "2026-07-21T10:00:00Z"
      }
    ],
    "next_cursor": "9"
  }
}
```

#### GET /articles/{id}

```bash
curl http://localhost:9090/api/v1/articles/1 \
  -H "Authorization: Bearer $TOKEN"
```

#### PUT /articles/{id}

```bash
curl -X PUT http://localhost:9090/api/v1/articles/1 \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"Updated title"}'
```

#### DELETE /articles/{id}

```bash
curl -X DELETE http://localhost:9090/api/v1/articles/1 \
  -H "Authorization: Bearer $TOKEN"
```

### 健康检查

#### GET /healthz

进程存活探针（无需鉴权）。

```bash
curl http://localhost:9090/healthz
# {"status":"ok"}
```

#### GET /readyz

依赖就绪检查（含 DB ping）。

```bash
curl http://localhost:9090/readyz
# {"status":"ready"} or 503 {"status":"unready",...}
```

### Metrics

#### GET /metrics

Prometheus 文本格式。

```bash
curl http://localhost:9090/metrics
# HELP http_requests_total Total number of HTTP requests ...
# TYPE http_requests_total counter
http_requests_total{method="GET",path="/api/v1/articles",status="200"} 5
```

---

## gRPC API

**地址**：`localhost:9091`
**调试工具**：`grpcurl`、`grpcui`

### 启用 reflection（已默认开启）

```bash
# 列出所有服务
grpcurl -plaintext localhost:9091 list

# 列出方法
grpcurl -plaintext localhost:9091 list article.v1.ArticleService

# 调用 GetArticle
grpcurl -plaintext -d '{"id": 1}' \
  localhost:9091 article.v1.ArticleService/GetArticle

# GUI 调试
grpcui -plaintext localhost:9091
```

### 服务定义

```protobuf
service ArticleService {
  rpc GetArticle(GetArticleRequest) returns (GetArticleResponse);
  rpc CreateArticle(CreateArticleRequest) returns (CreateArticleResponse);
  rpc ListArticles(ListArticlesRequest) returns (ListArticlesResponse);
}
```

完整 `.proto` 定义见 `api/proto/article/v1/article.proto`。

> **注意**：v2 仅示例规模实现，AuthorService / AuthService 暂为 stub。

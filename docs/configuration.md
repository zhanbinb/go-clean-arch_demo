# 配置说明

## 配置加载顺序

优先级从低到高：

```
1. configs/config.yaml              ← 默认配置（已入库）
2. .env 文件                         ← 本地覆盖（gitignored，通过 godotenv 加载）
3. configs/config.<APP_ENV>.yaml    ← 环境特定覆盖（已入库或 gitignored）
4. APP_<KEY> shell 环境变量          ← 最高优先级（容器/CI 注入）
```

`config.Load(env)` 实现：
1. `godotenv.Load()` 读 `.env` 进 os.Environ（缺失不报错，存在但解析失败才报错）
2. Viper 读 `configs/config.yaml`（必需）
3. 如果 `env != ""`，合并 `configs/config.<env>.yaml`（可选，缺失不报错）
4. Viper `AutomaticEnv()` 启用 `APP_` 前缀的 env vars，从 os.Environ 读
5. Unmarshal 到 Config 结构

### 验证示例

测试场景（`config.yaml` 默认 `jwt.ttl = 1h`）：

| `.env` 内容 | shell 环境变量 | 实际生效 TTL | 来源 |
|---|---|---|---|
| （无） | （无） | 1h | config.yaml |
| `APP_JWT_TTL=99h` | （无） | 99h | .env 覆盖 config.yaml |
| `APP_JWT_TTL=99h` | `APP_JWT_TTL=5s` | 5s | shell 覆盖 .env |

### .env 文件约定

- 文件名：`.env`（无扩展）
- 路径：项目根目录（与 `go.mod` 同级）
- 格式：`KEY=VALUE`，每行一个，不支持引号嵌套或多行
- 注释：以 `#` 开头
- 缺失：`.env` 不存在时 godotenv 静默跳过（生产用环境变量即可）
- 模板：`.env.example`（已入库，含完整字段说明）

## 文件清单

| 文件 | 跟踪状态 | 用途 |
|---|---|---|
| `configs/config.yaml` | ✅ 入库 | 默认配置，所有环境通用 |
| `configs/config.local.yaml.example` | ✅ 入库 | 本地覆盖模板（不含真实值） |
| `configs/config.local.yaml` | ❌ gitignore | 个人本地覆盖，**不**入库 |
| `configs/config.prod.yaml.example` | ✅ 入库 | 生产环境模板（不含真实 secret） |
| `configs/config.prod.yaml` | ❌ 不存在 | 如需创建则用上面 example 复制 |
| `.env.example` | ✅ 入库 | 所有可用的 `APP_*` 环境变量 |
| `.env` | ❌ gitignore | 本地覆盖，godotenv 自动加载，**不**入库 |

## 快速开始

```bash
# 1. 复制 example 文件
cp .env.example .env
cp configs/config.local.yaml.example configs/config.local.yaml

# 2. 编辑填入本地值（开发一般是默认值就能跑）
$EDITOR .env

# 3. 启动时指定环境
APP_ENV=local ./bin/rest      # 读 config.yaml + config.local.yaml
APP_ENV=prod ./bin/rest       # 读 config.yaml + config.prod.yaml
./bin/rest                    # 默认（只读 config.yaml）
```

## 环境变量完整列表

所有变量以 `APP_` 为前缀（Viper 自动识别）：

### Runtime
| 变量 | 默认 | 说明 |
|---|---|---|
| `APP_ENV` | （空） | 加载 config.<env>.yaml 的依据 |

### Server
| 变量 | 默认 | 说明 |
|---|---|---|
| `APP_SERVER_HTTP_PORT` | 9090 | HTTP 监听端口 |
| `APP_SERVER_GRPC_PORT` | 9091 | gRPC 监听端口 |
| `APP_SERVER_MODE` | debug | `debug`/`release`/`test` |

### Database
| 变量 | 默认 | 说明 |
|---|---|---|
| `APP_DATABASE_HOST` | 127.0.0.1 | MySQL 主机 |
| `APP_DATABASE_PORT` | 3306 | MySQL 端口 |
| `APP_DATABASE_USER` | app | 用户名 |
| `APP_DATABASE_PASSWORD` | app | 密码（生产用 secret 注入） |
| `APP_DATABASE_NAME` | article | 数据库名 |
| `APP_DATABASE_MAX_OPEN_CONNS` | 50 | 连接池上限 |
| `APP_DATABASE_MAX_IDLE_CONNS` | 10 | 空闲连接上限 |
| `APP_DATABASE_CONN_MAX_LIFETIME` | 3600 | 连接最大存活时间（秒） |
| `APP_DATABASE_LOG_LEVEL` | warn | GORM 日志：silent/error/warn/info |

### JWT
| 变量 | 默认 | 说明 |
|---|---|---|
| `APP_JWT_SECRET` | （必需） | **至少 16 字节**，生产用 `openssl rand -base64 48` |
| `APP_JWT_TTL` | 1h | access token 有效期（Go duration 格式） |
| `APP_JWT_REFRESH_TTL` | 24h | refresh token 有效期 |

### Logging
| 变量 | 默认 | 说明 |
|---|---|---|
| `APP_LOG_LEVEL` | info | debug/info/warn/error |
| `APP_LOG_FORMAT` | console | console（开发）/ json（生产） |

### CORS
| 变量 | 默认 | 说明 |
|---|---|---|
| `APP_CORS_ALLOW_ORIGINS` | * | 逗号分隔的 origin 列表 |
| `APP_CORS_ALLOW_METHODS` | GET,POST,... | 逗号分隔 |
| `APP_CORS_ALLOW_HEADERS` | Origin,... | 逗号分隔 |
| `APP_CORS_EXPOSE_HEADERS` | X-Request-ID | 逗号分隔 |
| `APP_CORS_ALLOW_CREDENTIALS` | false | true/false |
| `APP_CORS_MAX_AGE` | 43200 | 预检缓存秒数 |

### Rate limit
| 变量 | 默认 | 说明 |
|---|---|---|
| `APP_RATELIMIT_ENABLED` | true | true/false |
| `APP_RATELIMIT_RPS` | 100 | 每秒令牌数 |
| `APP_RATELIMIT_BURST` | 200 | 桶容量 |
| `APP_RATELIMIT_CLEANUP_INTERVAL` | 300 | visitor 清理间隔（秒） |
| `APP_RATELIMIT_DIMENSION` | ip | `ip` / `user` |

### Swagger
| 变量 | 默认 | 说明 |
|---|---|---|
| `APP_SWAGGER_ENABLED` | true | 是否暴露 /swagger UI |
| `APP_SWAGGER_PATH` | /swagger | UI 路径前缀 |

## 注意事项

1. **不要把 `.env` 入库**：含密码、secret
2. **不要把 `configs/config.local.yaml` 入库**：个人配置
3. **生产 secret 通过 env 注入**：`config.prod.yaml` 不放密码字段
4. **改 `config.yaml` 要走 PR review**：是共享默认
5. **YAML 用小写下划线命名**：Viper mapstructure 自动映射 Go 字段

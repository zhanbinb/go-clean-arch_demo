# 项目基础设施文档

本文档梳理项目中除 Clean Architecture 业务代码外的所有基础设施：构建、容器化、配置、静态检查等。

---

## 一、文件清单与职责

| 文件 | 类别 | 一句话作用 |
|---|---|---|
| `Makefile` | 构建入口 | **所有日常命令的入口**（build/test/lint/up/down） |
| `misc/make/tools.Makefile` | 工具安装 | 把 migrate/air/gotestsum/tparse/mockery/golangci-lint 下载到 `bin/` |
| `misc/make/help.Makefile` | 自文档化 | `make help` 自动从注释生成命令说明 |
| `compose.yaml` | 容器编排 | Docker Compose：本地起 web + MySQL |
| `Dockerfile` | 镜像构建 | 多阶段构建（builder + alpine 分发） |
| `.dockerignore` | 构建上下文 | 排除 `engine`、`*.out` 等构建产物 |
| `.air.toml` | 热重载 | `cosmtrek/air` 的配置（监听 `.go/.yaml/.toml` 改动） |
| `.golangci.yaml` | 静态检查 | golangci-lint 启用 15 个 linter 的规则 |
| `.gitignore` | VCS 排除 | 忽略 vendor/、build artifacts、`.env` |
| `.env` / `example.env` | 配置 | godotenv 加载的环境变量模板与本地值 |

---

## 二、Makefile —— 日常命令入口

### 2.1 设计上的几个亮点

**亮点 1：环境变量自动 export**

```makefile
export PATH   := $(PWD)/bin:$(PATH)   # bin/ 加入 PATH
export SHELL  := bash
export OSTYPE := $(shell uname -s | tr A-Z a-z)   # darwin/linux
export ARCH   := $(shell uname -m)                 # arm64/amd64
```

这意味着在 Makefile 任何位置调 `golangci-lint`、`air` 等命令时，**会优先用 `bin/` 里的本地版本**——不污染全局 PATH。

**亮点 2：注释里的 `##` 是 `make help` 的元数据**

每个 target 后跟 `## 描述`：
```makefile
build: ## Builds binary
dev-env: ## Bootstrap Environment (with a Docker-Compose help).
```

`misc/make/help.Makefile` 里的 `help` target 用 awk 抓这些注释，生成可读的彩色帮助文档。

**亮点 3：被注释掉的 migrate 命令**

文件下半部分有一大块 migrate 命令全部被 `#` 注释掉，旁边注释解释："this not relevant for the project, we load the DB data from the SQL file"——这个项目**用 SQL 初始化脚本（`article.sql`）而不是版本化迁移**。这是个有意为之的简化选择。

### 2.2 命令分组（按 `make help` 实际输出）

| 分组 | 命令 | 作用 |
|---|---|---|
| **开发环境** | `make up` | 拉起 MySQL + 启动 air 热重载（**最常用**） |
|  | `make down` | docker-compose down |
|  | `make destroy` | down + 删 volumes + 清临时文件 |
|  | `make dev-env` | 仅启动 MySQL 容器 |
|  | `make dev-env-test` | 启动 MySQL + 构建镜像 + 启动 web 容器 |
|  | `make dev-air` | 启动 air（持续编译） |
|  | `make install-deps` | 安装本地开发工具链到 `bin/` |
| **代码动作** | `make build` | `go build -trimpath -o engine ./app/` |
|  | `make build-race` | 同上 + `-race` 竞态检测 |
|  | `make lint` | `golangci-lint run -c .golangci.yaml ./...` |
|  | `make go-generate` | `go generate ./...`（生成 mocks） |
|  | `make tests` | `gotestsum` 跑测试（带 race + 覆盖率） |
|  | `make tests-complete` | 跑完测试 + tparse 输出详细表格 |
| **Docker** | `make image-build` | `docker build` → tag 为 `go-clean-arch` |
| **清理** | `make clean` | 清构建产物 + 悬空镜像 |
|  | `make clean-artifacts` | 删 `*.out` |
|  | `make clean-docker` | `docker image prune -f` |

### 2.3 `make tests` 实际发生了什么

```makefile
TESTS_ARGS := --format testname --jsonfile gotestsum.json.out
TESTS_ARGS += --max-fails 2
TESTS_ARGS += -- ./...
TESTS_ARGS += -test.parallel 2
TESTS_ARGS += -test.count    1
TESTS_ARGS += -test.failfast
TESTS_ARGS += -test.coverprofile   coverage.out
TESTS_ARGS += -test.timeout        5s
TESTS_ARGS += -race

tests: $(GOTESTSUM)
	@ gotestsum $(TESTS_ARGS) -short
```

- **gotestsum**：把 `go test` 输出格式化成人类可读 + JSON 双格式；支持重跑失败用例
- **tparse**：把 gotestsum 生成的 JSON 转成漂亮的 ASCII 表格（`make tests-complete` 才用到）
- **-race**：开启竞态检测器，**双倍编译时间**但能查出并发 bug
- **-test.coverprofile**：生成 coverage.out，可用 `go tool cover -html=coverage.out` 看
- **-test.failfast + --max-fails 2**：失败 2 次就停

---

## 三、`misc/make/tools.Makefile` —— 工具链管理

**模式**：每个工具一个 target，**如果系统已有就用系统的，否则下载到 `bin/`**。

```makefile
MIGRATE := $(shell command -v migrate || echo "bin/migrate")
migrate: bin/migrate ## Install migrate (database migration)

bin/migrate: VERSION := 4.17.0
bin/migrate: GITHUB  := golang-migrate/migrate
bin/migrate: ARCHIVE := migrate.$(OSTYPE)-$(ARCH).tar.gz
bin/migrate: bin
	@ printf "Install migrate... "
	@ curl -Ls $(shell echo $(call github_url) | tr A-Z a-z) | tar -zOxf - ./migrate > $@ && chmod +x $@
	@ echo "done."
```

每个工具的安装逻辑都一样：
1. `command -v <tool>` 查 PATH，没有就 fallback 到 `bin/<tool>`
2. 从 GitHub release 下载对应 `darwin/linux-amd64/arm64` 的 tar.gz
3. 解压到 `bin/<tool>` 并 chmod +x

**这套机制等价于一个迷你版的 `asdf`/`rtx`**：每个工具固定版本、跨平台、与项目绑定，不污染系统。

| 工具 | 固定版本 | 用途 |
|---|---|---|
| `migrate` | 4.17.0 | （注释掉，未启用）数据库 schema 迁移 |
| `air` | 1.49.0 | 热重载 |
| `gotestsum` | 1.11.0 | 测试输出格式化 |
| `tparse` | 0.13.2 | 把 JSON 测试输出转表格 |
| `mockery` | 2.42.0 | 根据 `//go:generate` 生成 mock |
| `golangci-lint` | 1.56.2 | 静态检查 |

> ⚠️ **潜在坑**：所有下载 URL 都是 `https://github.com/...`。在 GFW 网络下会失败。开发者需要配置代理或在 Makefile 里把 URL 改成镜像。

---

## 四、`misc/make/help.Makefile` —— 自文档化

```makefile
help: dep-gawk
	@cat $(MAKEFILE_LIST) | \
		grep -E '^# ~~~ .*? [~]+$$|^[a-zA-Z0-9_-]+:.*?## .*$$' | \
		awk '{if ( $$1=="#" ) { \
			match($$0, /^# ~~~ (.+?) [~]+$$/, a);\
			{print "\n", a[1], ""}\
		} else { \
			match($$0, /^([0-9a-zA-Z_-]+):.*?## (.*)$$/, a); \
			{printf "  - \033[32m%-20s\033[0m %s\n",   a[1], a[2]} \
 		}}'
```

**这段 awk 做了什么**：

1. 用 grep 抓两种行：
   - `# ~~~ 分组名 ~~~~` — 分组标题
   - `target: ... ## 描述` — 命令行
2. 用 awk 的 `match` + 命名捕获：
   - 分组行：`# ~~~ 后面的字符串 ~`
   - 命令行：`target` + `## 后的描述`
3. 输出彩色：`-` 前缀 + 绿色 target 名 + 描述

`dep-gawk` target 还**跨平台装 gawk**（brew / apt / yum / apk），因为 macOS 默认是 BSD awk 而这个脚本需要 GNU awk 的扩展正则。

---

## 五、`compose.yaml` —— 本地编排

```yaml
version: "3.7"
services:
  web:
    image: go-clean-arch              # ← 用现成镜像，不在这里 build
    container_name: article_management_api
    ports:
      - 9090:9090
    depends_on:
      mysql:
        condition: service_healthy    # 等 mysql 健康检查通过才启动
    volumes:
      - ./config.json:/app/config.json  # ⚠️ 这个文件不存在！

  mysql:
    image: mysql:8.3
    container_name: go_clean_arch_mysql
    command: mysqld --user=root
    volumes:
      - ./article.sql:/docker-entrypoint-initdb.d/init.sql   # 首次启动时自动建表 + 灌数据
    ports:
      - 3306:3306
    environment:
      - MYSQL_DATABASE=article
      - MYSQL_USER=user
      - MYSQL_PASSWORD=password
      - MYSQL_ROOT_PASSWORD=root
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost"]
      timeout: 5s
      retries: 10
```

### 5.1 关键设计点

**① `web` 用 `image:` 不带 `build:`**

容器镜像必须先用 `make image-build` 打好。`make up` 走的是另一条路径：只起 MySQL + 本地 air，不依赖这个文件。

**② MySQL 用 `docker-entrypoint-initdb.d/init.sql` 灌数据**

`article.sql` 在容器首次启动时被自动执行（建表 + 插数据）。这是 MySQL 镜像的官方约定，省去了手写 migrate。

**③ 健康检查决定启动顺序**

```
mysql (healthcheck: mysqladmin ping)
   ↓ healthy
web (depends_on: condition: service_healthy)
```

不健康就一直等，最多 `5s × 10 = 50s`，避免 web 比 DB 早起来导致连不上。

**④ ⚠️ 一个小坑：`./config.json:/app/config.json`**

`config.json` 文件**在仓库里不存在**，但代码 `app/main.go` 用的是 `os.Getenv`（通过 godotenv 加载 `.env`），不读 `config.json`。这个 volume mount 是**遗留配置**——很可能是早期版本的兼容。如果跑 `docker-compose up web`，容器会因为找不到 `config.json` 报错。**建议删掉这一行**。

---

## 六、`Dockerfile` —— 多阶段构建

```dockerfile
# Stage 1: Builder
FROM golang:1.20.7-alpine3.17 as builder

RUN apk update && apk upgrade && \
    apk --update add git make bash build-base   # 装构建工具

WORKDIR /app

COPY . .                  # 整个项目拷贝进 builder

RUN make build            # ← 关键：用 Makefile 而不是直接 go build

# Stage 2: Distribution
FROM alpine:latest

RUN apk update && apk upgrade && \
    apk --update --no-cache add tzdata && \    # ← 时区数据
    mkdir /app 

WORKDIR /app 

EXPOSE 9090

COPY --from=builder /app/engine /app/   # ← 只拷贝构建产物

CMD /app/engine
```

### 6.1 设计解读

**① 两阶段的目的：减小镜像体积**

- Builder 阶段：~1GB（Go 工具链 + 源码 + 构建缓存）
- Distribution 阶段：~10MB（alpine + 时区数据 + 单个二进制）

最终镜像里**没有 Go 编译器、没有源码**，攻击面小。

**② 装 `tzdata` 是因为 MySQL DSN 用了 `loc=Asia/Shanghai`**

main.go 里数据库连接的时区字符串要靠 `/usr/share/zoneinfo/` 下的数据文件解析。Alpine 默认不带，必须显式 `apk add tzdata`，否则 MySQL 驱动会报"unknown time zone"。

**③ `make build` 而不是 `go build`**

让镜像复用项目自身的 Makefile，保持构建命令单一来源。`make build` 包含 `-trimpath`（去掉绝对路径，便于 reproducible build）。

**④ 可以改进的地方（学习视角）**

```dockerfile
# 当前：
COPY . .                     # 拷贝全部文件，含 .git, *.md, .env ...
RUN make build

# 改进（用 builder pattern）：
COPY go.mod go.sum ./
RUN go mod download            # 单独缓存依赖层
COPY . .
RUN make build                 # 源码变化才重新编译
```

这样依赖改动比源码改动频率低得多，能充分利用 Docker layer cache。

---

## 七、`.dockerignore` —— 镜像构建上下文

```
engine
*.out
```

只有 2 行。排除：
- `engine` — 构建产物（避免 COPY 时把旧二进制带进 builder）
- `*.out` — 测试输出（`coverage.out` 等）

⚠️ **缺很多东西**：`vendor/`、`.git/`、`bin/`、`tmp_app/`、`.env`、`README.md`、`docs/` 都**没排除**。当前 builder 阶段会拿到这些无用的文件，浪费传输时间和 layer 空间。

推荐的 `.dockerignore`：

```gitignore
.git
.github
.bin
*.out
*.md
!README.md
docs/
tmp_app/
.env
example.env
.dockerignore
Dockerfile
```

---

## 八、`.air.toml` —— 热重载

```toml
root = "."
tmp_dir = "tmp_app"

[build]
cmd = "go build -o ./tmp_app/app/engine ./app/."
bin = "tmp_app/app"
full_bin = "./tmp_app/app/engine"
log = "air_errors.log"
include_ext = ["go", "yaml", "toml"]      # 监听这 3 种文件
exclude_regex = ["_test\\.go"]             # 测试文件改动不触发
exclude_dir = ["tmp_app", "tmp"]           # 不监听自己的产物目录
delay = 1000                               # 防抖 1s

[misc]
clean_on_exit = true                       # Ctrl+C 时清掉 tmp_app
```

### 工作流

1. 你 `make up` → `make dev-env` 起 MySQL → `make dev-air` 起 air
2. air 监控 `.go / .yaml / .toml` 文件
3. 你改 `article/service.go` → air 调用 `go build -o ./tmp_app/app/engine ./app/.`
4. 构建成功后自动重启 `./tmp_app/app/engine`
5. HTTP 请求来 → 新二进制接管 → 无需手动重启

**注意**：`include_ext` 只有 `go/yaml/toml`，**改 `.env` 不会触发重启**——这是有意的，因为 env 是进程启动时读的，air 重启进程自然生效。但如果有谁改了 `.env` 想立即生效，需要手动 Ctrl+C air 再起。

`tmp_app/` 也在 `.gitignore` 里，所以这些热重载产物不入版本库。

---

## 九、`.golangci.yaml` —— 静态检查

启用 15 个 linter：

| Linter | 检查什么 |
|---|---|
| `errcheck` | 返回的 error 是否被忽略（除 `_ =` 显式忽略外） |
| `funlen` | 函数行数 / 语句数过多（lines:150 / statements:80） |
| `goconst` | 重复字符串应提为常量 |
| `gocyclo` | 圈复杂度 ≥50 报警 |
| `gosec` | 安全漏洞（SQL 注入、硬编码密码等） |
| `gosimple` | 可简化的代码 |
| `govet` | go vet + check-shadowing（变量遮蔽） |
| `ineffassign` | 无效赋值 |
| `lll` | 行长 ≤ 160 |
| `misspell` | 英文拼写 |
| `revive` | golint 替代品 |
| `staticcheck` | 100+ 高级检查 |
| `typecheck` | 类型检查 |
| `unconvert` | 不必要的类型转换 |
| `unparam` | 未使用的函数参数 |
| `unused` | 未使用的变量/函数/类型 |

显式 `disable-all: true` + `enable: [...]` 是为了**不让新版本 golangci-lint 自动开启新 linter**导致 CI 突然挂掉。

⚠️ **`lll: 160` 偏宽松**：Go 社区惯例是 120。这个项目允许长行，可能是历史遗留。

`skip-files: ".*_test\\.go$"` 让测试文件不参与 lint——合理，避免在 test 里写"故意复杂"的代码被 lint 警告。

---

## 十、`.gitignore` / `.env` / `example.env`

**`.gitignore`**：

| 模式 | 用途 |
|---|---|
| `vendor/` | 不使用 vendor 模式（go modules 走 GOPATH） |
| `article_clean` | 临时清洗脚本目录（项目里没出现过，应该是遗留） |
| `_*` | 以 `_` 开头的临时文件 |
| `*.test` | go test 产物 |
| `.DS_Store` | macOS 元数据 |
| `engine` | Makefile build 产物 |
| `bin/` | tools.Makefile 装工具的目录 |
| `*.out` | 测试覆盖率、gotestsum 输出 |
| `tmp_app/` | air 的工作目录 |
| `.env` | ★ **本地凭证不入库** |

**`.env` / `example.env`**：

- `example.env` 入库（**模板**，只包含默认值如 `localhost:3306`）
- `.env` 被 `.gitignore` 排除（**本地实际值**，可能含真密码）

`app/main.go` 的 `init()` 调 `godotenv.Load()` 加载 `.env`，缺文件就 fallback 到 OS 环境变量（生产环境友好）。

---

## 十一、典型工作流汇总

| 场景 | 命令序列 |
|---|---|
| **首次启动** | `cp example.env .env` → `make install-deps` → `make up` |
| **日常开发** | `make dev-air`（在另一个 terminal 跑 `make dev-env` 起 DB） |
| **跑测试** | `make tests` 或 `make tests-complete` |
| **跑 lint** | `make lint` |
| **生成 mock** | `make go-generate` |
| **构建并运行容器** | `make image-build` → `make dev-env-test` |
| **完全清理** | `make destroy` |

---

## 十二、可以改进的地方（学习要点）

| # | 问题 | 文件 | 改进建议 |
|---|---|---|---|
| 1 | `compose.yaml` 引用了不存在的 `config.json` | compose.yaml | 删除该 volume 行 |
| 2 | `.dockerignore` 几乎为空 | .dockerignore | 加 `.git/`、`bin/`、`tmp_app/`、`.env` 等 |
| 3 | Dockerfile 无 layer cache 优化 | Dockerfile | 先 COPY go.mod/go.sum + `go mod download` |
| 4 | tools 下载走 github.com，GFW 下失败 | misc/make/tools.Makefile | 加镜像 URL 选项 |
| 5 | migrate 命令被注释但仍下载到 bin/ | Makefile + tools.Makefile | 要么启用 migrate，要么从 tools 里删掉 |
| 6 | `version: "3.7"` 已过时 | compose.yaml | Docker Compose v2 已忽略该字段，建议删除 |
| 7 | `lll: 160` 偏宽松 | .golangci.yaml | 改成 120 更符合 Go 社区规范 |
| 8 | `make build` 不带 `-race`，要手动 `build-race` | Makefile | 文档里说明二选一 |

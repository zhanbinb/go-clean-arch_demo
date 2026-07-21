# 项目文档目录

| 文档 | 内容 |
|---|---|
| [v2-design.md](./v2-design.md) | v2 企业级架构升级完整设计（1854 行，含 17 章） |
| [architecture.md](./architecture.md) | 分层架构、依赖方向、关键设计决策 |
| [api.md](./api.md) | REST + gRPC 接口文档（含 curl / grpcurl 示例） |
| [infrastructure.md](./infrastructure.md) | Makefile / Docker / air / golangci 等基础设施详解 |
| [configuration.md](./configuration.md) | 配置层级、example 文件、环境变量参考 |
## 快速链接

- REST API Swagger UI：启动后访问 `http://localhost:9090/swagger/index.html`
- gRPC 调试 GUI：`make grpc-ui`（grpcurl/grpcui 已通过 install-tools 安装）
- Prometheus Metrics：`http://localhost:9090/metrics`
- pprof 调试（debug 模式）：`http://localhost:9090/debug/pprof/`

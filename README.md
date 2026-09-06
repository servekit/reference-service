# reference-service

基于 [go-common](https://github.com/servekit/go-common) 的**纯 gRPC** 服务：不监听 HTTP，proto 不带 `google.api.http` 注解；对外 HTTP 面由网关（当前为 testkit-service）提供。

## 构建与运行

服务与数据库迁移合并为同一个二进制（`cmd/server`），通过子命令区分：

| 命令 | 作用 |
|---|---|
| `bin/reference-service` 或 `bin/reference-service serve` | 启动 gRPC 服务（默认） |
| 其他 | 打印用法，exit 2 |

本地开发常用命令（完整 target 见 `Makefile`）：

```bash
make build       # 产出 bin/reference-service
make run         # 本地启动（auto-cp config.example.yaml -> config.yaml）
make regenerate  # = tidy（改 model 后跑；proto 改在 ../api：make -C ../api gen）
make test        # 测试（race + coverage）
make lint        # golangci-lint
make docker-up   # 起完整 docker 栈
```

gRPC 监听 `:19094`。proto 契约在 [`../api`](../api) 仓库（三分文件），本仓库只消费 `github.com/servekit/api/gen/go` 生成代码——本地靠 servekit 根 `go.work` 联动，`go.mod` 里的 require 是远端 pin（`../api` 推送后 `go get github.com/servekit/api/gen/go@<commit> && go mod tidy` 跟进）。不监听 HTTP（gRPC-only）。

## 配置

`config.example.yaml` 是**纯结构**——每个值都是 `${VAR}` 占位符，由 configx `WithExpandEnv` 从进程环境展开。`.env.example` 是 **docker-compose 取向**的默认值源（host 名是 compose 服务名）。前者管结构，后者管值。

**本地跑（`make run`）：**

```bash
cp .env.example .env
# 把 docker host 名改成本地地址（config.yaml 由 make run 自动从 config.example.yaml 拷出）：
#   REFERENCE_SERVICE_DATABASE_HOST: postgres  ->  localhost
make run            # 需要本机 PostgreSQL
```

**docker compose 跑（`make docker-up`）：** 无需改 `.env`——compose 注入全部 env，host 名正好是服务名（`postgres`）。改配置值都改 `.env`，不要往 `config.example.yaml` 写字面量。

**构建在受限网络（proxy.golang.org 不通）？** 在 `.env` 加 `GOPROXY=https://goproxy.cn,direct` 再 `make docker-build` / `make docker-up`。

## 测试调用

```bash
# gRPC（按 proto 字段调整 payload）
grpcurl -plaintext -d '{"name":"hi"}' \
  localhost:19094 reference.v1.ReferenceService/CreateReference

```

> 架构规范（分层、枚举、服务间依赖、lifecycle 等）见 `CLAUDE.md`。

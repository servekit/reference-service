# reference-service

基于 [go-common](https://github.com/servekit/go-common) 的**纯 gRPC** 服务：全局静态参考数据的单一权威源——国家+区号、时区、语言、货币、大洲/区域分组，外加基于 libphonenumber 的手机号解析。**数据编译进二进制**（无 DB、无 Redis、无运行时管理，数据随版本发布更新），grpc + module 双模式接入。

## RPC 一览（`reference.v1.ReferenceService`）

| RPC | 说明 |
|---|---|
| `ListCountries` | ISO 3166-1 + E.164 区号 + 国旗 emoji + alpha-3，按请求 locale 名称排序 |
| `ListTimezones` | IANA canonical 时区 + 别名（backward links，输入归一用，picker 不展示）+ 国家映射 |
| `ListLanguages` | BCP 47 语言标签（639-1 集合 + zh-Hans/zh-Hant） |
| `ListCurrencies` | ISO 4217 + 符号 + 小数位（minor_units）+ 当前使用国 |
| `ListRegionGroups` | UN M49 层级（顶级=五大洲，不下发世界根），每层直接成员国 |
| `ParsePhone` | libphonenumber 规则：E.164 归一、归属国、号码类型（mobile/fixed/…）；坏输入不报错，`is_valid=false` + `error_reason` |
| `ResolveCodes` | 跨域批量码解析（取实体+验码，一次调用服务注册表单）；时区别名在此归一（`PRC`→`Asia/Shanghai`）；未命中进 `missing_*`，永不报错 |

每个 List/Resolve 响应带 `data_version`（数据生成时间戳）。

## locale 约定（i18n）

- 所有 List / `ResolveCodes` 请求带可选 `locale`（BCP 47）；**缺省 = `zh-Hans`**。
- 实体返回单一 `name`（请求语言的名称）。
- fallback 链：精确匹配 → likelySubtags 展开（`zh-TW`→`zh-Hant`）→ 基语言（`pt-BR`→`pt`）→ `en`；locale 不识别静默落 `en`，**永不因 locale 报错**。
- 编译的 locale 集合（生成器配置，当前 11 语）：`zh-Hans, zh-Hant, ja, ko, en, fr, de, pt, es, ar, ru`。加语言 = 改 `tools/gen` 的 locales 清单 + `internal/data/data.go` 的 `Locales`（两者由生成器自检对齐）→ `make data` → PR → 发版，**API 不变、调用方零改动**。
- 列表顺序按请求 locale 的 CLDR collation 预排（zh 拼音序、en 字母序…）。
- 同屏双语（如 zh+en）= 两次不同 locale 的调用（module 模式进程内直调，微秒级）。

## 数据再生成

```bash
make data   # 从 pinned 上游重生成 internal/data/*_data.go（首次需联网，结果缓存 .cache/）
```

上游：**nyaruka/phonenumbers 元数据**（国家条目集与区号——与 `ParsePhone` 同一份元数据，升级依赖后 `make data` 即同步再生，目录与解析规则永不脱节）、CLDR 45.0.0（unpkg 分包，全部 11 语言的名称）、lukes ISO 3166（alpha-3）、IANA `zone1970.tab` + `backward`。更新流程：`make data` → review diff → PR → 发版。

> 历史注记：迁移期的区号种子表（自 message-service 一次性转写，保证当时前端数据零差异）已于切换到 nyaruka 推导后删除；切换前对拍确认五项零差异（条目/区号/中英名）。

## 构建与运行

```bash
make build       # 产出 bin/reference-service
make run         # 本地启动（auto-cp config.example.yaml -> config.yaml）
make test        # 测试（race + coverage）
make lint        # golangci-lint
make docker-up   # docker 栈
```

gRPC 监听 `:19094`。proto 契约在 [`../api`](../api) 仓库（`reference/v1/` 四文件），本仓库只消费 `github.com/servekit/api/gen/go` 生成代码——本地靠 servekit 根 `go.work` 联动，`go.mod` 的 require 是远端 pin（`../api` 推送后 `go get github.com/servekit/api/gen/go@<commit> && go mod tidy` 跟进）。不监听 HTTP（gRPC-only）。

## 接入方式（grpc / module）

```go
// grpc：
c, _ := referenceservice.NewClient("reference-service:19094")
resp, _ := c.ListCountries(ctx, &referencev1.ListCountriesRequest{Locale: "ja"})

// module（进程内，零网络零序列化）：
hdl, _ := referenceservice.NewModule(&config.Config{})
defer hdl.Stop()
resp, _ = hdl.ListCountries(ctx, &referencev1.ListCountriesRequest{Locale: "ja"})

// 经 lifecycle 解析（testkit 模式）：
svc, _, err := referenceservice.Connect(referenceservice.ConnectConfig{
    Mode: cfg.Mode, Target: cfg.Target, Config: cfg.Config,
}, mgr)
```

无 DB / Redis / 外部依赖——module 模式零基础设施成本。

## 配置

`config.example.yaml` 是**纯结构**（`${VAR}` 占位符，configx `WithExpandEnv` 展开）；`.env.example` 是 docker-compose 取向的默认值源。本地 `make run` 只需 `REFERENCE_SERVICE_SERVER_GRPC_ADDR`。

## 测试调用

```bash
grpcurl -plaintext -d '{"locale":"ja"}' \
  localhost:19094 reference.v1.ReferenceService/ListCountries
grpcurl -plaintext -d '{"raw":"+8613800138000"}' \
  localhost:19094 reference.v1.ReferenceService/ParsePhone
```

> 架构规范（分层、枚举、服务间依赖、lifecycle 等）见 `CLAUDE.md`；设计决策与边界（为什么不上 DB、后续候选数据域）见 `specs/2026-09-06-reference-service-design.md`。

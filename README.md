# CodeX

CodeX 是一个基于 Go 和 Gin 的轻量代理服务，用来管理多个 Codex 会话，并将兼容 OpenAI Responses 风格的请求转发到 Codex 后端。

项目主要提供两类能力：

- 通过 `/codex/cli` 完成设备码登录，拉取并保存会话。
- 通过 `/codex/responses` 统一转发请求，并在多个会话之间轮询切换。

## 功能特性

- 支持 Codex 设备码授权登录
- 本地持久化多个会话到 `session/*.json`
- 自动轮询可用会话，避免单账号集中消耗
- 会话即将过期时自动刷新 token
- 请求 429 时临时禁用对应会话
- 按 API Key 统计 Today / Yesterday / Total token 使用量
- 本地持久化每个 API Key 的 token usage 到独立 JSON 文件
- 内置 API Key 校验
- 提供 Prometheus 指标接口

## 项目结构

```text
.
├── conf/                 # 配置文件
├── doc/                  # 补充文档
├── internal/
│   ├── handler/          # HTTP 路由与处理器
│   ├── logic/            # 授权与代理逻辑
│   ├── middleware/       # 鉴权、CORS、指标中间件
│   ├── svc/              # 服务上下文与会话管理
│   └── types/            # 请求/响应类型定义
├── main.go               # 程序入口
└── makefile              # 构建脚本
```

## 运行要求

- Go 1.25+

## 快速开始

1. 安装依赖并启动服务：

```bash
go run .
```

2. 或使用 make 构建：

```bash
make build
./bin/codex-web -l 127.0.0.1:8000 -f conf/conf.yml
```

## 配置说明

默认配置文件路径为 `conf/conf.yml`。

```yaml
SessionDir: session
TokenUsageDir: token_usage
Secret:
  - your-api-key
Metrics: /your-metrics-path
ModelPlanType:
  gpt-5: pro
```

字段说明：

- `SessionDir`：会话文件目录，设备授权成功后会在该目录下生成 `*.json`
- `TokenUsageDir`：API Key token 使用量持久化目录，会按 `sk.json` 分别保存
- `Secret`：访问接口时使用的 API Key 列表
- `Metrics`：Prometheus 指标路径
- `ModelPlanType`：模型与账号套餐类型的映射，用于按请求模型选择对应 `PlanType` 的会话

建议将 `Secret` 替换为你自己的密钥，不要直接在公网暴露服务。

`ModelPlanType` 的 key 对应请求体中的 `model` 字段，value 会和会话文件中的 `planType` 做包含匹配。例如 `gpt-5: pro` 会让 `model` 为 `gpt-5` 的请求优先选择 `planType` 包含 `pro` 的会话。未配置映射或映射为空时，会在全部可用会话中轮询。

## 接口说明

### 1. 获取设备码并完成登录

访问：

```text
GET /codex/cli?auth=<Secret>
```

账号添加示例：

```text
https://your-domain.example.com/codex/cli?auth=your-api-key
```

页面会展示设备码，并引导打开 OpenAI 授权页完成登录。授权成功后，会话会保存到 `SessionDir` 目录。

### 2. 代理 Responses 请求

访问：

```text
POST /codex/responses
Authorization: Bearer <Secret>
```

服务会将请求转换为 Codex 所需结构，并转发到：

```text
https://chatgpt.com/backend-api/codex/responses
```

请求体使用兼容 OpenAI Responses 的 JSON 结构，例如：

```json
{
  "model": "gpt-5",
  "input": [
    {
      "role": "user",
      "content": "hello"
    }
  ],
  "stream": false
}
```

说明：

- 若 `input` 中包含 `developer` 消息，会被转换为 Codex 的 `instructions`
- 可通过 `X-Session-Affinity` 指定透传的会话亲和标识
- 服务会根据请求体的 `model` 查询 `ModelPlanType`，并在匹配的 `PlanType` 会话中独立轮询；未匹配到模型映射时，在全部可用会话中轮询
- 流式 SSE 响应中的 `response.completed` 会被用来累计当前 API Key 的 token usage

### 3. 查看 Token Usage

访问：

```text
GET /codex/usage?auth=<Secret>
```

或：

```text
GET /codex/usage
Authorization: Bearer <Secret>
```

说明：

- 浏览器访问且 `Accept` 包含 `text/html` 时，返回可视化页面
- 其它客户端默认返回 JSON
- 返回内容包含 `today`、`yesterday`、`total` 三组 token 统计
- 服务启动时会从 `TokenUsageDir` 读取已有 usage，服务关闭时会统一刷回文件

## 监控

Prometheus 指标地址由 `Metrics` 配置决定，例如：

```text
GET /xxxxx/metrics
```

## OpenCode 配置

如果你希望在 OpenCode 中通过自定义 `baseURL` 接入本服务，可以在项目根目录添加或修改 `opencode.json`：

```json
{
  "$schema": "https://opencode.ai/config.json",
  "permission": "allow",
  "provider": {
    "openai": {
      "name": "Codex",
      "options": {
        "baseURL": "https://your-domain.example.com/codex"
      }
    }
  }
}
```

说明：

- `baseURL` 指向本服务的 `/codex` 根路径
- 如果服务开启了 API Key 校验，请确保客户端请求携带对应的 `Authorization: Bearer <Secret>` 或 `auth` 参数

## 构建

```bash
make build
```

构建产物默认输出到 `bin/` 目录，包含多个平台版本。

## Linux 服务

项目内置了常见 Linux 发行版的服务文件示例，目录如下：

```text
deploy/linux-services/
```

包含：

- `systemd/`：适用于 Debian、Ubuntu 以及大多数 systemd 发行版
- `openrc/`：适用于 Alpine Linux

相关文件：

- `deploy/linux-services/systemd/codex-web.service`
- `deploy/linux-services/systemd/codex-web.env.example`
- `deploy/linux-services/openrc/codex-web`
- `deploy/linux-services/openrc/codex-web.confd`

详细安装和启动方式见：

```text
deploy/linux-services/README.md
```

## 许可说明

本项目采用 `PolyForm Noncommercial License 1.0.0`。

说明：该协议强调禁止商业使用，严格来说更接近源码可见许可，并不属于传统意义上的 OSI 开源许可证。

- 允许个人学习、研究、非商业用途使用、修改和分发
- 禁止将本项目或基于本项目的衍生版本用于商业用途

详细条款见 `LICENSE`。

<p align="center">
  <img src="webui/public/ds2api-favicon.svg" width="128" height="128" alt="DS2API icon" />
</p>

# DS2API

<a href="https://trendshift.io/repositories/24508" target="_blank"><img src="https://trendshift.io/api/badge/repositories/24508" alt="ojbkxc%2Fds2api | Trendshift" style="width: 250px; height: 55px;" width="250" height="55"/></a>

[![License](https://img.shields.io/github/license/ojbkxc/ds2api.svg)](LICENSE)
![Stars](https://img.shields.io/github/stars/ojbkxc/ds2api.svg)
![Forks](https://img.shields.io/github/forks/ojbkxc/ds2api.svg)
[![Release](https://img.shields.io/github/v/release/ojbkxc/ds2api?display_name=tag)](https://github.com/ojbkxc/ds2api/releases)
[![Docker](https://img.shields.io/badge/docker-ready-blue.svg)](docs/DEPLOY.md)

[![Deploy on Zeabur](https://zeabur.com/button.svg)](https://zeabur.com/templates/L4CFHP)
[![Deploy with Vercel](https://vercel.com/button)](https://vercel.com/new/clone?repository-url=https://github.com/ojbkxc/ds2api)

语言 / Language: [中文](README.MD) | [English](README.en.md)

�?DeepSeek Web 对话能力转换�?OpenAI、Claude �?Gemini 兼容 API。核心后端以 **Go** 实现，Vercel 流式桥接额外使用少量 Node Runtime，前端为 React WebUI 管理台（源码�?`webui/`，部署时自动构建�?`static/admin`）�?
文档入口：[文档导航](docs/README.md) / [架构说明](docs/ARCHITECTURE.md) / [接口文档](API.md)

【感谢Linux.do社区及GitHub社区各位开发者对项目的支持与贡献�?
## Star History

<a href="https://www.star-history.com/?repos=ojbkxc%2Fds2api&type=date&legend=top-left">
 <picture>
   <source media="(prefers-color-scheme: dark)" srcset="https://api.star-history.com/chart?repos=ojbkxc/ds2api&type=date&theme=dark&legend=top-left" />
   <source media="(prefers-color-scheme: light)" srcset="https://api.star-history.com/chart?repos=ojbkxc/ds2api&type=date&legend=top-left" />
   <img alt="Star History Chart" src="https://api.star-history.com/chart?repos=ojbkxc/ds2api&type=date&legend=top-left" />
 </picture>
</a>

> **重要免责声明**
>
> 本仓库仅供学习、研究、个人实验和内部验证使用，不提供任何形式的商业授权、适用性保证或结果保证�?>
> 作者及仓库维护者不对因使用、修改、分发、部署或依赖本项目而产生的任何直接或间接损失、账号封禁、数据丢失、法律风险或第三方索赔负责�?>
> 请勿将本项目用于违反服务条款、协议、法律法规或平台规则的场景。商业使用前请自行确�?`LICENSE`、相关协议以及你是否获得了作者的书面许可�?
## 目录

- [架构概览（摘要）](#架构概览摘要)
- [核心能力](#核心能力)
- [平台兼容矩阵](#平台兼容矩阵)
- [模型支持](#模型支持)
  - [OpenAI 接口](#openai-接口get-v1models)
  - [Claude 接口](#claude-接口get-anthropicv1models)
  - [Gemini 接口](#gemini-接口)
- [快速开始](#快速开�?
  - [方式一：下�?Release 构建包](#方式一下载-release-构建�?
  - [方式二：Docker 运行](#方式二docker-运行)
  - [方式三：Vercel 部署](#方式三vercel-部署)
  - [方式四：本地源码运行](#方式四本地源码运�?
- [配置说明](#配置说明)
- [鉴权模式](#鉴权模式)
- [并发模型](#并发模型)
- [Tool Call 适配](#tool-call-适配)
- [本地开发抓包工具](#本地开发抓包工�?
- [文档索引](#文档索引)
- [测试](#测试)
- [Release 自动构建（GitHub Actions）](#release-自动构建github-actions)
- [免责声明](#免责声明)

## 架构概览（摘要）

```mermaid
flowchart LR
    Client["🖥�?客户�?/ SDK\n(OpenAI / Claude / Gemini)"]
    Upstream["☁️ DeepSeek API"]

    subgraph DS2API["DS2API 4.x（模块化 HTTP surface + PromptCompat 内核�?]
        Router["chi Router + 中间件\n(RequestID / RealIP / Logger / Recoverer / CORS)"]

        subgraph HTTP["HTTP API surface"]
            OA["OpenAI\nchat / responses / files / embeddings"]
            CA["Claude\n/anthropic/* + /v1/messages"]
            GA["Gemini\n/v1beta/models/* + /v1/models/*"]
            Admin["Admin API\n资源子包"]
            WebUI["WebUI\n/admin（静态托管）"]
            Vercel["Vercel Node Stream\n/v1/chat/completions"]
        end

        subgraph Runtime["运行时核心能�?]
            Compat["PromptCompat\n(API -> 网页纯文本上下文)"]
            Completion["Completion Runtime\n(Session / PoW / Completion)"]
            Turn["AssistantTurn\n(输出语义归一)"]
            Auth["Auth Resolver\n(API key / bearer / x-goog-api-key)"]
            Pool["Account Pool + Queue\n(并发槽位 + 等待队列)"]
            DSClient["DeepSeek Client\n(Session / Auth / Completion / Files)"]
            Pow["PoW 实现\n(�?Go)"]
            Tool["Tool Sieve\n(Go/Node 语义对齐)"]
            History["Current Input File\n(deepseek.txt)"]
        end
    end

    Client --> Router
    Router --> OA & CA & GA
    Router --> Admin
    Router --> WebUI
    Router --> Vercel

    OA --> Compat
    CA & GA --> Compat
    Compat --> Completion
    Completion -.完整上下�?-> History
    Completion --> Turn
    Vercel -.Go prepare.-> Completion
    Vercel -.Node SSE.-> Tool
    Completion --> Auth
    Completion -.账号轮询.-> Pool
    Completion -.工具调用解析.-> Tool
    Completion -.PoW 计算.-> Pow
    Auth --> DSClient
    DSClient --> Upstream
    Upstream --> DSClient
    Turn --> Client
    Vercel --> Client
```

详细架构拆分与目录职责见 [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)�?
- **后端**：Go（`cmd/ds2api/`、`api/`、`internal/`），不依�?Python 运行�?- **前端**：React 管理台（`webui/`），运行时托管静态构建产�?- **部署**：本地运行、Docker、Vercel Serverless、Linux systemd

## 核心能力

| 能力 | 说明 |
| --- | --- |
| OpenAI 兼容 | `GET /v1/models`、`GET /v1/models/{id}`、`POST /v1/chat/completions`、`POST /v1/responses`、`GET /v1/responses/{response_id}`、`POST /v1/embeddings`、`POST /v1/files`、`GET /v1/files/{file_id}` |
| Claude 兼容 | `GET /anthropic/v1/models`、`POST /anthropic/v1/messages`、`POST /anthropic/v1/messages/count_tokens`（及快捷路径 `/v1/messages`、`/messages`�?|
| Gemini 兼容 | `POST /v1beta/models/{model}:generateContent`、`POST /v1beta/models/{model}:streamGenerateContent`（及 `/v1/models/{model}:*` 路径�?|
| Ollama 兼容 | `GET /api/version`、`GET /api/tags`、`POST /api/show` |
| 统一 CORS 兼容 | `/v1/*`、`/anthropic/*`、`/v1beta/models/*`、`/api/*`、`/admin/*` 统一走同一�?CORS 策略；Vercel �?`/v1/chat/completions` �?Node Runtime 也对齐相同放行规则，尽量减少第三方预检请求头限�?|
| 多账号轮询 | 自动 token 刷新（含随机抖动避免同时刷新）、邮箱/手机号双登录方式，401/403 自动禁用账号 |
| 并发队列控制 | 每账号 in-flight 上限 + 等待队列，动态计算建议并发，会话池复用 + parent_message_id 链式连接，72h TTL，单账号 50 消息上限 |
| DeepSeek PoW | �?Go 高性能实现（DeepSeekHashV1），毫秒级响�?|
| Tool Calling | 防泄漏处理：非代码块高置信特征识别、`delta.tool_calls` 早发、结构化增量输出 |
| Admin API | 配置管理、运行时设置热更新、代理管理、账号测�?/ 批量测试、会话清理、导入导出、Vercel 同步、版本检�?|
| WebUI 管理�?| `/admin` 单页应用（中英文双语、深色模式，支持查看服务器端对话记录�?|
| 运维探针 | `GET /healthz`（存活）、`GET /readyz`（就绪） |

OpenAI `/v1/*` 仍是推荐的规范路径；同时支持 `/models`、`/chat/completions`、`/responses`、`/embeddings`、`/files`、`/files/{file_id}` 等根路径快捷路由，方便只配置 DS2API 根地址的第三方客户端�?
## 平台兼容矩阵

| 级别 | 平台 | 当前状�?|
| --- | --- | --- |
| P0 | Codex CLI/SDK（`wire_api=chat` / `wire_api=responses`�?| �?|
| P0 | OpenAI SDK（JS/Python，chat + responses�?| �?|
| P0 | Vercel AI SDK（openai-compatible�?| �?|
| P0 | Anthropic SDK（messages�?| �?|
| P0 | Google Gemini SDK（generateContent�?| �?|
| P1 | LangChain / LlamaIndex / OpenWebUI（OpenAI 兼容接入�?| �?|

## 模型支持

### OpenAI 接口（`GET /v1/models`�?
| 模型类型 | 模型 ID | thinking | search |
| --- | --- | --- | --- |
| default | `deepseek-v4-flash` | 默认开启，可由请求参数控制 | �?|
| default | `deepseek-v4-flash-nothinking` | 永久关闭，不受请求参数影�?| �?|
| expert | `deepseek-v4-pro` | 默认开启，可由请求参数控制 | �?|
| expert | `deepseek-v4-pro-nothinking` | 永久关闭，不受请求参数影�?| �?|
| default | `deepseek-v4-flash-search` | 默认开启，可由请求参数控制 | �?|
| default | `deepseek-v4-flash-search-nothinking` | 永久关闭，不受请求参数影�?| �?|
| expert | `deepseek-v4-pro-search` | 默认开启，可由请求参数控制 | �?|
| expert | `deepseek-v4-pro-search-nothinking` | 永久关闭，不受请求参数影�?| �?|
| vision | `deepseek-v4-vision` | 默认开启，可由请求参数控制 | �?|
| vision | `deepseek-v4-vision-nothinking` | 永久关闭，不受请求参数影�?| �?|

除原生模型外，也支持常见 alias 输入（如 `gpt-4.1`、`gpt-5`、`gpt-5-codex`、`o3`、`claude-*`、`gemini-*` 等），但 `/v1/models` 返回的是规范化后�?DeepSeek 原生模型 ID。若 alias 名本身追�?`-nothinking` 后缀，也会映射到对应的强制关思考模型。完�?alias 行为�?[API.md](API.md#模型-alias-解析策略) �?`config.example.json` 为准�?当前上游视觉模型只暴�?`vision` 通道，不提供独立的联网搜索视觉变体�?
### Claude 接口（`GET /anthropic/v1/models`�?
| 当前常用模型 | 默认映射 |
| --- | --- |
| `claude-sonnet-4-6` | `deepseek-v4-flash` |
| `claude-sonnet-4-6-nothinking` | `deepseek-v4-flash-nothinking` |
| `claude-haiku-4-5`（兼�?`claude-3-5-haiku-latest`�?| `deepseek-v4-flash` |
| `claude-haiku-4-5-nothinking` | `deepseek-v4-flash-nothinking` |
| `claude-opus-4-6` | `deepseek-v4-pro` |
| `claude-opus-4-6-nothinking` | `deepseek-v4-pro-nothinking` |

可通过配置中的 `model_aliases` 覆盖映射关系；若请求模型名带 `-nothinking`，会在最终映射结果上强制追加无思考语义�?`/anthropic/v1/models` 除上述主别名外，还会返回 Claude 4.x snapshots�?.x 历史模型 ID 与常�?alias，便于旧客户端直接兼容�?
#### Claude Code 接入避坑（实测）

- `ANTHROPIC_BASE_URL` 推荐直接指向 DS2API 根地址（例�?`http://127.0.0.1:5001`），Claude Code 会请�?`/v1/messages?beta=true`�?- `ANTHROPIC_API_KEY` 需要与 `config.json` �?`keys` 一致；建议同时保留常规 key �?`sk-ant-*` 形�?key，兼容不同客户端校验习惯�?- 若系统设置了代理，建议对 DS2API 地址配置 `NO_PROXY=127.0.0.1,localhost,<你的主机IP>`，避免本地回环请求被代理拦截�?- 如遇“工具调用输出成文本、未执行”问题，请优先检查模型输出是否为推荐的半角管道符 DSML 工具块：`<|DSML|tool_calls><|DSML|invoke name="..."><|DSML|parameter name="...">...`。兼容层也接受旧�?canonical XML：`<tool_calls><invoke name="..."><parameter name="...">...`；旧�?`<tools>` / `<tool_call>` / `<tool_name>` / `<param>`、`<function_call>`、`tool_use` 或纯 JSON `tool_calls` 片段不会执行，会作为普通文本处理�?
### Gemini 接口

Gemini 适配器将模型名通过 `model_aliases` 或内置精�?alias 映射�?DeepSeek 原生模型（覆�?`gemini-2.5-*`、`gemini-3*`、`gemini-pro-vision` 等常见名称），支�?`generateContent` �?`streamGenerateContent` 两种调用方式，并完整支持 Tool Calling（`functionDeclarations` �?`functionCall` 输出）。若 Gemini 模型名带 `-nothinking` 后缀，例�?`gemini-2.5-pro-nothinking`，会映射到对应的强制关闭思考模型�?
## 快速开�?
### 部署方式优先级建�?
推荐按以下顺序选择部署方式�?
1. **下载 Release 构建包运�?*：最省事，产物已编译完成，最适合大多数用户�?2. **Docker / GHCR 镜像部署**：适合需要容器化、编排或云环境部署�?3. **Vercel 部署**：适合已有 Vercel 环境且接受其平台约束的场景�?4. **本地源码运行 / 自行编译**：适合开发、调试或需要自行修改代码的场景�?
### 通用第一步（所有部署方式）

�?`config.json` 作为唯一配置源（推荐做法）：

```bash
cp config.example.json config.json
# 编辑 config.json
```

后续部署建议�?- 本地运行：直接读�?`config.json`
- Docker / Vercel：由 `config.json` 生成 `DS2API_CONFIG_JSON`（Base64）注入环境变量，也可以直接写原始 JSON

WebUI 管理台里的“全量配置模板”也直接复用同一�?`config.example.json`，所以更新示例文件后，前端模板会自动保持一致�?
### 方式一：下�?Release 构建�?
每次发布 Release 时，GitHub Actions 会自动构建多平台二进制包�?
```bash
# 下载对应平台的压缩包�?tar -xzf ds2api_<tag>_linux_amd64.tar.gz
cd ds2api_<tag>_linux_amd64
cp config.example.json config.json
# 编辑 config.json
./ds2api
```

### 方式二：Docker 运行

```bash
# 1. 准备环境变量和配置文�?cp .env.example .env
cp config.example.json config.json

# 2. 编辑 .env（至少设�?DS2API_ADMIN_KEY；如需修改宿主机端口，可额外设�?DS2API_HOST_PORT�?#    DS2API_ADMIN_KEY=请替换为强密�?
# 3. 启动
docker-compose up -d

# 4. 查看日志
docker-compose logs -f
```

默认 `docker-compose.yml` 会把宿主�?`6011` 映射到容器内�?`5001`。如果你希望直接对外暴露 `5001`，请设置 `DS2API_HOST_PORT=5001`（或者手动调�?`ports` 配置）�?同时默认�?`./config.json` 挂载到容�?`/data/config.json`，并设置 `DS2API_CONFIG_PATH=/data/config.json`，用于避�?`/app` 只读导致运行�?token 持久化失败�?镜像会预创建 `/data` 并授权给�?root �?`ds2api` 用户；如果使用单文件 bind mount，请确保宿主�?`config.json` 对容器用户可读写，例�?`chmod 644 config.json`�?
更新镜像：`docker-compose up -d --build`

#### 方式二.2：直接拉取镜像运行

**从 GHCR 拉取（推荐，无需额外配置）**：
```bash
docker pull ghcr.io/ojbkxc/ds2api:latest

docker run -d \
  --name ds2api \
  -p 6011:5001 \
  -v ./config.json:/data/config.json \
  -e DS2API_CONFIG_PATH=/data/config.json \
  -e DS2API_ADMIN_KEY=你的强密码 \
  ghcr.io/ojbkxc/ds2api:latest
```

**从 Docker Hub 拉取**：
```bash
docker pull ojbkxc/ds2api:latest

docker run -d \
  --name ds2api \
  -p 6011:5001 \
  -v ./config.json:/data/config.json \
  -e DS2API_CONFIG_PATH=/data/config.json \
  -e DS2API_ADMIN_KEY=你的强密码 \
  ojbkxc/ds2api:latest
```


#### Zeabur 一键部署（Dockerfile�?
1. 点击上方 “Deploy on Zeabur�?按钮，一键部署�?2. 部署完成后访�?`/admin`，使�?Zeabur 环境变量/模板指引中的 `DS2API_ADMIN_KEY` 登录�?3. 在管理台导入/编辑配置（会写入并持久化�?`/data/config.json`）�?
Zeabur 首次空卷启动时可以没�?`/data/config.json`；DS2API 会先使用空的文件模式配置启动，并在管理台首次保存时创建该文件�?
不依赖模板手动部署时，在 Zeabur 中选择 GitHub 仓库服务，Root Directory 保持 `/`，使用仓库根目录 `Dockerfile` 构建；添加持久卷 `/data`，设�?`PORT=5001`、`DS2API_ADMIN_KEY=你的强密钥`、`DS2API_CONFIG_PATH=/data/config.json`，然后暴�?HTTP 端口 `5001`。更完整步骤�?[docs/DEPLOY.md](docs/DEPLOY.md#不使用模板手动部�?�?
说明：Zeabur 使用仓库�?`Dockerfile` 直接构建时，不需要额外传�?`BUILD_VERSION`；镜像会优先读取该构建参数，未提供时自动回退到仓库根目录�?`VERSION` 文件�?
### 方式三：Vercel 部署

1. Fork 仓库到自己的 GitHub
2. �?Vercel 上导入项�?3. 配置环境变量（最少设�?`DS2API_ADMIN_KEY`；推荐同时设�?`DS2API_CONFIG_JSON`�?4. 部署

建议先在仓库目录复制模板并填写：

```bash
cp config.example.json config.json
# 编辑 config.json
```

推荐：先本地�?`config.json` 转成 Base64，再粘贴�?`DS2API_CONFIG_JSON`，避�?JSON 格式错误�?
```bash
base64 < config.json | tr -d '\n'
```

> **流式说明**：OpenAI Chat 流式�?Vercel 上会�?`api/chat-stream.js`（Node Runtime）承接，�?`vercel.json` 只把规范路径 `/v1/chat/completions` 重写�?Node；根路径快捷别名 `/chat/completions` 仍走 Go 主链路。鉴权、账号选择、会�?PoW 准备仍由 Go 内部 prepare 接口完成；流式响应（�?`tools`）在 Node 侧执行与 Go 对齐的输出组装与防泄漏处理。Vercel 上需要实时流式时请使�?`/v1/chat/completions`�?
详细部署说明请参�?[部署指南](docs/DEPLOY.md)�?
### 方式四：本地源码运行

**前置要求**：Go 1.26+，Node.js `20.19+` �?`22.12+`（仅在需要构�?WebUI 时；CI / Docker 构建使用 Node 24）；同时确保 `npm` 可用，建�?`npm 10+`

```bash
# 1. 克隆仓库
git clone https://github.com/ojbkxc/ds2api.git
cd ds2api

# 2. 配置
cp config.example.json config.json
# 编辑 config.json，填入你�?DeepSeek 账号信息�?API key

# 3. 启动
go run ./cmd/ds2api
```

默认本地访问地址：`http://127.0.0.1:5001`

服务实际绑定：`0.0.0.0:5001`，因此同一局域网设备通常也可以通过你的内网 IP 访问�?
> **WebUI 自动构建**：本地首次启动时，若 WebUI 静态目录不存在，会自动尝试执行 `npm ci --prefix webui`（仅在缺少依赖时）和 `npm run build --prefix webui -- --outDir static/admin --emptyOutDir`（需要本机有 Node.js �?npm；静态目录可�?`DS2API_STATIC_ADMIN_DIR` 覆盖）。你也可以手动构建：`./scripts/build-webui.sh`

## 配置说明

`README` 只保留快速入口，完整字段请以 [config.example.json](config.example.json) 为模板，并参�?[部署指南](docs/DEPLOY.md#0-前置要求) �?[API 配置最佳实践](API.md#配置最佳实�?�?
常用字段�?
- `keys` / `api_keys`：客户端访问密钥，`api_keys` 支持 `name` �?`remark` 元信息，`keys` 继续兼容�?- `accounts`：DeepSeek 托管账号，支�?`email` �?`mobile` 登录，可配置代理、名称和备注�?- `model_aliases`：OpenAI / Claude / Gemini 共用的模�?alias 映射�?- `runtime`：账号并发、队列、token 刷新策略（含刷新抖动）、最大账号切换次数与每会话消息上限，可通过 Admin Settings 热更新�?- `auto_delete.mode`：请求结束后的远端会话清理策略，支持 `none` / `single` / `all`；`auto_delete.delay_hours` 控制延迟清理小时数（默认 72h）�?- `current_input_file`：全局生效的上下文拆分上传策略；默认开启，首条消息无历史时阈值提升至 3 倍（最低 2000 字符），触发时将完整上下文合并上传�?`deepseek.txt` 上下文文件�?- 如果关闭 `current_input_file`，请求会直接透传，不上传拆分上下文文件�?- `thinking_injection`：默认开启；在最�?user 消息末尾追加思考增强提示词，提高高强度推理与工具调用前的思考稳定性；`prompt` 留空时使用内置默认提示词�?
环境变量完整列表�?[部署指南](docs/DEPLOY.md)，接口鉴权规则见 [API.md](API.md#鉴权规则)�?
## 鉴权模式

调用业务接口（`/v1/*`、`/anthropic/*`、Gemini 路由）时支持两种模式�?
| 模式 | 说明 |
| --- | --- |
| **托管账号模式** | `Bearer` �?`x-api-key` 传入 `config.keys` 中的 key，由服务自动轮询选择账号 |
| **直�?token 模式** | 传入 token 不在 `config.keys` 中时，直接作�?DeepSeek token 使用 |

可选请求头 `X-Ds2-Target-Account`：指定使用某个托管账号（值为 email �?mobile）�?如果指定账号不存在，或者当前管理账号队列已满，请求会返�?`429`；当�?`429` 不附�?`Retry-After` 头。若账号存在但登�?刷新失败，则返回对应的鉴权错误�?未指定目标账号时，如�?completion 因上�?thinking-only 空输出在同账号补偿重试后仍将返回 `429 upstream_empty_output`，托管账号模式会自动切到下一个可用账号，新建 session，并用原�?payload �?fresh retry 一次�?Gemini 路由还可以使�?`x-goog-api-key`，或在没有认证头时使�?`?key=` / `?api_key=` 作为调用方凭据�?
## 并发模型

```
每账号可用并�?= DS2API_ACCOUNT_MAX_INFLIGHT（默�?2�?建议并发�?= 账号数量 × 每账号并发上�?等待队列上限 = DS2API_ACCOUNT_MAX_QUEUE（默�?= 建议并发值）
429 阈�?= in-flight + 等待队列 �?账号数量 × 4
```

- �?in-flight 槽位满时，请求进入等待队列，**不会立即 429**
- 超出总承载上限后才返�?`429 Too Many Requests`，当前响应不附带 `Retry-After`
- completion 空输出类 429 会先做同账号补偿重试；托管账号模式还会在最终返�?429 前切到另一个可用账�?fresh retry 一�?- `GET /admin/queue/status` 返回实时并发状�?
## Tool Call 适配

当请求中�?`tools` 时，DS2API 会做防泄漏处理与结构化转译：

1. 只在**�?Markdown 代码上下�?*启用执行�?toolcall 识别（fenced code block 和行�?code span 中的示例默认不触发）
2. 解析层当前把半角管道�?DSML 外壳视为推荐可执行调用：`<|DSML|tool_calls>` �?`<|DSML|invoke name="...">` �?`<|DSML|parameter name="...">`；兼容旧�?canonical XML `<tool_calls>` �?`<invoke name="...">` �?`<parameter name="...">`，以及若�?DSML 前缀/分隔符漂移。DSML 只是外壳别名，内部仍�?XML 解析语义为准；旧�?`<tools>` / `<tool_call>` / `<tool_name>` / `<param>`、`<function_call>`、`tool_use` / antml 变体与纯 JSON `tool_calls` 片段都会按普通文本处理，完整�?malformed �?wrapper 也会作为普通文本释�?3. `responses` 流式严格使用官方 item 生命周期事件（`response.output_item.*`、`response.content_part.*`、`response.function_call_arguments.*`�?4. `responses` 支持并执�?`tool_choice`（`auto`/`none`/`required`/强制函数）；`required` 违规时非流式返回 `422`，流式返�?`response.failed`
5. 客户端请求哪种协议，就按该协议返回工具调用（OpenAI/Claude/Gemini 各自原生结构）；模型侧优先约束输出规�?XML，再由兼容层转译

> 说明：当前版�?parser 层以”尽量解析成功”为优先，所有格式合法的 XML 工具调用都会通过，不做工具名 allow-list 过滤�?> 解析层会保留显式空字符串或纯空白参数；Prompt 会要求模型不要主动输出空参数，缺�?空命令的拒绝应由工具执行侧或客户�?schema 校验负责�?>
> 想评估”把工具调用封装�?XML 再输入模型”的方案，可参考：`docs/toolcall-semantics.md`�?
## 本地开发抓包工�?
用于定位「responses 思考流/工具调用」等问题。开启后会自动记录最�?N �?DeepSeek 对话上游请求体与响应体（默认 20 条，超出自动淘汰；单条响应体默认最多记�?5 MB）�?
启用示例�?
```bash
DS2API_DEV_PACKET_CAPTURE=true \
DS2API_DEV_PACKET_CAPTURE_LIMIT=20 \
go run ./cmd/ds2api
```

查询/清空（需 Admin JWT）：

- `GET /admin/dev/captures`：查看抓包列表（最新在前）
- `DELETE /admin/dev/captures`：清空抓�?- `GET /admin/dev/raw-samples/query?q=关键�?limit=20`：按问题关键词查询当前内存抓包，并按 `chat_session_id` 归并 `completion + continue` �?- `POST /admin/dev/raw-samples/save`：把命中的某条抓包链保存�?`tests/raw_stream_samples/<sample-id>/` 回放样本

返回字段包含�?
- `request_body`：发送给 DeepSeek 的完整请求体
- `response_body`：上游返回的原始流式内容拼接文本
- `response_truncated`：是否触发单条大小截�?
保存接口支持�?`query`、`chain_key` �?`capture_id` 选中目标。例如：

```json
{"query":"广州天气","sample_id":"gz-weather-from-memory"}
```

## 文档索引

| 文档 | 说明 |
| --- | --- |
| [API.md](API.md) / [API.en.md](API.en.md) | API 接口文档（含请求/响应示例�?|
| [DEPLOY.md](docs/DEPLOY.md) / [DEPLOY.en.md](docs/DEPLOY.en.md) | 部署指南（本�?Docker/Vercel/systemd�?|
| [CONTRIBUTING.md](docs/CONTRIBUTING.md) / [CONTRIBUTING.en.md](docs/CONTRIBUTING.en.md) | 贡献指南 |
| [TESTING.md](docs/TESTING.md) | 测试集使用指�?|

## 测试

详细测试指南请参�?[docs/TESTING.md](docs/TESTING.md)�?
### 快速测试命�?
```bash
# 本地 PR 门禁
./scripts/lint.sh
./tests/scripts/check-refactor-line-gate.sh
./tests/scripts/run-unit-all.sh
npm run build --prefix webui

# 端到端全链路测试（真实账号，生成完整请求/响应日志�?./tests/scripts/run-live.sh
```

## Release 自动构建（GitHub Actions�?
工作流文件：`.github/workflows/release-artifacts.yml`

- **触发条件**：默认仅�?GitHub Release `published` 时自动触发；也支持在 Actions 页面手动 `workflow_dispatch`，并填写 `release_tag` 复跑/补发
- **构建产物**：多平台二进制包（`linux/amd64`、`linux/arm64`、`linux/armv7`、`darwin/amd64`、`darwin/arm64`、`windows/amd64`、`windows/arm64`）、Linux Docker 镜像导出�?+ `sha256sums.txt`
- **容器镜像发布**：仅推送到 GHCR（`ghcr.io/ojbkxc/ds2api`�?- **每个二进制压缩包包含**：`ds2api` 可执行文件、`static/admin`、`config.example.json`、`.env.example`、`README.MD`、`README.en.md`、`LICENSE`

## 免责声明

本项目基于逆向方式实现，仅供学习、研究、个人实验和内部验证使用，不提供任何商业授权、稳定性保证或可用性保证�?作者及仓库维护者不对因使用、修改、分发、部署或依赖本项目而产生的任何直接或间接损失、账号封禁、数据丢失、法律风险或第三方索赔负责�?
请勿将本项目用于违反服务条款、协议、法律法规或平台规则的场景。商业使用前请自行确�?`LICENSE`、相关协议以及你是否获得了作者的书面许可�?
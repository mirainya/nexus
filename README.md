# Nexus

Nexus 是一个基于 LLM 的智能文档处理与知识图谱构建平台。通过可配置的 Pipeline 流水线，自动完成文档解析、实体提取、关系发现、向量化等任务，将非结构化数据转化为结构化的知识图谱。

## 核心特性

- **Pipeline 流水线引擎** — 可视化编排多步骤处理流程，支持条件执行、并行分组、嵌套子流水线、错误重试
- **多 LLM 服务商** — 支持 OpenAI、Anthropic、豆包，每个步骤可独立配置 provider 和 model，支持租户自带 LLM 凭证
- **12 个内置处理器** — OCR、人脸识别、图片评估、实体提取、LLM 审核、实体对齐、分类、摘要、向量化、上下文加载、条件路由、子流水线
- **知识图谱** — 自动提取实体和关系，支持图谱可视化浏览和语义搜索
- **混合搜索** — 向量语义搜索 + 关键词搜索 + LLM 重排序，三种模式可选
- **多租户隔离** — 租户级数据隔离，API Key 绑定租户，支持请求/Token 配额管理
- **双认证体系** — 管理控制台 JWT 认证 + 外部 API Key 认证，独立的权限和配额控制
- **人工审核** — 低置信度实体自动进入审核队列，支持批准/拒绝/修改三种操作
- **异步任务** — Redis + Asynq 任务队列，SSE 实时进度推送（支持 Redis Pub/Sub 多实例）
- **内容去重** — SHA256 内容哈希 + Redis 缓存，避免重复处理
- **Webhook 回调** — HMAC-SHA256 签名，SSRF 防护，指数退避重试
- **成本追踪** — 按 provider 配置 token 价格，自动计算每步骤费用
- **可观测性** — Pipeline 性能分析、LLM 调用追踪、错误监控仪表盘
- **管理控制台** — React 前端，响应式布局，流水线管理、提示词模板、任务监控、实体浏览、图谱可视化

## 技术栈

| 层级 | 技术 |
|------|------|
| 后端 | Go, Gin, GORM, Asynq |
| 前端 | React, TypeScript, TanStack Query, Tailwind CSS |
| 数据库 | PostgreSQL / SQLite |
| 缓存/队列 | Redis |
| 向量数据库 | Milvus（可选） |
| 图谱可视化 | Sigma.js |
| 迁移工具 | golang-migrate |

## 项目结构

```
nexus/
├── cmd/
│   ├── server/          # HTTP 服务入口
│   └── migrate/         # 数据库迁移工具
├── internal/
│   ├── api/             # HTTP 路由、Handler、中间件
│   │   ├── handler/     # 请求处理器
│   │   └── middleware/  # JWT认证、API Key认证、配额检查
│   ├── llm/             # LLM Gateway（OpenAI/Anthropic/Doubao）
│   ├── model/           # 数据模型（GORM）
│   ├── pipeline/        # Pipeline 引擎、处理器接口、条件求值
│   ├── processor/       # 12个处理器实现
│   ├── service/         # 业务逻辑层
│   ├── sse/             # Server-Sent Events（支持Redis Pub/Sub）
│   └── worker/          # Asynq 异步任务 Worker
├── pkg/                 # 公共包（config, crypto, database, logger, cache, vectordb, httputil）
├── console/             # React 前端（Vite + TypeScript + Tailwind）
├── database/migrations/ # SQL 迁移文件
├── docs/                # Swagger API 文档
└── configs/             # 配置文件
```

## 处理器一览

| 处理器 | 功能 | 输入 → 输出 |
|--------|------|-------------|
| `ocr` | 图片文字识别 | 图片 → 覆盖文档内容为识别文本 |
| `face` | 人脸/人物视觉分析 | 图片 → person_visual 实体 + 场景信息 |
| `image_assess` | 图片质量/场景评估 | 图片 → 评估结果（用于推荐） |
| `classifier` | 内容分类 | 文本 → 分类标签 |
| `llm_extract` | 实体关系提取 | 文本+视觉数据 → 实体 + 关系 + 摘要 |
| `llm_review` | LLM 二次审核校正 | 已提取实体关系 → 修正后的实体关系 |
| `context_loader` | 加载已有图谱实体 | DB查询 → 匹配的已有实体列表 |
| `entity_align` | 新旧实体对齐去重 | 新实体+已有实体 → 对齐后的实体 |
| `embedding` | 文本向量化 | 文本/摘要 → 向量（存入 Milvus） |
| `summarizer` | 文本摘要生成 | 文本 → 摘要 |
| `router` | 条件路由分支 | 条件表达式 → 执行匹配分支 |
| `sub_pipeline` | 嵌套子流水线 | pipeline_id → 执行另一个 Pipeline |

## 快速开始

### 环境要求

- Go 1.22+
- Node.js 18+
- PostgreSQL 15+
- Redis 7+
- Milvus 2.3+（可选，向量检索功能需要）

### 安装

```bash
git clone https://github.com/mirainya/nexus.git
cd nexus

# 复制配置文件并修改
cp configs/config.example.yaml configs/config.yaml
```

编辑 `configs/config.yaml`，填写数据库连接、Redis 地址、JWT Secret（至少 32 字符）等。

### 数据库迁移

```bash
export NEXUS_DATABASE_URL="postgres://nexus:nexus@localhost:5432/nexus?sslmode=disable"
go run cmd/migrate/main.go -cmd up
```

### 启动服务

```bash
# 编译并运行
go run cmd/server/main.go

# 或者编译后运行
go build -o nexus cmd/server/main.go
./nexus
```

服务默认监听 `:8080`，管理控制台访问 `http://localhost:8080`。

### 前端开发

```bash
cd console
npm install
npm run dev
```

## 配置 LLM 服务商

启动后通过管理控制台「设置」页面添加 LLM 服务商：

1. 填写服务商标识名（如 `openai`）、Base URL、API Key
2. 设置默认模型和 token 价格（用于成本追踪）
3. 在 Pipeline 步骤中可选择使用特定的服务商和模型

## API

### 外部 API（API Key 认证）

通过 `X-API-Key` Header 或 `api_key` Query 参数传递 API Key。每个 API Key 绑定一个租户，支持日/月请求配额和 Token 配额。

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/parse` | 同步解析文档（带 Redis 缓存） |
| POST | `/api/v1/jobs` | 提交异步任务（支持内容去重） |
| GET | `/api/v1/jobs/:id` | 查询任务状态和结果 |
| GET | `/api/v1/jobs/:id/events` | SSE 实时进度推送 |
| POST | `/api/v1/search` | 混合语义搜索 |
| GET | `/api/v1/entities` | 查询实体列表 |
| GET | `/api/v1/entities/:id` | 查询实体详情 |
| GET | `/api/v1/entities/:id/relations` | 查询实体关系 |
| GET | `/api/v1/graph` | 获取知识图谱数据 |
| POST | `/api/v1/recommend` | 场景推荐 |
| POST | `/api/v1/upload` | 文件上传 |

### 管理 API（JWT 认证）

通过 `POST /api/admin/login` 获取 JWT Token，在 `Authorization: Bearer <token>` 中传递。

管理 Pipeline、Prompt Template、Job、Entity、Review、LLM Provider、API Key、Credential、Tenant 等资源的完整 CRUD。详见 Swagger 文档（`/swagger/index.html`）。

## 核心流程

```
提交任务 → SHA256去重检查 → 创建Document+Job → Asynq入队
    → Worker执行Pipeline → 各处理器按序/并行执行
    → 实体持久化（≥0.8自动确认，<0.8进审核队列）
    → 向量入库 → Webhook回调 → SSE推送完成
```

## License

MIT

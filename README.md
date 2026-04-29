# Nexus

Nexus 是一个基于 LLM 的智能文档处理与知识图谱构建平台。通过可配置的 Pipeline 流水线，自动完成文档解析、实体提取、关系发现、向量化等任务，将非结构化数据转化为结构化的知识图谱。

## 核心特性

- **Pipeline 流水线引擎** — 可视化编排多步骤处理流程，支持条件执行、并行分组、错误重试
- **多 LLM 服务商** — 支持 OpenAI、Anthropic、豆包等，每个步骤可独立配置 provider 和 model
- **智能处理器** — 内置 OCR、人脸识别、图片评估、实体提取、实体对齐、分类、审核等处理器
- **知识图谱** — 自动提取实体和关系，构建可查询的知识图谱
- **向量检索** — 基于 Milvus 的文档向量化与语义搜索
- **异步任务** — 基于 Redis + Asynq 的任务队列，支持 SSE 实时进度推送
- **成本追踪** — 按 provider 配置 token 价格，自动计算每步骤费用
- **管理控制台** — React 前端，提供流水线管理、提示词模板、任务监控、实体浏览等功能

## 技术栈

| 层级 | 技术 |
|------|------|
| 后端 | Go, Gin, GORM, Asynq |
| 前端 | React, TypeScript, TanStack Query, Tailwind CSS |
| 数据库 | PostgreSQL |
| 缓存/队列 | Redis |
| 向量数据库 | Milvus |
| 迁移工具 | golang-migrate |

## 项目结构

```
nexus/
├── cmd/
│   ├── server/          # HTTP 服务入口
│   └── migrate/         # 数据库迁移工具
├── internal/
│   ├── api/             # HTTP 路由、Handler、中间件
│   ├── llm/             # LLM Gateway（OpenAI/Anthropic/Doubao）
│   ├── model/           # 数据模型（GORM）
│   ├── pipeline/        # Pipeline 引擎、处理器接口
│   ├── processor/       # 各类处理器实现
│   ├── service/         # 业务逻辑层
│   ├── sse/             # Server-Sent Events
│   └── worker/          # 异步任务 Worker
├── pkg/                 # 公共包（config, database, logger, storage, errors）
├── console/             # React 前端
├── database/migrations/ # SQL 迁移文件
└── configs/             # 配置文件
```

## 快速开始

### 环境要求

- Go 1.26+
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

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/v1/parse` | 同步解析文档 |
| POST | `/api/v1/jobs` | 提交异步任务 |
| GET | `/api/v1/jobs/:id` | 查询任务状态 |
| GET | `/api/v1/jobs/:id/events` | SSE 实时进度 |
| POST | `/api/v1/search` | 语义搜索 |

### 管理 API（JWT 认证）

Pipeline、Prompt、Job、Entity、Review、LLM Provider 等资源的完整 CRUD，详见 `internal/api/router.go`。

## License

MIT

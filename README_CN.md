# TokenGate

TokenGate 是一个面向订阅制 AI 账号/API 资源的网关产品。它把上游 AI 产品订阅、OAuth 账号和 Provider API Key 统一包装成可管理、可计费、可分发的 API 服务，适合先自用，后续逐步开放给公众用户。

当前仓库按线上拆分部署设计：

| 层级 | 平台 | 子目录 |
| --- | --- | --- |
| 后端 API | Railway | 仓库根目录 |
| 前端 Dashboard | Vercel | `frontend` |
| 数据库 | Railway Postgres | 托管服务 |
| 缓存 | Railway Redis | 托管服务 |

TokenGate 基于开源 Sub2API 能力继续产品化，会作为独立服务持续演进。

## 核心能力

- 提供 OpenAI-compatible 和 Anthropic-compatible 网关端点。
- 管理 OpenAI、Claude 等上游账号/API Key。
- 给用户生成 TokenGate API Key，供下游业务调用。
- 按模型和 token 统计用量、成本和余额。
- 支持用户、套餐、余额、支付、管理员后台等 SaaS 能力。
- 内置上线前检查、生产 smoke test 和运维 runbook。

## 文档入口

- [快速开始](docs/TOKENGATE_QUICKSTART.md)
- [计费模型](docs/TOKENGATE_BILLING_MODEL.md)
- [部署检查清单](docs/TOKENGATE_DEPLOYMENT_CHECKLIST.md)
- [运维 Runbook](docs/TOKENGATE_OPERATIONS_RUNBOOK.md)
- [上线路线图](docs/TOKENGATE_LAUNCH_ROADMAP.md)
- [产品策略](docs/TOKENGATE_PRODUCT_STRATEGY.md)

## 部署方式

### 后端部署到 Railway

Railway 使用仓库根目录部署，构建根目录的 `Dockerfile`。

Railway 需要的组件：

- TokenGate backend service
- PostgreSQL service
- Redis service

后端核心环境变量：

```bash
DATABASE_URL="${{Postgres.DATABASE_URL}}"
REDIS_URL="${{Redis.REDIS_URL}}"
JWT_SECRET="replace_with_openssl_rand_hex_32"
TOTP_ENCRYPTION_KEY="replace_with_openssl_rand_hex_32"
RUN_MODE="standard"
SERVER_MODE="release"
LOG_SERVICE_NAME="tokengate"
LOG_ENV="production"
FRONTEND_URL="https://your-frontend-domain"
CORS_ALLOWED_ORIGINS="https://your-frontend-domain"
```

生产密钥生成方式：

```bash
openssl rand -hex 32
```

### 前端部署到 Vercel

Vercel 使用 `frontend` 子目录部署。

前端必须配置：

```bash
VITE_API_BASE_URL="https://your-railway-backend-domain"
```

不要把 `VITE_API_BASE_URL` 指向 Vercel 前端域名，它必须指向 Railway 后端域名。

## 验证方式

部署前或修改环境变量后，先跑环境检查：

```bash
TOKENGATE_BACKEND_URL="https://your-railway-backend-domain" \
TOKENGATE_FRONTEND_URL="https://your-frontend-domain" \
DATABASE_URL="postgresql://example" \
REDIS_URL="redis://example" \
JWT_SECRET="$(openssl rand -hex 32)" \
TOTP_ENCRYPTION_KEY="$(openssl rand -hex 32)" \
VITE_API_BASE_URL="https://your-railway-backend-domain" \
tools/check_tokengate_env.sh
```

创建 TokenGate API Key 后，可以跑真实网关 smoke test：

```bash
TOKENGATE_BASE_URL="https://your-railway-backend-domain" \
TOKENGATE_API_KEY="tg_live_or_sub2api_key" \
TOKENGATE_RUN_CLAUDE=1 \
TOKENGATE_RUN_OPENAI=1 \
tools/tokengate_smoke_test.sh
```

## 本地开发

后端：

```bash
cd backend
go test ./...
```

前端：

```bash
cd frontend
npm install
npm run build:standalone
```

## 当前状态

TokenGate 已经可以进入私有生产验证阶段。下一阶段重点是支付测试模式、SMTP/密码重置、备份恢复演练、正式套餐设计，以及完整的公众用户 onboarding 文档。

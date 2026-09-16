# MPC 钱包系统交付物

本目录是论文配套的可运行演示交付物。它实现 RESTful 服务、会话审批绑定、7 张 MySQL 表定义、Docker 部署文件、分层接口测试、3节点DKG/2节点门限签名集成测试和 100 并发会话创建压力测试。

## 运行

```bash
go test ./... -count=1 -v
go run . httpServer
```

启动后访问 `http://127.0.0.1:8080` 可打开随服务提供的管理后台。前端位于 `web/`，无需单独安装 Node.js 依赖；可以创建签名会话，并以两个不同节点完成审批，随后在交易与审计页面查看联动结果。

### MySQL and Redis runtime

复制 `.env.example` 并设置 `MYSQL_DSN` 后，启动入口会在监听 HTTP 前执行 `databases/migrations` 下的版本化迁移。设置 `REDIS_ADDR` 后，签名会话的创建、审批和取消事件会写入 Redis Stream，`consumers/` 中的消费者负责异步消费与确认。

命令结构与参考项目保持一致：

```bash
go run . httpServer       # HTTP API 与后台页面
go run . startConsumers   # Redis Stream 消费者
go run . cron             # 定时任务进程
```

完整依赖环境可通过以下命令启动：

```bash
docker compose -f deploy/docker-compose.yml up --build
```

## 项目结构

- `routes/`：Gin 路由注册与 `/api/v1` 业务分组。
- `app/tss_wallet/api/controllers/`：请求绑定和统一响应出口。
- `app/tss_wallet/api/services/`：钱包、交易、签名会话、节点与审计的业务状态机。
- `app/tss_wallet/api/requests/`：接口请求对象与参数校验规则。
- `models/`：钱包、参与方、交易、会话、密钥版本和审计记录模型。
- `common/`：与管理项目一致的 `code`、`message`、`data` 响应结构。

## API 概览

| 路由 | 方法 | 作用 |
|---|---|---|
| `/api/v1/health` | GET | 健康检查与门限参数 |
| `/api/v1/wallets` | GET/POST | 查询或创建钱包元数据 |
| `/api/v1/wallets/:id/addresses` | GET | 查询地址 |
| `/api/v1/wallets/:id/balance` | GET | 查询演示余额 |
| `/api/v1/transactions` | GET/POST | 查询或发起转账 |
| `/api/v1/transactions/:id` | GET | 查询交易详情 |
| `/api/v1/sign-sessions` | GET/POST | 查询或创建签名会话 |
| `/api/v1/sign-sessions/:id/approve` | POST | 节点审批会话 |
| `/api/v1/sign-sessions/:id/cancel` | POST | 取消会话 |
| `/api/v1/participants` | GET | 查询参与方节点 |
| `/api/v1/participants/:id/refresh` | POST | 发起重新共享请求 |
| `/api/v1/nodes/heartbeat` | POST | 上报节点心跳 |
| `/api/v1/audit-logs` | GET | 查询审计日志 |
| `/api/v1/system/metrics` | GET | 查询系统指标 |

## 门限参数与协议边界

`go.mod` 固定使用 `github.com/bnb-chain/tss-lib v1.5.0`。`internal/tssconfig/config.go` 将 3 个参与方和库参数 `threshold=1` 映射为 2-of-3 策略：该参数为多项式阶数，所需协作参与方数为 `threshold + 1`。

库路径为 `ecdsa/keygen`、`ecdsa/signing` 和 `ecdsa/resharing`。门限签名按库的多轮 `LocalParty` 消息状态机执行，协调器只转发经认证的协议消息并收集最终输出，不能将两个独立 ECDSA 签名简单相加。`resharing` 用于重新分配参与方分片；本演示不将其表述为独立的 proactive refresh 接口。

## 交付结构

- `main.go`：进程入口，加载 `cmd` 注册的命令
- `cmd/http_server.go`：HTTP API 与管理后台进程
- `cmd/start_consumer.go`：Redis Stream 消费者进程
- `cmd/cron.go`：定时任务进程
- `internal/tssconfig`：`tss.NewParameters` 的 2-of-3 参数核验
- `databases/migrations`：版本化 MySQL 表迁移
- `core/httpserver`、`core/sql`：HTTP 与 SQL 连接池基础设施
- `tools/redis`：Redis 客户端和 Stream 事件发布器
- `consumers`：Redis Stream 会话事件消费者
- `deploy/docker-compose.yml` 与 `Dockerfile`：部署入口
- `测试报告.md`：20 条可追溯用例和并发测试范围

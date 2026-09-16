# MPC 钱包系统交付物

本目录是论文配套的可运行演示交付物。它实现 RESTful 服务、会话审批绑定、7 张 MySQL 表定义、Docker 部署文件、20 项以上自动化功能测试、3节点DKG/2节点门限签名集成测试和 100 并发会话创建压力测试。

## 运行

```bash
go test ./... -count=1 -v
go run ./cmd/server
```

启动后访问 `http://127.0.0.1:8080` 可打开随服务提供的管理后台。前端位于 `web/`，无需单独安装 Node.js 依赖；可以创建签名会话，并以两个不同节点完成审批，随后在交易与审计页面查看联动结果。

## 门限参数与协议边界

`go.mod` 固定使用 `github.com/bnb-chain/tss-lib v1.5.0`。`internal/tssconfig/config.go` 将 3 个参与方和库参数 `threshold=1` 映射为 2-of-3 策略：该参数为多项式阶数，所需协作参与方数为 `threshold + 1`。

库路径为 `ecdsa/keygen`、`ecdsa/signing` 和 `ecdsa/resharing`。门限签名按库的多轮 `LocalParty` 消息状态机执行，协调器只转发经认证的协议消息并收集最终输出，不能将两个独立 ECDSA 签名简单相加。`resharing` 用于重新分配参与方分片；本演示不将其表述为独立的 proactive refresh 接口。

## 交付结构

- `cmd/server`：REST API 和审批状态机
- `internal/tssconfig`：`tss.NewParameters` 的 2-of-3 参数核验
- `deploy/schema.sql`：7 张数据表
- `deploy/docker-compose.yml` 与 `Dockerfile`：部署入口
- `测试报告.md`：20 条可追溯用例和并发测试范围

# 本地测试命令

本文档用于启动本地测试版钱包服务、测试 TSS 节点和管理后台。

## 环境要求

- Go 1.27 或更高版本
- Docker Desktop（仅使用 Docker 方式时需要）
- Redis 和 MySQL 不是单机内存测试的必需项

进入项目目录：

```bash
cd /Users/admin/code/tss-otc
```

首次执行脚本前运行：

```bash
chmod +x ./scripts/local-test.sh
```

## 运行全部测试

```bash
./scripts/local-test.sh test
```

## 本地启动方式

前端是后端内置的静态页面，不需要单独运行 Node.js 或 Vite。

请打开四个终端窗口，分别执行：

终端 1：

```bash
./scripts/local-test.sh tss1
```

终端 2：

```bash
./scripts/local-test.sh tss2
```

终端 3：

```bash
./scripts/local-test.sh tss3
```

终端 4：

```bash
./scripts/local-test.sh backend
```

浏览器打开：

```text
http://127.0.0.1:8080
```

## 检查服务

```bash
curl http://127.0.0.1:9441/health
curl http://127.0.0.1:9442/health
curl http://127.0.0.1:9443/health
curl http://127.0.0.1:8080/api/v1/health
```

## Redis 消费者和定时任务

本机已启动 Redis 时，可分别在两个终端执行：

```bash
./scripts/local-test.sh consumers
```

```bash
./scripts/local-test.sh cron
```

## Docker 启动方式

一次启动 MySQL、Redis、后端和消费者：

```bash
./scripts/local-test.sh docker
```

停止 Docker 服务：

```bash
docker compose -f deployments/docker/docker-compose.yml down
```

## 直接使用 Go 命令

```bash
SIGNER_MODE=local go run ./apps/tss-wallet-service/cmd httpServer
```

```bash
go run ./apps/tss-wallet-service/cmd tssNode --node-id node-1 --addr 127.0.0.1:9441
```

```bash
go run ./apps/tss-wallet-service/cmd tssNode --node-id node-2 --addr 127.0.0.1:9442
```

```bash
go run ./apps/tss-wallet-service/cmd tssNode --node-id node-3 --addr 127.0.0.1:9443
```

```bash
go run ./apps/tss-wallet-service/cmd startConsumers
go run ./apps/tss-wallet-service/cmd cron
```

停止服务时在对应终端按 `Ctrl+C`。

测试版使用 `SIGNER_MODE=local`，密钥只存在当前后端进程内，适合本地联调；生产环境不能使用该模式。

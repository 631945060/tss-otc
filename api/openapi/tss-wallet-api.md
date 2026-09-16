# TSS Wallet API

所有业务接口都使用统一结构：

```json
{
  "code": 0,
  "message": "success",
  "data": {}
}
```

`code=0` 表示成功；`1001` 表示参数错误；`1004` 表示资源不存在；`1009` 表示状态冲突或重复操作。路由按钱包、交易、会话、参与方和系统模块分组，对应论文的 4.4 节。

## Transaction and signing flow

1. `POST /api/v1/transactions` 创建待审批交易。
2. `POST /api/v1/sign-sessions` 将交易 ID、钱包 ID 和摘要绑定到 2-of-3 会话。
3. 两个不同节点调用 `POST /api/v1/sign-sessions/:id/approve`。
4. 会话状态变为 `success`，关联交易状态变为 `signed`，并写入审计日志。

审批接口维护的是业务层门限授权状态。实际 tss-lib 的 DKG、两方协作签名与 ECDSA 验签路径由 `internal/tssprotocol` 自动化测试验证，协调器接口不保存私钥分片，也不把两份独立签名相加。

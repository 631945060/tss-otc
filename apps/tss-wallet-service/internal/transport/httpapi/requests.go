package httpapi

type CreateWalletReq struct {
	Network string `json:"network" binding:"omitempty,oneof=testnet mainnet"`
}

type CreateTransactionReq struct {
	WalletID  string `json:"wallet_id" binding:"required,max=96"`
	ToAddress string `json:"to_address" binding:"required,max=160"`
	Amount    string `json:"amount" binding:"required,max=64"`
	Fee       string `json:"fee" binding:"omitempty,max=64"`
	Digest    string `json:"digest" binding:"required,max=160"`
}

type CreateSignSessionReq struct {
	WalletID      string `json:"wallet_id" binding:"required,max=96"`
	TransactionID string `json:"transaction_id" binding:"omitempty,max=96"`
	Digest        string `json:"digest" binding:"required,max=160"`
}

type ApproveSessionReq struct {
	NodeID string `json:"node_id" binding:"required,oneof=node-1 node-2 node-3"`
}

type HeartbeatReq struct {
	NodeID  string `json:"node_id" binding:"required,oneof=node-1 node-2 node-3"`
	Version string `json:"version" binding:"required,max=32"`
}

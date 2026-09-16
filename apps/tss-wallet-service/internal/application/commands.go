package application

type CreateWalletInput struct{ Network string }

type CreateTransactionInput struct {
	WalletID  string
	ToAddress string
	Amount    string
	Fee       string
	Digest    string
}

type CreateSignSessionInput struct {
	WalletID      string
	TransactionID string
	Digest        string
}

type HeartbeatInput struct {
	NodeID  string
	Version string
}

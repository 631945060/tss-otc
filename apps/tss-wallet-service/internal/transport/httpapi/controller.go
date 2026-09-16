package httpapi

import (
	"errors"

	"github.com/gin-gonic/gin"

	"mpc-wallet-demo/apps/tss-wallet-service/internal/application"
)

type TSSWalletController struct {
	BaseController
	service *application.TSSWalletService
}

func NewTSSWalletController(service *application.TSSWalletService) *TSSWalletController {
	return &TSSWalletController{service: service}
}

func (c *TSSWalletController) Health(ctx *gin.Context) {
	c.ResponseSuccess(ctx, gin.H{"status": "ok", "protocol": "tss-lib v1.5.0", "threshold": "2-of-3 (library threshold=1)"})
}

func (c *TSSWalletController) WalletList(ctx *gin.Context) {
	c.ResponseSuccess(ctx, gin.H{"list": c.service.ListWallets()})
}

func (c *TSSWalletController) WalletCreate(ctx *gin.Context) {
	var req CreateWalletReq
	if !c.bind(ctx, &req) {
		return
	}
	c.ResponseSuccess(ctx, c.service.CreateWallet(application.CreateWalletInput{Network: req.Network}), "wallet created")
}

func (c *TSSWalletController) WalletDetail(ctx *gin.Context) {
	data, err := c.service.Wallet(ctx.Param("id"))
	if c.respondError(ctx, err) {
		return
	}
	c.ResponseSuccess(ctx, data)
}

func (c *TSSWalletController) WalletAddresses(ctx *gin.Context) {
	data, err := c.service.WalletAddresses(ctx.Param("id"))
	if c.respondError(ctx, err) {
		return
	}
	c.ResponseSuccess(ctx, data)
}

func (c *TSSWalletController) WalletBalance(ctx *gin.Context) {
	data, err := c.service.WalletBalance(ctx.Param("id"))
	if c.respondError(ctx, err) {
		return
	}
	c.ResponseSuccess(ctx, data)
}

func (c *TSSWalletController) TransactionList(ctx *gin.Context) {
	c.ResponseSuccess(ctx, gin.H{"list": c.service.ListTransactions()})
}

func (c *TSSWalletController) TransactionCreate(ctx *gin.Context) {
	var req CreateTransactionReq
	if !c.bind(ctx, &req) {
		return
	}
	data, err := c.service.CreateTransaction(application.CreateTransactionInput{WalletID: req.WalletID, ToAddress: req.ToAddress, Amount: req.Amount, Fee: req.Fee, Digest: req.Digest})
	if c.respondError(ctx, err) {
		return
	}
	c.ResponseSuccess(ctx, data, "transaction created")
}

func (c *TSSWalletController) TransactionDetail(ctx *gin.Context) {
	data, err := c.service.Transaction(ctx.Param("id"))
	if c.respondError(ctx, err) {
		return
	}
	c.ResponseSuccess(ctx, data)
}

func (c *TSSWalletController) SessionList(ctx *gin.Context) {
	c.ResponseSuccess(ctx, gin.H{"list": c.service.ListSessions()})
}

func (c *TSSWalletController) SessionCreate(ctx *gin.Context) {
	var req CreateSignSessionReq
	if !c.bind(ctx, &req) {
		return
	}
	data, err := c.service.CreateSession(application.CreateSignSessionInput{WalletID: req.WalletID, TransactionID: req.TransactionID, Digest: req.Digest})
	if c.respondError(ctx, err) {
		return
	}
	c.ResponseSuccess(ctx, data, "signing session created")
}

func (c *TSSWalletController) SessionDetail(ctx *gin.Context) {
	data, err := c.service.Session(ctx.Param("id"))
	if c.respondError(ctx, err) {
		return
	}
	c.ResponseSuccess(ctx, data)
}

func (c *TSSWalletController) SessionApprove(ctx *gin.Context) {
	var req ApproveSessionReq
	if !c.bind(ctx, &req) {
		return
	}
	data, err := c.service.ApproveSession(ctx.Param("id"), req.NodeID)
	if c.respondError(ctx, err) {
		return
	}
	c.ResponseSuccess(ctx, data, "approval recorded")
}

func (c *TSSWalletController) SessionCancel(ctx *gin.Context) {
	data, err := c.service.CancelSession(ctx.Param("id"))
	if c.respondError(ctx, err) {
		return
	}
	c.ResponseSuccess(ctx, data, "session cancelled")
}

func (c *TSSWalletController) ParticipantList(ctx *gin.Context) {
	c.ResponseSuccess(ctx, gin.H{"list": c.service.ListParticipants()})
}

func (c *TSSWalletController) ParticipantRefresh(ctx *gin.Context) {
	data, err := c.service.RefreshParticipant(ctx.Param("id"))
	if c.respondError(ctx, err) {
		return
	}
	c.ResponseSuccess(ctx, data, "resharing requested")
}

func (c *TSSWalletController) NodeHeartbeat(ctx *gin.Context) {
	var req HeartbeatReq
	if !c.bind(ctx, &req) {
		return
	}
	data, err := c.service.Heartbeat(application.HeartbeatInput{NodeID: req.NodeID, Version: req.Version})
	if c.respondError(ctx, err) {
		return
	}
	c.ResponseSuccess(ctx, data, "heartbeat accepted")
}

func (c *TSSWalletController) AuditLogList(ctx *gin.Context) {
	c.ResponseSuccess(ctx, gin.H{"list": c.service.ListAuditLogs()})
}
func (c *TSSWalletController) Metrics(ctx *gin.Context) { c.ResponseSuccess(ctx, c.service.Metrics()) }

func (c *TSSWalletController) bind(ctx *gin.Context, req any) bool {
	if err := ctx.ShouldBindJSON(req); err != nil {
		c.ResponseError(ctx, InvalidParamCode, err.Error())
		return false
	}
	return true
}

func (c *TSSWalletController) respondError(ctx *gin.Context, err error) bool {
	if err == nil {
		return false
	}
	switch {
	case errors.Is(err, application.ErrNotFound):
		c.ResponseError(ctx, NotFoundCode, err.Error())
	case errors.Is(err, application.ErrConflict), errors.Is(err, application.ErrInvalidState):
		c.ResponseError(ctx, ConflictCode, err.Error())
	default:
		c.ResponseError(ctx, InternalErrorCode, "internal server error")
	}
	return true
}

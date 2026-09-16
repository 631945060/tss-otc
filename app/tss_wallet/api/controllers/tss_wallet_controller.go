package controllers

import (
	"errors"

	"github.com/gin-gonic/gin"

	"mpc-wallet-demo/app/tss_wallet/api/requests"
	"mpc-wallet-demo/app/tss_wallet/api/services"
	"mpc-wallet-demo/common"
)

type TSSWalletController struct {
	common.BaseController
	service *services.TSSWalletService
}

func NewTSSWalletController(service *services.TSSWalletService) *TSSWalletController {
	return &TSSWalletController{service: service}
}

func (c *TSSWalletController) Health(ctx *gin.Context) {
	c.ResponseSuccess(ctx, gin.H{"status": "ok", "protocol": "tss-lib v1.5.0", "threshold": "2-of-3 (library threshold=1)"})
}

func (c *TSSWalletController) WalletList(ctx *gin.Context) {
	c.ResponseSuccess(ctx, gin.H{"list": c.service.ListWallets()})
}

func (c *TSSWalletController) WalletCreate(ctx *gin.Context) {
	var req requests.CreateWalletReq
	if !c.bind(ctx, &req) {
		return
	}
	c.ResponseSuccess(ctx, c.service.CreateWallet(req), "wallet created")
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
	var req requests.CreateTransactionReq
	if !c.bind(ctx, &req) {
		return
	}
	data, err := c.service.CreateTransaction(req)
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
	var req requests.CreateSignSessionReq
	if !c.bind(ctx, &req) {
		return
	}
	data, err := c.service.CreateSession(req)
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
	var req requests.ApproveSessionReq
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
	var req requests.HeartbeatReq
	if !c.bind(ctx, &req) {
		return
	}
	data, err := c.service.Heartbeat(req)
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
		c.ResponseError(ctx, common.InvalidParamCode, err.Error())
		return false
	}
	return true
}

func (c *TSSWalletController) respondError(ctx *gin.Context, err error) bool {
	if err == nil {
		return false
	}
	switch {
	case errors.Is(err, services.ErrNotFound):
		c.ResponseError(ctx, common.NotFoundCode, err.Error())
	case errors.Is(err, services.ErrConflict), errors.Is(err, services.ErrInvalidState):
		c.ResponseError(ctx, common.ConflictCode, err.Error())
	default:
		c.ResponseError(ctx, common.InternalErrorCode, "internal server error")
	}
	return true
}

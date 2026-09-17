package httpapi

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"mpc-wallet-demo/apps/tss-wallet-service/internal/application"
)

// NewRouter keeps controller construction and route grouping in one location,
// matching the route-registration style of the supplied reference project.
func NewRouter(staticDir string, servicesOverride ...*application.TSSWalletService) *gin.Engine {
	engine := gin.New()
	engine.Use(gin.Logger(), gin.Recovery(), cors())
	tssService := application.NewTSSWalletService()
	if len(servicesOverride) > 0 && servicesOverride[0] != nil {
		tssService = servicesOverride[0]
	}
	controller := NewTSSWalletController(tssService)

	api := engine.Group("/api/v1")
	{
		api.GET("/health", controller.Health)
		api.GET("/wallets", controller.WalletList)
		api.POST("/wallets", controller.WalletCreate)
		api.GET("/wallets/:id", controller.WalletDetail)
		api.GET("/wallets/:id/addresses", controller.WalletAddresses)
		api.GET("/wallets/:id/balance", controller.WalletBalance)

		transactionGroup := api.Group("/transactions")
		{
			transactionGroup.GET("", controller.TransactionList)
			transactionGroup.POST("", controller.TransactionCreate)
			transactionGroup.GET("/:id", controller.TransactionDetail)
		}
		sessionGroup := api.Group("/sign-sessions")
		{
			sessionGroup.GET("", controller.SessionList)
			sessionGroup.POST("", controller.SessionCreate)
			sessionGroup.GET("/:id", controller.SessionDetail)
			sessionGroup.POST("/:id/approve", controller.SessionApprove)
			sessionGroup.POST("/:id/cancel", controller.SessionCancel)
		}
		participantGroup := api.Group("/participants")
		{
			participantGroup.GET("", controller.ParticipantList)
			participantGroup.POST("/:id/refresh", controller.ParticipantRefresh)
		}
		api.GET("/key-epochs", controller.KeyEpochList)
		api.GET("/key-epochs/:id", controller.KeyEpochDetail)
		api.GET("/audit-logs", controller.AuditLogList)
		api.POST("/nodes/heartbeat", controller.NodeHeartbeat)
		api.GET("/system/metrics", controller.Metrics)
	}
	// Gin cannot register a root catch-all static route beside /api. Keeping the
	// file server in NoRoute preserves the API tree while serving the frontend.
	engine.NoRoute(gin.WrapH(http.FileServer(http.Dir(staticDir))))
	return engine
}

func cors() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Header("Access-Control-Allow-Origin", "*")
		ctx.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		ctx.Header("Access-Control-Allow-Headers", "Content-Type")
		if ctx.Request.Method == http.MethodOptions {
			ctx.Status(http.StatusNoContent)
			ctx.Abort()
			return
		}
		ctx.Next()
	}
}

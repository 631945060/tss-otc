package httpapi

import "github.com/gin-gonic/gin"

const (
	SuccessCode       = 0
	InvalidParamCode  = 1001
	NotFoundCode      = 1004
	ConflictCode      = 1009
	InternalErrorCode = 1500
)

type Response[T any] struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data"`
}

// BaseController matches the response responsibility of the reference project.
type BaseController struct{}

func (BaseController) ResponseSuccess(c *gin.Context, data any, message ...string) {
	msg := "success"
	if len(message) > 0 {
		msg = message[0]
	}
	c.JSON(200, Response[any]{Code: SuccessCode, Message: msg, Data: data})
}

func (BaseController) ResponseError(c *gin.Context, code int, message string) {
	c.JSON(200, Response[any]{Code: code, Message: message, Data: nil})
}

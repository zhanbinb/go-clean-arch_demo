// Package response defines the unified HTTP response envelope and Gin helpers.
package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/zhanbinb/go-clean-arch_demo/pkg/errcode"
	"github.com/zhanbinb/go-clean-arch_demo/pkg/logger"
)

// Resp is the standard JSON envelope returned by every API endpoint.
type Resp struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// OK writes a 200 response with the given payload.
func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Resp{
		Code:    0,
		Message: "OK",
		Data:    data,
	})
}

// Created writes a 201 response with the given payload.
func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, Resp{
		Code:    0,
		Message: "Created",
		Data:    data,
	})
}

// NoContent writes a 204 response.
func NoContent(c *gin.Context) { c.Status(http.StatusNoContent) }

// Error writes an error response. The HTTP status is derived from the
// errcode; the underlying cause (if any) is logged but never sent to client.
func Error(c *gin.Context, e *errcode.Error) {
	if e == nil {
		e = errcode.ErrInternal
	}
	if log := loggerFrom(c); log != nil && e.Unwrap() != nil {
		log.Warn("api error", zap.Int("code", int(e.Code)), zap.Error(e.Unwrap()))
	}
	c.AbortWithStatusJSON(e.HTTPStatus(), Resp{
		Code:    int(e.Code),
		Message: e.Message,
	})
}

// loggerFrom extracts a *logger.Logger from gin context (set by middleware).
func loggerFrom(c *gin.Context) *logger.Logger {
	if v, ok := c.Get("logger"); ok {
		if l, ok := v.(*logger.Logger); ok {
			return l
		}
	}
	return nil
}

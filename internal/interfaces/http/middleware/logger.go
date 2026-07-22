package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/zhanbinb/go-clean-arch_demo/pkg/logger"
)

// CtxKeyLogger is the gin.Context key for the request-scoped logger.
const CtxKeyLogger = "logger"

// Logger returns a gin middleware that logs each request via Zap.
func Logger(base *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		rid, _ := c.Get(CtxKeyRequestID)
		ridStr, _ := rid.(string)

		// Build request-scoped logger with request_id baked in
		reqLog := base.With(zap.String("request_id", ridStr))
		c.Set(CtxKeyLogger, reqLog)

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		fields := []zap.Field{
			zap.Int("status", status),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("query", c.Request.URL.RawQuery),
			zap.String("client_ip", c.ClientIP()),
			zap.Duration("latency", latency),
			zap.Int("size", c.Writer.Size()),
		}
		switch {
		case status >= 500:
			reqLog.Error("http request", fields...)
		case status >= 400:
			reqLog.Warn("http request", fields...)
		default:
			reqLog.Info("http request", fields...)
		}
	}
}

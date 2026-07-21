package middleware

import (
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/zhanbinb/go-clean-arch_demo/pkg/errcode"
	"github.com/zhanbinb/go-clean-arch_demo/pkg/logger"
	"github.com/zhanbinb/go-clean-arch_demo/pkg/response"
)

// Recovery catches panics and returns a 500 response instead of crashing the server.
func Recovery(log *logger.Logger) gin.HandlerFunc {
	zlog := log.Zap()
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				zlog.Error("panic recovered",
					zap.Any("panic", r),
					zap.String("path", c.Request.URL.Path),
					zap.String("stack", string(debug.Stack())),
				)
				response.Error(c, errcode.ErrInternal)
			}
		}()
		c.Next()
	}
}


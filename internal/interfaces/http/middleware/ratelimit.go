package middleware

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/zhanbinb/go-clean-arch_demo/pkg/errcode"
	"github.com/zhanbinb/go-clean-arch_demo/pkg/ratelimit"
	"github.com/zhanbinb/go-clean-arch_demo/pkg/response"
)

// RateLimit denies requests exceeding the per-key token bucket.
// Dimension is determined by the limiter config ("ip" or "user").
func RateLimit(l *ratelimit.Limiter, dimension string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var key string
		switch dimension {
		case "user":
			if uid, ok := c.Get(CtxKeyUserID); ok {
				key = fmt.Sprintf("user:%v", uid)
			} else {
				key = "ip:" + c.ClientIP() // fallback before JWT
			}
		default:
			key = "ip:" + c.ClientIP()
		}
		if !l.Allow(key) {
			c.Header("Retry-After", strconv.Itoa(1))
			response.Error(c, errcode.ErrTooManyRequests)
			c.Abort()
			return
		}
		c.Next()
	}
}

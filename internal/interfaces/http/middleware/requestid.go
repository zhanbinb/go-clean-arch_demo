package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/zhanbinb/go-clean-arch_demo/pkg/logger"
)

// HeaderXRequestID is the request/response header carrying the correlation id.
const HeaderXRequestID = "X-Request-ID"

// CtxKeyRequestID is the gin.Context key for the request id.
const CtxKeyRequestID = "request_id"

// RequestID extracts X-Request-ID from incoming requests (or generates a UUID),
// stores it in gin.Context and the request's context.Context, and echoes it
// back to the client.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader(HeaderXRequestID)
		if rid == "" {
			rid = uuid.NewString()
		}
		c.Set(CtxKeyRequestID, rid)
		c.Writer.Header().Set(HeaderXRequestID, rid)

		ctx := logger.WithRequestID(c.Request.Context(), rid)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

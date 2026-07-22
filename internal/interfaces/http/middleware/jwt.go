package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/zhanbinb/go-clean-arch_demo/pkg/errcode"
	"github.com/zhanbinb/go-clean-arch_demo/pkg/jwt"
	"github.com/zhanbinb/go-clean-arch_demo/pkg/logger"
	"github.com/zhanbinb/go-clean-arch_demo/pkg/response"
)

// CtxKeyUserID is the gin.Context key for the authenticated user id.
const CtxKeyUserID = "user_id"

// CtxKeyUsername is the gin.Context key for the authenticated username.
const CtxKeyUsername = "username"

// JWT validates the Authorization: Bearer <token> header.
func JWT(mgr *jwt.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw := c.GetHeader("Authorization")
		if raw == "" {
			response.Error(c, errcode.ErrUnauthorized)
			c.Abort()
			return
		}
		const prefix = "Bearer "
		if !strings.HasPrefix(raw, prefix) {
			response.Error(c, errcode.ErrTokenInvalid.WithMsg("expected Bearer scheme"))
			c.Abort()
			return
		}
		token := strings.TrimPrefix(raw, prefix)

		claims, err := mgr.Parse(token)
		if err != nil {
			response.Error(c, errcode.ErrTokenInvalid.WithCause(err))
			c.Abort()
			return
		}
		c.Set(CtxKeyUserID, claims.UserID)
		c.Set(CtxKeyUsername, claims.Username)

		// also propagate user_id via context.Context
		ctx := logger.WithUserID(c.Request.Context(), claims.UserID)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

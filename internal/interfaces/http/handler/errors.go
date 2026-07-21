package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/zhanbinb/go-clean-arch_demo/pkg/errcode"
	"github.com/zhanbinb/go-clean-arch_demo/pkg/response"
)

// writeErr translates any error to the unified response.
// If err is already *errcode.Error it is used as-is; otherwise wrapped into ErrInternal.
func writeErr(c *gin.Context, err error) {
	response.Error(c, errcode.FromError(err))
}

package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/zhanbinb/go-clean-arch_demo/internal/application/auth"
	"github.com/zhanbinb/go-clean-arch_demo/pkg/errcode"
	"github.com/zhanbinb/go-clean-arch_demo/pkg/response"
)

// AuthHandler exposes login / refresh endpoints (public, no JWT required).
type AuthHandler struct {
	svc *auth.Service
}

func NewAuthHandler(svc *auth.Service) *AuthHandler { return &AuthHandler{svc: svc} }

// Login godoc
// @Summary      Login
// @Description  Exchange username/password for an access + refresh token pair.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input  body  auth.LoginInput  true  "Credentials"
// @Success      200  {object}  response.Resp{data=auth.LoginResult}
// @Failure      401  {object}  response.Resp
// @Router       /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var in auth.LoginInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Error(c, errcode.ErrBadRequest.WithCause(err))
		return
	}
	out, err := h.svc.Login(c.Request.Context(), in)
	if err != nil {
		writeErr(c, err)
		return
	}
	response.OK(c, out)
}

// Refresh godoc
// @Summary      Refresh access token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input  body  auth.RefreshInput  true  "Refresh token"
// @Success      200  {object}  response.Resp{data=auth.LoginResult}
// @Failure      401  {object}  response.Resp
// @Router       /api/v1/auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	var in auth.RefreshInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Error(c, errcode.ErrBadRequest.WithCause(err))
		return
	}
	out, err := h.svc.Refresh(c.Request.Context(), in)
	if err != nil {
		writeErr(c, err)
		return
	}
	response.OK(c, out)
}

// Register godoc
// @Summary      Register a new user (admin/bootstrap)
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input  body  auth.RegisterInput  true  "Username and password"
// @Success      204
// @Failure      409  {object}  response.Resp
// @Router       /api/v1/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var in auth.RegisterInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Error(c, errcode.ErrBadRequest.WithCause(err))
		return
	}
	if err := h.svc.Register(c.Request.Context(), in.Username, in.Password); err != nil {
		writeErr(c, err)
		return
	}
	response.NoContent(c)
}

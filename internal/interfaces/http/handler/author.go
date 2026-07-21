package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/zhanbinb/go-clean-arch_demo/internal/application/author"
	"github.com/zhanbinb/go-clean-arch_demo/pkg/errcode"
	"github.com/zhanbinb/go-clean-arch_demo/pkg/response"
)

// AuthorHandler exposes HTTP endpoints for the author use case.
type AuthorHandler struct {
	svc *author.Service
}

func NewAuthorHandler(svc *author.Service) *AuthorHandler { return &AuthorHandler{svc: svc} }

// Create godoc
// @Summary      Create author
// @Tags         authors
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        input  body  author.CreateInput  true  "Author payload"
// @Success      201  {object}  response.Resp{data=author.AuthorDTO}
// @Failure      400  {object}  response.Resp
// @Failure      409  {object}  response.Resp
// @Router       /api/v1/authors [post]
func (h *AuthorHandler) Create(c *gin.Context) {
	var in author.CreateInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Error(c, errcode.ErrBadRequest.WithCause(err))
		return
	}
	out, err := h.svc.Create(c.Request.Context(), in)
	if err != nil {
		writeErr(c, err)
		return
	}
	response.Created(c, out)
}

// GetByID godoc
// @Summary      Get author by id
// @Tags         authors
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  int  true  "Author ID"
// @Success      200  {object}  response.Resp{data=author.AuthorDTO}
// @Failure      404  {object}  response.Resp
// @Router       /api/v1/authors/{id} [get]
func (h *AuthorHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errcode.ErrBadRequest.WithMsg("invalid id"))
		return
	}
	out, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		writeErr(c, err)
		return
	}
	response.OK(c, out)
}

// Update godoc
// @Summary      Update author (partial)
// @Tags         authors
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id     path  int                   true  "Author ID"
// @Param        input  body  author.UpdateInput    true  "Partial update"
// @Success      200  {object}  response.Resp{data=author.AuthorDTO}
// @Router       /api/v1/authors/{id} [put]
func (h *AuthorHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errcode.ErrBadRequest.WithMsg("invalid id"))
		return
	}
	var in author.UpdateInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Error(c, errcode.ErrBadRequest.WithCause(err))
		return
	}
	out, err := h.svc.Update(c.Request.Context(), id, in)
	if err != nil {
		writeErr(c, err)
		return
	}
	response.OK(c, out)
}

// Delete godoc
// @Summary      Delete author
// @Tags         authors
// @Security     BearerAuth
// @Param        id  path  int  true  "Author ID"
// @Success      204
// @Router       /api/v1/authors/{id} [delete]
func (h *AuthorHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errcode.ErrBadRequest.WithMsg("invalid id"))
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		writeErr(c, err)
		return
	}
	response.NoContent(c)
}

// List godoc
// @Summary      List authors
// @Tags         authors
// @Security     BearerAuth
// @Param        limit   query  int  false  "Page size"
// @Param        offset  query  int  false  "Page offset"
// @Success      200  {object}  response.Resp{data=[]author.AuthorDTO}
// @Router       /api/v1/authors [get]
func (h *AuthorHandler) List(c *gin.Context) {
	limit, _ := strconv.Atoi(c.Query("limit"))
	offset, _ := strconv.Atoi(c.Query("offset"))
	out, err := h.svc.List(c.Request.Context(), limit, offset)
	if err != nil {
		writeErr(c, err)
		return
	}
	response.OK(c, out)
}

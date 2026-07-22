// Package handler contains the HTTP request handlers.
package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/zhanbinb/go-clean-arch_demo/internal/application/article"
	"github.com/zhanbinb/go-clean-arch_demo/pkg/errcode"
	"github.com/zhanbinb/go-clean-arch_demo/pkg/response"
)

// ArticleHandler exposes HTTP endpoints for the article use case.
type ArticleHandler struct {
	svc *article.Service
}

// NewArticleHandler wires the handler.
func NewArticleHandler(svc *article.Service) *ArticleHandler { return &ArticleHandler{svc: svc} }

// Create godoc
// @Summary      Create article
// @Description  Create a new article. Requires authentication.
// @Tags         articles
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        input  body      article.CreateInput  true  "Article payload"
// @Success      201    {object}  response.Resp{data=article.ArticleDTO}
// @Failure      400    {object}  response.Resp
// @Failure      401    {object}  response.Resp
// @Router       /api/v1/articles [post]
func (h *ArticleHandler) Create(c *gin.Context) {
	var in article.CreateInput
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
// @Summary      Get article by id
// @Tags         articles
// @Produce      json
// @Security     BearerAuth
// @Param        id  path  int  true  "Article ID"
// @Success      200  {object}  response.Resp{data=article.ArticleDTO}
// @Failure      404  {object}  response.Resp
// @Router       /api/v1/articles/{id} [get]
func (h *ArticleHandler) GetByID(c *gin.Context) {
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

// List godoc
// @Summary      List articles (cursor pagination)
// @Tags         articles
// @Produce      json
// @Security     BearerAuth
// @Param        cursor  query  string  false  "Pagination cursor"
// @Param        limit   query  int     false  "Page size (default 10)"
// @Success      200  {object}  response.Resp{data=article.ListResult}
// @Router       /api/v1/articles [get]
func (h *ArticleHandler) List(c *gin.Context) {
	cursor := c.Query("cursor")
	limit, _ := strconv.Atoi(c.Query("limit"))
	out, err := h.svc.List(c.Request.Context(), cursor, limit)
	if err != nil {
		writeErr(c, err)
		return
	}
	response.OK(c, out)
}

// Update godoc
// @Summary      Update article (partial)
// @Tags         articles
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id     path  int                    true  "Article ID"
// @Param        input  body  article.UpdateInput    true  "Partial update"
// @Success      200  {object}  response.Resp{data=article.ArticleDTO}
// @Failure      404  {object}  response.Resp
// @Router       /api/v1/articles/{id} [put]
func (h *ArticleHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errcode.ErrBadRequest.WithMsg("invalid id"))
		return
	}
	var in article.UpdateInput
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
// @Summary      Delete article
// @Tags         articles
// @Security     BearerAuth
// @Param        id  path  int  true  "Article ID"
// @Success      204
// @Failure      404  {object}  response.Resp
// @Router       /api/v1/articles/{id} [delete]
func (h *ArticleHandler) Delete(c *gin.Context) {
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

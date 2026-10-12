package article

import (
	"github.com/gin-gonic/gin"

	"flexgin/internal/httpx"
)

type ArticleHandler struct {
	svc ArticleService
}

func NewArticleHandler(svc ArticleService) *ArticleHandler {
	return &ArticleHandler{svc: svc}
}

// ArticleHandlers attach errors with c.Error and return; middleware.ErrorArticleHandler renders them.

// @BasePath /api/v1

// Create godoc
// @Summary      Create an article
// @Description  Creates a new article
// @Tags         articles
// @Accept       json
// @Produce      json
// @Param        body  body      CreateArticleRequest  true  "Article data"
// @Success      201   {object}  httpx.Envelope        "Article data"
// @Failure      400   {object}  httpx.Envelope        "Validation error"
// @Router       /articles [post]
func (h *ArticleHandler) Create(c *gin.Context) {
	var req CreateArticleRequest
	if err := httpx.BindJSON(c, &req); err != nil {
		_ = c.Error(err)
		return
	}
	a, err := h.svc.Create(c.Request.Context(), req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	httpx.Created(c, a)
}

func (h *ArticleHandler) Get(c *gin.Context) {
	id := c.Param("id")
	// if err != nil {
	// 	_ = c.Error(err)
	// 	return
	// }
	a, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		_ = c.Error(err)
		return
	}
	httpx.OK(c, a)
}

func (h *ArticleHandler) List(c *gin.Context) {
	page, err := httpx.ParsePage(c)
	if err != nil {
		_ = c.Error(err)
		return
	}
	items, total, err := h.svc.List(c.Request.Context(), page.Limit(), page.Offset())
	if err != nil {
		_ = c.Error(err)
		return
	}
	httpx.List(c, items, page.Meta(total))
}

func (h *ArticleHandler) Update(c *gin.Context) {
	id := c.Param("id")
	// id, err := httpx.ParamInt64(c, "id")
	// if err != nil {
	// 	_ = c.Error(err)
	// 	return
	// }
	var req UpdateArticleRequest
	if err := httpx.BindJSON(c, &req); err != nil {
		_ = c.Error(err)
		return
	}
	a, err := h.svc.Update(c.Request.Context(), id, req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	httpx.OK(c, a)
}

func (h *ArticleHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	// id, err := httpx.ParamInt64(c, "id")
	// if err != nil {
	// 	_ = c.Error(err)
	// 	return
	// }
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		_ = c.Error(err)
		return
	}
	httpx.NoContent(c)
}

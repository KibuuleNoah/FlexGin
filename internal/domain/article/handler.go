package article

import (
	"github.com/gin-gonic/gin"

	"flexgin/internal/httpx"
	"flexgin/internal/store"
)

type ArticleHandler struct {
	svc ArticleService
}

func NewArticleHandler(svc ArticleService) *ArticleHandler {
	return &ArticleHandler{svc: svc}
}

// ArticleHandlers attach errors with c.Error and return; middleware.ErrorArticleHandler renders them.

// Create godoc
// @Summary      Create an article
// @Description  Creates a new article
// @Tags         articles
// @Accept       json
// @Produce      json
// @Param        body  body      CreateArticleRequest  true  					"Article data"
// @Success      201   {object}  httpx.EnvelopD{data=store.Article}         "Article data"
// @Failure      400   {object}  httpx.EnvelopD        								"Validation error"
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
	httpx.Created[store.Article](c, a)
}

// @Summary      Get an article
// @Description  Returns a single article
// @Tags         articles
// @Accept       json
// @Produce      json
// @Param        id  path      string  true  "Article ID" minlength(6) maxlength(6)
// @Success      200  {object}  httpx.EnvelopD{data=store.Article}        "Article data"
// @Failure      400  {object}  httpx.EnvelopD        				       "Validation error"
// @Failure      404  {object}  httpx.EnvelopD                       "Not found"
// @Router       /articles/{id} [get]
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
	httpx.OK[store.Article](c, a)
}

// @Summary      List articles
// @Description  Returns a list of articles
// @Tags         articles
// @Accept       json
// @Produce      json
// @Param        page  query     int     false  "Page number"
// @Param        per_page  query  int     false  "Number of items per page"
// @Success      200  {object}  httpx.EnvelopD{data=[]store.Article}        "Article data"
// @Failure      400  {object}  httpx.EnvelopD                         "Validation error"
// @Router       /articles [get]
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
	httpx.List[store.Article](c, items, page.Meta(total))
}

// @Summary      Update an article
// @Description  Updates a single article
// @Tags         articles
// @Accept       json
// @Produce      json
// @Param        id  path      string  true  "Article ID" minlength(6) maxlength(6)
// @Param        body  body      UpdateArticleRequest  true  "Article data"
// @Success      200  {object}  httpx.EnvelopD{data=store.Article}        "Article data"
// @Failure      400  {object}  httpx.EnvelopD                       "Validation error"
// @Failure      404  {object}  httpx.EnvelopD                       "Not found"
// @Router       /articles/{id} [patch]
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
	httpx.OK[store.Article](c, a)
}

// @Summary      Delete an article
// @Description  Deletes a single article
// @Tags         articles
// @Accept       json
// @Produce      json
// @Param        id  path      string  true  "Article ID" minlength(6) maxlength(6)
// @Success      204  {object}  string                  "Article deleted Successfully"
// @Failure      400  {object}  httpx.EnvelopD        "Validation error"
// @Failure      404  {object}  httpx.EnvelopD        "Not found"
// @Router       /articles/{id} [delete]
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

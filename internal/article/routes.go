package article

import (
	"flex/internal/store"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes mounts the article endpoints on the given group (e.g. /api/v1).
func RegisterArticleRoutes(rg *gin.RouterGroup, q store.Querier) {
	h := NewHandler(*NewService(q))
	g := rg.Group("/articles")
	g.POST("", h.Create)
	g.GET("", h.List)
	g.GET("/:id", h.Get)
	g.PATCH("/:id", h.Update)
	g.DELETE("/:id", h.Delete)
}

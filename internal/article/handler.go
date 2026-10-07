package article

import (
	"github.com/gin-gonic/gin"

	"FlexGin/internal/httpx"
)

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

// Handlers attach errors with c.Error and return; middleware.ErrorHandler renders them.

func (h *Handler) Create(c *gin.Context) {
	var req CreateRequest
	if err := httpx.BindJSON(c, &req); err != nil {
		_ = c.Error(err)
		return
	}
	a, err := h.svc.Create(c.Request.Context(), req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	httpx.Created(c, ToResponse(a))
}

func (h *Handler) Get(c *gin.Context) {
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
	httpx.OK(c, ToResponse(a))
}

func (h *Handler) List(c *gin.Context) {
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
	httpx.List(c, ToResponses(items), page.Meta(total))
}

func (h *Handler) Update(c *gin.Context) {
	id := c.Param("id")
	// id, err := httpx.ParamInt64(c, "id")
	// if err != nil {
	// 	_ = c.Error(err)
	// 	return
	// }
	var req UpdateRequest
	if err := httpx.BindJSON(c, &req); err != nil {
		_ = c.Error(err)
		return
	}
	a, err := h.svc.Update(c.Request.Context(), id, req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	httpx.OK(c, ToResponse(a))
}

func (h *Handler) Delete(c *gin.Context) {
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

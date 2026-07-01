package inventory

import (
	"strconv"

	"backend/internal/platform/response"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service        *Service
	authMiddleware gin.HandlerFunc
}

func NewHandler(service *Service, authMiddleware gin.HandlerFunc) *Handler {
	return &Handler{service: service, authMiddleware: authMiddleware}
}

func (h *Handler) RegisterRoutes(group *gin.RouterGroup) {
	group.GET("/inventory/ping", h.Ping)

	admin := group.Group("/admin/goods")
	if h.authMiddleware != nil {
		admin.Use(h.authMiddleware)
	}
	{
		admin.GET("", h.ListGoods)
		admin.GET("/:id", h.GetGoods)
		admin.POST("", h.CreateGoods)
		admin.GET("/categories", h.ListCategories)
		admin.POST("/categories", h.CreateCategory)
		admin.GET("/:id/skus", h.GetSKUs)
	}

	client := group.Group("/client/goods")
	{
		client.GET("/:id", h.GetGoodsPublic)
		client.GET("/categories", h.ListCategories)
	}
}

func (h *Handler) Ping(c *gin.Context) {
	response.Success(c, gin.H{"module": "inventory", "status": "ok"})
}

func (h *Handler) ListGoods(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	goods, total, err := h.service.ListGoods(c.Request.Context(), page, pageSize)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"list": goods, "total": total})
}

func (h *Handler) GetGoods(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	goods, err := h.service.GetGoods(c.Request.Context(), id)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, goods)
}

func (h *Handler) GetGoodsPublic(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	goods, err := h.service.GetGoods(c.Request.Context(), id)
	if err != nil {
		response.Fail(c, err)
		return
	}
	if goods != nil && !goods.IsVisible {
		response.Fail(c, nil)
		return
	}
	response.Success(c, goods)
}

func (h *Handler) CreateGoods(c *gin.Context) {
	response.Success(c, gin.H{"message": "create goods endpoint"})
}

func (h *Handler) ListCategories(c *gin.Context) {
	categories, err := h.service.ListCategories(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, categories)
}

func (h *Handler) CreateCategory(c *gin.Context) {
	response.Success(c, gin.H{"message": "create category endpoint"})
}

func (h *Handler) GetSKUs(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	skus, err := h.service.GetSKUsByGoodsID(c.Request.Context(), id)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, skus)
}

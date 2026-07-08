package inventory

import (
	"net/http"
	"strconv"

	apperrors "backend/internal/platform/errors"
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

	// Admin goods routes
	admin := group.Group("/admin/goods")
	if h.authMiddleware != nil {
		admin.Use(h.authMiddleware)
	}
	{
		admin.GET("", h.ListGoods)
		admin.GET("/:id", h.GetGoods)
		admin.POST("", h.CreateGoods)
		admin.PUT("/:id", h.UpdateGoods)
		admin.DELETE("/:id", h.DeleteGoods)
		admin.PUT("/:id/status", h.UpdateGoodsStatus)
		admin.GET("/stats", h.GetGoodsStats)
		admin.GET("/:id/skus", h.GetSKUs)
	}

	// Admin SKU routes
	adminSKU := group.Group("/admin/goods/sku")
	if h.authMiddleware != nil {
		adminSKU.Use(h.authMiddleware)
	}
	{
		adminSKU.POST("", h.CreateSKU)
		adminSKU.PUT("/:id", h.UpdateSKU)
		adminSKU.DELETE("/:id", h.DeleteSKU)
	}

	// Admin category routes
	adminCategories := group.Group("/admin/categories")
	if h.authMiddleware != nil {
		adminCategories.Use(h.authMiddleware)
	}
	{
		adminCategories.GET("", h.ListCategories)
		adminCategories.GET("/all", h.ListAllCategories)
		adminCategories.GET("/:id", h.GetCategory)
		adminCategories.POST("", h.CreateCategory)
		adminCategories.PUT("/:id", h.UpdateCategory)
		adminCategories.DELETE("/:id", h.DeleteCategory)
	}

	// Admin purchase limit routes
	adminPurchaseLimit := group.Group("/admin/purchase-limit")
	if h.authMiddleware != nil {
		adminPurchaseLimit.Use(h.authMiddleware)
	}
	{
		adminPurchaseLimit.GET("", h.ListPurchaseLimitRules)
		adminPurchaseLimit.POST("", h.CreatePurchaseLimitRule)
		adminPurchaseLimit.PUT("/:id", h.UpdatePurchaseLimitRule)
		adminPurchaseLimit.DELETE("/:id", h.DeletePurchaseLimitRule)
	}

	// Client goods routes
	client := group.Group("/client/goods")
	{
		client.GET("", h.ListGoods)
		client.GET("/categories", h.ListCategories)
		client.GET("/detail/:goodsId", h.GetGoodsDetailWithSKUs)
		client.POST("/sku/check", h.CheckSKUStock)
		client.GET("/:id", h.GetGoodsPublic)
	}

	// Client categories routes
	categories := group.Group("/client/categories")
	{
		categories.GET("", h.ListCategories)
		categories.GET("/:id", h.GetCategory)
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
	if id == 0 {
		response.Fail(c, apperrors.New(apperrors.CodeInvalidArgument, "invalid goods id", http.StatusBadRequest, nil))
		return
	}
	goods, err := h.service.GetGoods(c.Request.Context(), id)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, goods)
}

func (h *Handler) GetGoodsPublic(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if id == 0 {
		response.Fail(c, apperrors.New(apperrors.CodeInvalidArgument, "invalid goods id", http.StatusBadRequest, nil))
		return
	}
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

func (h *Handler) GetGoodsDetailWithSKUs(c *gin.Context) {
	goodsID, _ := strconv.ParseInt(c.Param("goodsId"), 10, 64)
	if goodsID == 0 {
		response.Fail(c, apperrors.New(apperrors.CodeInvalidArgument, "invalid goods id", http.StatusBadRequest, nil))
		return
	}
	goods, err := h.service.GetGoods(c.Request.Context(), goodsID)
	if err != nil {
		response.Fail(c, err)
		return
	}
	skus, err := h.service.GetSKUsByGoodsID(c.Request.Context(), goodsID)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"goods": goods, "skus": skus})
}

func (h *Handler) CreateGoods(c *gin.Context) {
	var req Goods
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, apperrors.New(apperrors.CodeInvalidArgument, "invalid input", http.StatusBadRequest, err))
		return
	}
	id, err := h.service.CreateGoods(c.Request.Context(), &req)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"id": id})
}

func (h *Handler) UpdateGoods(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if id == 0 {
		response.Fail(c, apperrors.New(apperrors.CodeInvalidArgument, "invalid goods id", http.StatusBadRequest, nil))
		return
	}
	var req Goods
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, apperrors.New(apperrors.CodeInvalidArgument, "invalid input", http.StatusBadRequest, err))
		return
	}
	req.ID = id
	if err := h.service.UpdateGoods(c.Request.Context(), &req); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"id": id})
}

func (h *Handler) DeleteGoods(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if id == 0 {
		response.Fail(c, apperrors.New(apperrors.CodeInvalidArgument, "invalid goods id", http.StatusBadRequest, nil))
		return
	}
	if err := h.service.DeleteGoods(c.Request.Context(), id); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"id": id})
}

func (h *Handler) UpdateGoodsStatus(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if id == 0 {
		response.Fail(c, apperrors.New(apperrors.CodeInvalidArgument, "invalid goods id", http.StatusBadRequest, nil))
		return
	}
	var req struct {
		Status int `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, apperrors.New(apperrors.CodeInvalidArgument, "invalid input", http.StatusBadRequest, err))
		return
	}
	if err := h.service.UpdateGoodsStatus(c.Request.Context(), id, req.Status); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"id": id, "status": req.Status})
}

func (h *Handler) GetGoodsStats(c *gin.Context) {
	stats, err := h.service.GetGoodsStats(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, stats)
}

func (h *Handler) ListCategories(c *gin.Context) {
	categories, err := h.service.ListCategories(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, categories)
}

func (h *Handler) ListAllCategories(c *gin.Context) {
	categories, err := h.service.ListAllCategories(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, categories)
}

func (h *Handler) GetCategory(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if id == 0 {
		response.Fail(c, apperrors.New(apperrors.CodeInvalidArgument, "invalid category id", http.StatusBadRequest, nil))
		return
	}
	category, err := h.service.GetCategory(c.Request.Context(), id)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, category)
}

func (h *Handler) CreateCategory(c *gin.Context) {
	var req GoodsCategory
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, apperrors.New(apperrors.CodeInvalidArgument, "invalid input", http.StatusBadRequest, err))
		return
	}
	if err := h.service.CreateCategory(c.Request.Context(), &req); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"id": req.ID})
}

func (h *Handler) UpdateCategory(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if id == 0 {
		response.Fail(c, apperrors.New(apperrors.CodeInvalidArgument, "invalid category id", http.StatusBadRequest, nil))
		return
	}
	var req GoodsCategory
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, apperrors.New(apperrors.CodeInvalidArgument, "invalid input", http.StatusBadRequest, err))
		return
	}
	req.ID = id
	if err := h.service.UpdateCategory(c.Request.Context(), &req); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"id": id})
}

func (h *Handler) DeleteCategory(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if id == 0 {
		response.Fail(c, apperrors.New(apperrors.CodeInvalidArgument, "invalid category id", http.StatusBadRequest, nil))
		return
	}
	if err := h.service.DeleteCategory(c.Request.Context(), id); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"id": id})
}

func (h *Handler) GetSKUs(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if id == 0 {
		response.Fail(c, apperrors.New(apperrors.CodeInvalidArgument, "invalid goods id", http.StatusBadRequest, nil))
		return
	}
	skus, err := h.service.GetSKUsByGoodsID(c.Request.Context(), id)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, skus)
}

func (h *Handler) CreateSKU(c *gin.Context) {
	var req GoodsSKU
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, apperrors.New(apperrors.CodeInvalidArgument, "invalid input", http.StatusBadRequest, err))
		return
	}
	id, err := h.service.CreateSKU(c.Request.Context(), &req)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"id": id})
}

func (h *Handler) UpdateSKU(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if id == 0 {
		response.Fail(c, apperrors.New(apperrors.CodeInvalidArgument, "invalid sku id", http.StatusBadRequest, nil))
		return
	}
	var req GoodsSKU
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, apperrors.New(apperrors.CodeInvalidArgument, "invalid input", http.StatusBadRequest, err))
		return
	}
	req.ID = id
	if err := h.service.UpdateSKU(c.Request.Context(), &req); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"id": id})
}

func (h *Handler) DeleteSKU(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if id == 0 {
		response.Fail(c, apperrors.New(apperrors.CodeInvalidArgument, "invalid sku id", http.StatusBadRequest, nil))
		return
	}
	if err := h.service.DeleteSKU(c.Request.Context(), id); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"id": id})
}

func (h *Handler) CheckSKUStock(c *gin.Context) {
	var req struct {
		SKUID    int64 `json:"skuId"`
		Quantity int   `json:"quantity"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, req.SKUID > 0 && req.Quantity > 0)
}

func (h *Handler) ListPurchaseLimitRules(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	rules, total, err := h.service.ListPurchaseLimitRules(c.Request.Context(), page, pageSize)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"list": rules, "total": total})
}

func (h *Handler) CreatePurchaseLimitRule(c *gin.Context) {
	var req PurchaseLimitRule
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, apperrors.New(apperrors.CodeInvalidArgument, "invalid input", http.StatusBadRequest, err))
		return
	}
	if err := h.service.CreatePurchaseLimitRule(c.Request.Context(), &req); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"id": req.ID})
}

func (h *Handler) UpdatePurchaseLimitRule(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if id == 0 {
		response.Fail(c, apperrors.New(apperrors.CodeInvalidArgument, "invalid rule id", http.StatusBadRequest, nil))
		return
	}
	var req PurchaseLimitRule
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, apperrors.New(apperrors.CodeInvalidArgument, "invalid input", http.StatusBadRequest, err))
		return
	}
	req.ID = id
	if err := h.service.UpdatePurchaseLimitRule(c.Request.Context(), &req); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"id": id})
}

func (h *Handler) DeletePurchaseLimitRule(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if id == 0 {
		response.Fail(c, apperrors.New(apperrors.CodeInvalidArgument, "invalid rule id", http.StatusBadRequest, nil))
		return
	}
	if err := h.service.DeletePurchaseLimitRule(c.Request.Context(), id); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"id": id})
}

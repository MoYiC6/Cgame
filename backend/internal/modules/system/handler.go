package system

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
	admin := group.Group("/admin/settings")
	if h.authMiddleware != nil {
		admin.Use(h.authMiddleware)
	}
	{
		admin.GET("", h.ListSettings)
		admin.GET("/prefix/:prefix", h.ListSettingsByPrefix)
		admin.GET("/:key", h.GetSetting)
		admin.PUT("/:key", h.SetSetting)
		admin.PUT("/batch", h.BatchSetSettings)
		admin.GET("/partner-config", h.GetPartnerConfig)
		admin.PUT("/partner-config", h.SetPartnerConfig)
		admin.GET("/faceid/config/:id", h.GetFaceIdConfig)
		admin.POST("/faceid/config", h.CreateFaceIdConfig)
		admin.PUT("/faceid/config/:id", h.UpdateFaceIdConfig)
		admin.DELETE("/faceid/config/:id", h.DeleteFaceIdConfig)
		admin.GET("/realname/logs", h.ListRealNameVerifyLogs)
	}

	client := group.Group("/client/settings")
	{
		client.GET("/notice", h.GetNotice)
		client.GET("/system-name", h.GetSystemName)
	}
}

func (h *Handler) ListSettings(c *gin.Context) {
	settings, err := h.service.GetAllSettings(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, settings)
}

func (h *Handler) ListSettingsByPrefix(c *gin.Context) {
	prefix := c.Param("prefix")
	settings, err := h.service.ListSettings(c.Request.Context(), prefix)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, settings)
}

func (h *Handler) GetSetting(c *gin.Context) {
	key := c.Param("key")
	value, err := h.service.GetSetting(c.Request.Context(), key)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"key": key, "value": value})
}

func (h *Handler) SetSetting(c *gin.Context) {
	key := c.Param("key")
	var req struct {
		Value string `json:"value"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, err)
		return
	}
	if err := h.service.SetSetting(c.Request.Context(), key, req.Value); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *Handler) BatchSetSettings(c *gin.Context) {
	var req map[string]string
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, err)
		return
	}
	for k, v := range req {
		if err := h.service.SetSetting(c.Request.Context(), k, v); err != nil {
			response.Fail(c, err)
			return
		}
	}
	response.Success(c, nil)
}

func (h *Handler) GetPartnerConfig(c *gin.Context) {
	value, err := h.service.GetPartnerConfig(c.Request.Context(), "default_partner_enabled")
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"enabled": value == "true"})
}

func (h *Handler) SetPartnerConfig(c *gin.Context) {
	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, err)
		return
	}
	value := "false"
	if req.Enabled {
		value = "true"
	}
	if err := h.service.SetPartnerConfig(c.Request.Context(), "default_partner_enabled", value); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *Handler) GetNotice(c *gin.Context) {
	notice, err := h.service.GetSetting(c.Request.Context(), "home.notice")
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"notice": notice})
}

func (h *Handler) GetSystemName(c *gin.Context) {
	name, err := h.service.GetSetting(c.Request.Context(), "system.name")
	if err != nil {
		name = "FeYo电竞"
	}
	response.Success(c, gin.H{"name": name})
}

func (h *Handler) GetFaceIdConfig(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, err)
		return
	}
	_ = id
	response.Success(c, gin.H{"message": "faceid config endpoint"})
}

func (h *Handler) CreateFaceIdConfig(c *gin.Context) {
	response.Success(c, gin.H{"message": "create faceid config endpoint"})
}

func (h *Handler) UpdateFaceIdConfig(c *gin.Context) {
	response.Success(c, gin.H{"message": "update faceid config endpoint"})
}

func (h *Handler) DeleteFaceIdConfig(c *gin.Context) {
	response.Success(c, gin.H{"message": "delete faceid config endpoint"})
}

func (h *Handler) ListRealNameVerifyLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	logs, total, err := h.service.ListRealNameVerifyLogs(c.Request.Context(), nil, "", page, pageSize)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"list": logs, "total": total})
}

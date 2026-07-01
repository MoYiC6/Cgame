package external

import (
	"fmt"
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
	wechat := group.Group("/wechat")
	{
		wechat.GET("/oauth/website", h.WebsiteOAuth)
		wechat.GET("/oauth/website/callback", h.WebsiteOAuthCallback)
		wechat.POST("/login", h.WechatLogin)
		wechat.POST("/bind", h.WechatBind)
		wechat.DELETE("/unbind", h.WechatUnbind)
		wechat.POST("/phone", h.WechatPhone)
		wechat.GET("/scan-login/generate-qrcode", h.GenerateScanLoginQR)
		wechat.GET("/scan-login/check-status", h.CheckScanLoginStatus)
		wechat.POST("/scan-login/scan", h.ScanLoginScan)
		wechat.POST("/scan-login/confirm", h.ScanLoginConfirm)
		wechat.DELETE("/scan-login/cancel", h.ScanLoginCancel)
	}

	admin := group.Group("/admin/wxpay/config")
	if h.authMiddleware != nil {
		admin.Use(h.authMiddleware)
	}
	{
		admin.GET("/page", h.ListWxPayConfigs)
		admin.GET("/:id", h.GetWxPayConfig)
		admin.POST("", h.CreateWxPayConfig)
		admin.PUT("/:id", h.UpdateWxPayConfig)
		admin.DELETE("/:id", h.DeleteWxPayConfig)
		admin.PUT("/:id/status", h.UpdateWxPayConfigStatus)
		admin.GET("/type/:configType", h.GetWxPayConfigByType)
	}
}

func (h *Handler) WebsiteOAuth(c *gin.Context) {
	redirect := c.Query("redirect")
	if redirect == "" {
		response.Fail(c, fmt.Errorf("redirect is required"))
		return
	}
	c.Redirect(302, redirect)
}

func (h *Handler) WebsiteOAuthCallback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")
	if code == "" {
		response.Fail(c, fmt.Errorf("code is required"))
		return
	}
	_ = state
	response.Success(c, gin.H{"code": code})
}

func (h *Handler) WechatLogin(c *gin.Context) {
	var req WechatLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, err)
		return
	}
	oauth, token, err := h.service.WechatLogin(c.Request.Context(), req.Platform, req.Code, req.AppID)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"oauth": oauth, "token": token})
}

func (h *Handler) WechatBind(c *gin.Context) {
	userID, _ := strconv.ParseInt(c.GetString("userID"), 10, 64)
	var req WechatBindRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, err)
		return
	}
	oauth, err := h.service.WechatBind(c.Request.Context(), userID, req.Platform, req.Code)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, oauth)
}

func (h *Handler) WechatUnbind(c *gin.Context) {
	userID, _ := strconv.ParseInt(c.GetString("userID"), 10, 64)
	platform := c.Query("platform")
	if platform == "" {
		platform = PlatformWechat
	}
	if err := h.service.WechatUnbind(c.Request.Context(), userID, platform); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *Handler) WechatPhone(c *gin.Context) {
	var req WechatPhoneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, err)
		return
	}
	phone, err := h.service.GetWechatPhone(c.Request.Context(), req.Code)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, phone)
}

func (h *Handler) GenerateScanLoginQR(c *gin.Context) {
	session, err := h.service.GenerateScanLoginQR(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, session)
}

func (h *Handler) CheckScanLoginStatus(c *gin.Context) {
	loginKey := c.Query("loginKey")
	if loginKey == "" {
		response.Fail(c, fmt.Errorf("loginKey is required"))
		return
	}
	session, err := h.service.CheckScanLoginStatus(c.Request.Context(), loginKey)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, session)
}

func (h *Handler) ScanLoginScan(c *gin.Context) {
	loginKey := c.Query("loginKey")
	if loginKey == "" {
		response.Fail(c, fmt.Errorf("loginKey is required"))
		return
	}
	response.Success(c, nil)
}

func (h *Handler) ScanLoginConfirm(c *gin.Context) {
	loginKey := c.Query("loginKey")
	if loginKey == "" {
		response.Fail(c, fmt.Errorf("loginKey is required"))
		return
	}
	var req struct {
		UserID int64  `json:"userId"`
		Token  string `json:"token"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, err)
		return
	}
	if err := h.service.ConfirmScanLogin(c.Request.Context(), loginKey, req.UserID, req.Token); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *Handler) ScanLoginCancel(c *gin.Context) {
	loginKey := c.Query("loginKey")
	if loginKey == "" {
		response.Fail(c, fmt.Errorf("loginKey is required"))
		return
	}
	response.Success(c, nil)
}

func (h *Handler) ListWxPayConfigs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	configType := c.Query("configType")
	
	var typePtr *string
	if configType != "" {
		typePtr = &configType
	}

	configs, total, err := h.service.ListWxPayConfigs(c.Request.Context(), page, pageSize, typePtr)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"total": total, "rows": configs})
}

func (h *Handler) GetWxPayConfig(c *gin.Context) {
	configType := c.Query("configType")
	
	var config *WxPayConfig
	var err error
	if configType != "" {
		config, err = h.service.GetWxPayConfig(c.Request.Context(), configType)
	} else {
		response.Fail(c, fmt.Errorf("configType is required"))
		return
	}
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, config)
}

func (h *Handler) GetWxPayConfigByType(c *gin.Context) {
	configType := c.Param("configType")
	config, err := h.service.GetWxPayConfig(c.Request.Context(), configType)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, config)
}

func (h *Handler) CreateWxPayConfig(c *gin.Context) {
	var config WxPayConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		response.Fail(c, err)
		return
	}
	id, err := h.service.CreateWxPayConfig(c.Request.Context(), &config)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"id": id})
}

func (h *Handler) UpdateWxPayConfig(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var config WxPayConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		response.Fail(c, err)
		return
	}
	config.ID = id
	if err := h.service.UpdateWxPayConfig(c.Request.Context(), &config); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *Handler) UpdateWxPayConfigStatus(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req struct {
		Status int `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, err)
		return
	}
	config := &WxPayConfig{ID: id, Status: req.Status}
	if err := h.service.UpdateWxPayConfig(c.Request.Context(), config); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *Handler) DeleteWxPayConfig(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.service.DeleteWxPayConfig(c.Request.Context(), id); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil)
}

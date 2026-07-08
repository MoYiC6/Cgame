package teacher

import (
	"net/http"
	"strconv"
	"time"

	apperrors "backend/internal/platform/errors"
	"backend/internal/platform/response"
	"backend/internal/platform/security"

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
	// Client teacher routes (authenticated)
	client := group.Group("/client/teacher")
	if h.authMiddleware != nil {
		client.Use(h.authMiddleware)
	}
	{
		client.GET("/my-status", h.GetMyStatus)
		client.GET("/levels", h.GetLevels)
		client.PUT("/online-status", h.UpdateOnlineStatus)
		client.GET("/status", h.GetCurrentStatus)
		client.POST("/heartbeat", h.TeacherHeartbeat)
		client.PUT("/auto-status", h.SetAutoStatus)
		client.GET("/auto-status", h.GetAutoStatus)
		client.POST("/application", h.ApplyTeacher)
	}

	// Public client routes
	public := group.Group("/client")
	{
		public.GET("/teachers", h.ListTeachers)
		public.GET("/teachers/:id", h.GetTeacherDetail)
	}

	// Teacher public routes (no auth required)
	teacher := group.Group("/teacher")
	{
		teacher.GET("/ranking", h.GetTeacherRanking)
		teacher.GET("/dashboard/stats", h.GetTeacherDashboardStats)
	}

	// Admin teacher routes
	admin := group.Group("/admin/teachers")
	if h.authMiddleware != nil {
		admin.Use(h.authMiddleware)
	}
	{
		admin.GET("", h.ListTeachers)
		admin.GET("/:id", h.GetAdminTeacherDetail)
		admin.PUT("/:id/status", h.UpdateTeacherStatus)
		admin.POST("/batch-status", h.BatchUpdateTeacherStatus)
		admin.GET("/:id/status-log", h.GetTeacherStatusLogs)
		admin.POST("/:id/verify", h.VerifyTeacher)
	}

	// Admin teacher application routes
	adminApps := group.Group("/admin/teacher/applications")
	if h.authMiddleware != nil {
		adminApps.Use(h.authMiddleware)
	}
	{
		adminApps.GET("", h.ListApplications)
		adminApps.POST("/:id/approve", h.ApproveApplication)
		adminApps.POST("/:id/reject", h.RejectApplication)
	}

	// Admin teacher level routes
	adminLevels := group.Group("/admin/teacher/levels")
	if h.authMiddleware != nil {
		adminLevels.Use(h.authMiddleware)
	}
	{
		adminLevels.GET("", h.GetLevels)
		adminLevels.PUT("/:id", h.UpdateTeacherLevel)
		adminLevels.DELETE("/:id", h.DeleteTeacherLevel)
		adminLevels.GET("/:id/goods", h.GetLevelGoods)
		adminLevels.PUT("/:id/goods", h.UpdateLevelGoods)
	}
}

func currentUserID(c *gin.Context) (int64, bool) {
	if principal, ok := security.PrincipalFromContext(c.Request.Context()); ok {
		userID, err := strconv.ParseInt(principal.UserID, 10, 64)
		if err == nil && userID != 0 {
			return userID, true
		}
	}
	userID, err := strconv.ParseInt(c.GetString("userID"), 10, 64)
	if err != nil || userID == 0 {
		return 0, false
	}
	return userID, true
}

func (h *Handler) GetMyStatus(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Fail(c, apperrors.New(apperrors.CodeForbidden, "unauthorized", http.StatusForbidden, nil))
		return
	}
	teacher, err := h.service.GetTeacher(c.Request.Context(), userID)
	if err != nil {
		response.Fail(c, err)
		return
	}
	if teacher == nil {
		response.Success(c, gin.H{"isTeacher": false})
		return
	}
	response.Success(c, teacher)
}

func (h *Handler) GetLevels(c *gin.Context) {
	levels, err := h.service.GetTeacherLevels(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, levels)
}

func (h *Handler) ListTeachers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	teachers, total, err := h.service.ListTeachers(c.Request.Context(), page, pageSize)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"list": teachers, "total": total})
}

func (h *Handler) GetTeacherDetail(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if id == 0 {
		response.Fail(c, apperrors.New(apperrors.CodeInvalidArgument, "invalid teacher id", http.StatusBadRequest, nil))
		return
	}
	teacher, err := h.service.GetTeacherByID(c.Request.Context(), id)
	if err != nil {
		response.Fail(c, err)
		return
	}
	if teacher == nil {
		response.Fail(c, apperrors.New(apperrors.CodeNotFound, "teacher not found", http.StatusNotFound, nil))
		return
	}
	response.Success(c, teacher)
}

func (h *Handler) UpdateOnlineStatus(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Fail(c, apperrors.New(apperrors.CodeForbidden, "unauthorized", http.StatusForbidden, nil))
		return
	}
	var req struct {
		Status int `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, apperrors.New(apperrors.CodeInvalidArgument, "invalid input", http.StatusBadRequest, err))
		return
	}
	if err := h.service.UpdateOnlineStatus(c.Request.Context(), userID, req.Status); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *Handler) GetCurrentStatus(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Fail(c, apperrors.New(apperrors.CodeForbidden, "unauthorized", http.StatusForbidden, nil))
		return
	}
	status, err := h.service.GetCurrentStatus(c.Request.Context(), userID)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"status": status})
}

func (h *Handler) TeacherHeartbeat(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Fail(c, apperrors.New(apperrors.CodeForbidden, "unauthorized", http.StatusForbidden, nil))
		return
	}
	if err := h.service.TeacherHeartbeat(c.Request.Context(), userID); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"timestamp": time.Now().Unix()})
}

func (h *Handler) SetAutoStatus(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Fail(c, apperrors.New(apperrors.CodeForbidden, "unauthorized", http.StatusForbidden, nil))
		return
	}
	var req struct {
		Enabled     bool    `json:"enabled"`
		OnlineTime  *string `json:"onlineTime"`
		OfflineTime *string `json:"offlineTime"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, apperrors.New(apperrors.CodeInvalidArgument, "invalid input", http.StatusBadRequest, err))
		return
	}
	if err := h.service.SetAutoStatus(c.Request.Context(), userID, req.Enabled, req.OnlineTime, req.OfflineTime); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *Handler) GetAutoStatus(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Fail(c, apperrors.New(apperrors.CodeForbidden, "unauthorized", http.StatusForbidden, nil))
		return
	}
	autoStatus, err := h.service.GetAutoStatus(c.Request.Context(), userID)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, autoStatus)
}

func (h *Handler) ApplyTeacher(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Fail(c, apperrors.New(apperrors.CodeForbidden, "unauthorized", http.StatusForbidden, nil))
		return
	}
	var req struct {
		Name      string         `json:"name" binding:"required"`
		Mobile    *string        `json:"mobile"`
		Avatar    *string        `json:"avatar"`
		Platforms map[string]any `json:"platforms"`
		Tags      []string       `json:"tags"`
		Intro     *string        `json:"intro"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, apperrors.New(apperrors.CodeInvalidArgument, "invalid input", http.StatusBadRequest, err))
		return
	}
	app := &TeacherApplication{
		UserID:    userID,
		Name:      req.Name,
		Mobile:    req.Mobile,
		Avatar:    req.Avatar,
		Platforms: req.Platforms,
		Tags:      req.Tags,
		Intro:     req.Intro,
		Status:    0, // pending
	}
	id, err := h.service.ApplyTeacher(c.Request.Context(), app)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"id": id})
}

func (h *Handler) GetTeacherRanking(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	items, err := h.service.GetTeacherRanking(c.Request.Context(), limit)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"list": items})
}

func (h *Handler) GetTeacherDashboardStats(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Fail(c, apperrors.New(apperrors.CodeForbidden, "unauthorized", http.StatusForbidden, nil))
		return
	}
	stats, err := h.service.GetTeacherDashboardStats(c.Request.Context(), userID)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, stats)
}

func (h *Handler) GetAdminTeacherDetail(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if id == 0 {
		response.Fail(c, apperrors.New(apperrors.CodeInvalidArgument, "invalid teacher id", http.StatusBadRequest, nil))
		return
	}
	teacher, err := h.service.GetTeacherByID(c.Request.Context(), id)
	if err != nil {
		response.Fail(c, err)
		return
	}
	if teacher == nil {
		response.Fail(c, apperrors.New(apperrors.CodeNotFound, "teacher not found", http.StatusNotFound, nil))
		return
	}
	response.Success(c, teacher)
}

func (h *Handler) UpdateTeacherStatus(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if id == 0 {
		response.Fail(c, apperrors.New(apperrors.CodeInvalidArgument, "invalid teacher id", http.StatusBadRequest, nil))
		return
	}
	var req struct {
		Status int     `json:"status" binding:"required"`
		Reason *string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, apperrors.New(apperrors.CodeInvalidArgument, "invalid input", http.StatusBadRequest, err))
		return
	}
	operatorID, _ := currentUserID(c)
	if err := h.service.UpdateTeacherStatus(c.Request.Context(), id, req.Status, req.Reason, operatorID); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *Handler) BatchUpdateTeacherStatus(c *gin.Context) {
	var req struct {
		IDs    []int64 `json:"ids" binding:"required"`
		Status int     `json:"status" binding:"required"`
		Reason *string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, apperrors.New(apperrors.CodeInvalidArgument, "invalid input", http.StatusBadRequest, err))
		return
	}
	operatorID, _ := currentUserID(c)
	if err := h.service.BatchUpdateTeacherStatus(c.Request.Context(), req.IDs, req.Status, req.Reason, operatorID); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *Handler) GetTeacherStatusLogs(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if id == 0 {
		response.Fail(c, apperrors.New(apperrors.CodeInvalidArgument, "invalid teacher id", http.StatusBadRequest, nil))
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	logs, total, err := h.service.GetTeacherStatusLogs(c.Request.Context(), id, page, pageSize)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"list": logs, "total": total})
}

func (h *Handler) VerifyTeacher(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if id == 0 {
		response.Fail(c, apperrors.New(apperrors.CodeInvalidArgument, "invalid teacher id", http.StatusBadRequest, nil))
		return
	}
	if err := h.service.VerifyTeacher(c.Request.Context(), id); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *Handler) ListApplications(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	statusStr := c.Query("status")
	var status *int
	if statusStr != "" {
		if s, err := strconv.Atoi(statusStr); err == nil {
			status = &s
		}
	}
	apps, total, err := h.service.ListApplications(c.Request.Context(), status, page, pageSize)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"list": apps, "total": total})
}

func (h *Handler) ApproveApplication(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if id == 0 {
		response.Fail(c, apperrors.New(apperrors.CodeInvalidArgument, "invalid application id", http.StatusBadRequest, nil))
		return
	}
	operatorID, _ := currentUserID(c)
	if err := h.service.ApproveApplication(c.Request.Context(), id, operatorID); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *Handler) RejectApplication(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if id == 0 {
		response.Fail(c, apperrors.New(apperrors.CodeInvalidArgument, "invalid application id", http.StatusBadRequest, nil))
		return
	}
	var req struct {
		Reason string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, apperrors.New(apperrors.CodeInvalidArgument, "invalid input", http.StatusBadRequest, err))
		return
	}
	operatorID, _ := currentUserID(c)
	if err := h.service.RejectApplication(c.Request.Context(), id, req.Reason, operatorID); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *Handler) UpdateTeacherLevel(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if id == 0 {
		response.Fail(c, apperrors.New(apperrors.CodeInvalidArgument, "invalid level id", http.StatusBadRequest, nil))
		return
	}
	var req struct {
		Name           string  `json:"name" binding:"required"`
		MinOrders      int     `json:"minOrders"`
		CommissionRate float64 `json:"commissionRate"`
		Priority       int     `json:"priority"`
		Status         int     `json:"status"`
		Description    *string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, apperrors.New(apperrors.CodeInvalidArgument, "invalid input", http.StatusBadRequest, err))
		return
	}
	level := &TeacherLevel{
		ID:             id,
		Name:           req.Name,
		MinOrders:      req.MinOrders,
		CommissionRate: req.CommissionRate,
		Priority:       req.Priority,
		Status:         &req.Status,
		Description:    req.Description,
	}
	if err := h.service.UpdateTeacherLevel(c.Request.Context(), level); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *Handler) DeleteTeacherLevel(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if id == 0 {
		response.Fail(c, apperrors.New(apperrors.CodeInvalidArgument, "invalid level id", http.StatusBadRequest, nil))
		return
	}
	if err := h.service.DeleteTeacherLevel(c.Request.Context(), id); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *Handler) GetLevelGoods(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if id == 0 {
		response.Fail(c, apperrors.New(apperrors.CodeInvalidArgument, "invalid level id", http.StatusBadRequest, nil))
		return
	}
	goods, err := h.service.GetLevelGoods(c.Request.Context(), id)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"list": goods})
}

func (h *Handler) UpdateLevelGoods(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if id == 0 {
		response.Fail(c, apperrors.New(apperrors.CodeInvalidArgument, "invalid level id", http.StatusBadRequest, nil))
		return
	}
	var req []struct {
		GoodsID        int64   `json:"goodsId" binding:"required"`
		CommissionRate float64 `json:"commissionRate"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, apperrors.New(apperrors.CodeInvalidArgument, "invalid input", http.StatusBadRequest, err))
		return
	}
	goods := make([]*TeacherLevelGoods, 0, len(req))
	for _, item := range req {
		goods = append(goods, &TeacherLevelGoods{
			LevelID:        id,
			GoodsID:        item.GoodsID,
			CommissionRate: item.CommissionRate,
		})
	}
	if err := h.service.UpdateLevelGoods(c.Request.Context(), id, goods); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil)
}

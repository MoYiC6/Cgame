package teacher

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
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
		client.GET("/realname/status", h.GetRealNameStatus)
		client.POST("/realname/face/initiate", h.InitiateFaceVerification)
		client.POST("/realname/face/verify", h.VerifyFace)
		client.GET("/payment-info", h.GetPaymentInfo)
		client.PUT("/payment-info", h.UpdatePaymentInfo)
		client.GET("/intro", h.GetIntro)
		client.PUT("/intro", h.UpdateIntro)
		client.GET("/orders", h.ListTeacherOrders)
		client.GET("/dynamics/:teacherId", h.GetTeacherDynamics)
		client.GET("/reviews/:teacherId", h.GetTeacherReviews)
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
		admin.GET("/:id/orders", h.ListAdminTeacherOrders)
		admin.PUT("/:id/status", h.UpdateTeacherStatus)
		admin.POST("/batch-status", h.BatchUpdateTeacherStatus)
		admin.GET("/:id/status-log", h.GetTeacherStatusLogs)
		admin.POST("/:id/verify", h.VerifyTeacher)
		admin.POST("/upgrade/manual/:teacherId", h.ManualUpgrade)
		admin.POST("/upgrade/check/:teacherId", h.CheckUpgrade)
		admin.GET("/upgrade/history", h.GetUpgradeHistory)
		admin.GET("/:teacherId/assessment-videos/:videoId", h.GetAssessmentVideo)
		admin.PUT("/:teacherId/assessment-videos/:videoId", h.UpdateAssessmentVideo)
		admin.PUT("/:teacherId/assessment-videos/:videoId/enabled", h.ToggleAssessmentVideo)
		admin.DELETE("/:teacherId/assessment-videos/:videoId", h.DeleteAssessmentVideo)
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

	// Teacher income routes
	adminIncome := group.Group("/admin/teacher-income")
	if h.authMiddleware != nil {
		adminIncome.Use(h.authMiddleware)
	}
	{
		adminIncome.GET("", h.ListTeacherIncome)
	}
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
		adminLevels.GET("/export", h.ExportTeacherLevels)
		adminLevels.POST("/import", h.ImportTeacherLevels)
		adminLevels.GET("/export/template", h.ExportTeacherLevelTemplate)
	}

	// Admin teacher dynamics routes
	adminDynamics := group.Group("/admin/teacher-dynamics")
	if h.authMiddleware != nil {
		adminDynamics.Use(h.authMiddleware)
	}
	{
		adminDynamics.GET("", h.ListAdminDynamics)
		adminDynamics.DELETE("/:id", h.DeleteDynamic)
		adminDynamics.DELETE("/batch", h.BatchDeleteDynamics)
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

func (h *Handler) GetPaymentInfo(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Fail(c, apperrors.New(apperrors.CodeForbidden, "unauthorized", http.StatusForbidden, nil))
		return
	}
	info, err := h.service.GetTeacherPaymentInfo(c.Request.Context(), userID)
	if err != nil {
		response.Fail(c, err)
		return
	}
	if info == nil {
		response.Success(c, gin.H{})
		return
	}
	response.Success(c, info)
}

func (h *Handler) UpdatePaymentInfo(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Fail(c, apperrors.New(apperrors.CodeForbidden, "unauthorized", http.StatusForbidden, nil))
		return
	}
	var req struct {
		AlipayAccount *string `json:"alipayAccount"`
		BankName      *string `json:"bankName"`
		BankAccount   *string `json:"bankAccount"`
		RealName      *string `json:"realName"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, apperrors.New(apperrors.CodeInvalidArgument, "invalid input", http.StatusBadRequest, err))
		return
	}
	info := &TeacherPaymentInfo{
		AlipayAccount: req.AlipayAccount,
		BankName:      req.BankName,
		BankAccount:   req.BankAccount,
		RealName:      req.RealName,
	}
	if err := h.service.UpdateTeacherPaymentInfo(c.Request.Context(), userID, info); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *Handler) GetIntro(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Fail(c, apperrors.New(apperrors.CodeForbidden, "unauthorized", http.StatusForbidden, nil))
		return
	}
	intro, err := h.service.GetTeacherIntro(c.Request.Context(), userID)
	if err != nil {
		response.Fail(c, err)
		return
	}
	if intro == nil {
		response.Success(c, gin.H{"intro": "", "tags": []string{}})
		return
	}
	response.Success(c, intro)
}

func (h *Handler) UpdateIntro(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Fail(c, apperrors.New(apperrors.CodeForbidden, "unauthorized", http.StatusForbidden, nil))
		return
	}
	var req struct {
		Intro string   `json:"intro"`
		Tags  []string `json:"tags"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, apperrors.New(apperrors.CodeInvalidArgument, "invalid input", http.StatusBadRequest, err))
		return
	}
	intro := &TeacherIntro{
		Intro: req.Intro,
		Tags:  req.Tags,
	}
	if err := h.service.UpdateTeacherIntro(c.Request.Context(), userID, intro); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil)
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

func (h *Handler) ExportTeacherLevels(c *gin.Context) {
	levels, err := h.service.GetTeacherLevels(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=teacher-levels.csv")
	var sb strings.Builder
	sb.WriteString("ID,Name,MinOrders,CommissionRate,Priority,Status,Description\n")
	for _, l := range levels {
		status := 0
		if l.Status != nil {
			status = *l.Status
		}
		desc := ""
		if l.Description != nil {
			desc = *l.Description
		}
		sb.WriteString(fmt.Sprintf("%d,%s,%d,%.2f,%d,%d,%s\n", l.ID, l.Name, l.MinOrders, l.CommissionRate, l.Priority, status, desc))
	}
	c.String(http.StatusOK, sb.String())
}

func (h *Handler) ImportTeacherLevels(c *gin.Context) {
	// TODO: implement CSV import parsing
	response.Success(c, gin.H{"message": "import placeholder - parse CSV and create levels"})
}

func (h *Handler) ExportTeacherLevelTemplate(c *gin.Context) {
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=teacher-levels-template.csv")
	c.String(http.StatusOK, "Name,MinOrders,CommissionRate,Priority,Status,Description\nExample Level,10,0.15,1,1,Description here\n")
}

func (h *Handler) GetRealNameStatus(c *gin.Context) {
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
		response.Fail(c, apperrors.New(apperrors.CodeNotFound, "teacher not found", http.StatusNotFound, nil))
		return
	}
	// Return real name status from user table (simplified)
	response.Success(c, gin.H{"status": 0, "message": "not verified"})
}

func (h *Handler) InitiateFaceVerification(c *gin.Context) {
	// TODO: integrate with FaceID SDK
	response.Success(c, gin.H{"message": "face verification initiation placeholder", "verifyUrl": ""})
}

func (h *Handler) VerifyFace(c *gin.Context) {
	// TODO: integrate with FaceID SDK
	response.Success(c, gin.H{"message": "face verification placeholder", "verified": false})
}

func (h *Handler) ListTeacherOrders(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Fail(c, apperrors.New(apperrors.CodeForbidden, "unauthorized", http.StatusForbidden, nil))
		return
	}
	// Get teacher first
	teacher, err := h.service.GetTeacher(c.Request.Context(), userID)
	if err != nil {
		response.Fail(c, err)
		return
	}
	if teacher == nil {
		response.Fail(c, apperrors.New(apperrors.CodeNotFound, "teacher not found", http.StatusNotFound, nil))
		return
	}
	// Placeholder: return empty list (order module has the actual data)
	response.Success(c, gin.H{"list": []any{}, "total": 0, "teacherId": teacher.ID})
}

func (h *Handler) ListAdminTeacherOrders(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if id == 0 {
		response.Fail(c, apperrors.New(apperrors.CodeInvalidArgument, "invalid teacher id", http.StatusBadRequest, nil))
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	// Placeholder: return empty list
	response.Success(c, gin.H{"list": []any{}, "total": 0, "teacherId": id, "page": page, "pageSize": pageSize})
}

func (h *Handler) ListTeacherIncome(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	teacherID, _ := strconv.ParseInt(c.Query("teacherId"), 10, 64)
	// Placeholder: return empty list (actual data from order/withdrawal modules)
	response.Success(c, gin.H{"list": []any{}, "total": 0, "teacherId": teacherID, "page": page, "pageSize": pageSize})
}

func (h *Handler) GetTeacherDynamics(c *gin.Context) {
	teacherID, _ := strconv.ParseInt(c.Param("teacherId"), 10, 64)
	if teacherID == 0 {
		response.Fail(c, apperrors.New(apperrors.CodeInvalidArgument, "invalid teacher id", http.StatusBadRequest, nil))
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	// Placeholder: return empty list (needs dynamics table)
	response.Success(c, gin.H{"list": []any{}, "total": 0, "teacherId": teacherID, "page": page, "pageSize": pageSize})
}

func (h *Handler) GetTeacherReviews(c *gin.Context) {
	teacherID, _ := strconv.ParseInt(c.Param("teacherId"), 10, 64)
	if teacherID == 0 {
		response.Fail(c, apperrors.New(apperrors.CodeInvalidArgument, "invalid teacher id", http.StatusBadRequest, nil))
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	// Placeholder: return empty list (needs reviews table or order module integration)
	response.Success(c, gin.H{"list": []any{}, "total": 0, "teacherId": teacherID, "page": page, "pageSize": pageSize})
}

func (h *Handler) ManualUpgrade(c *gin.Context) {
	teacherID, _ := strconv.ParseInt(c.Param("teacherId"), 10, 64)
	if teacherID == 0 {
		response.Fail(c, apperrors.New(apperrors.CodeInvalidArgument, "invalid teacher id", http.StatusBadRequest, nil))
		return
	}
	var req struct {
		LevelID int64 `json:"levelId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, apperrors.New(apperrors.CodeInvalidArgument, "invalid input", http.StatusBadRequest, err))
		return
	}
	// TODO: implement manual upgrade logic
	_ = teacherID
	response.Success(c, nil)
}

func (h *Handler) CheckUpgrade(c *gin.Context) {
	teacherID, _ := strconv.ParseInt(c.Param("teacherId"), 10, 64)
	if teacherID == 0 {
		response.Fail(c, apperrors.New(apperrors.CodeInvalidArgument, "invalid teacher id", http.StatusBadRequest, nil))
		return
	}
	// TODO: implement upgrade condition check
	response.Success(c, gin.H{"canUpgrade": false, "reason": "not enough orders"})
}

func (h *Handler) GetUpgradeHistory(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	// Placeholder: return empty list (needs upgrade history table)
	response.Success(c, gin.H{"list": []any{}, "total": 0, "page": page, "pageSize": pageSize})
}

func (h *Handler) GetAssessmentVideo(c *gin.Context) {
	videoID, _ := strconv.ParseInt(c.Param("videoId"), 10, 64)
	if videoID == 0 {
		response.Fail(c, apperrors.New(apperrors.CodeInvalidArgument, "invalid video id", http.StatusBadRequest, nil))
		return
	}
	// Placeholder
	response.Success(c, gin.H{"id": videoID, "url": "", "status": "pending"})
}

func (h *Handler) UpdateAssessmentVideo(c *gin.Context) {
	videoID, _ := strconv.ParseInt(c.Param("videoId"), 10, 64)
	if videoID == 0 {
		response.Fail(c, apperrors.New(apperrors.CodeInvalidArgument, "invalid video id", http.StatusBadRequest, nil))
		return
	}
	var req struct {
		URL    string `json:"url"`
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, apperrors.New(apperrors.CodeInvalidArgument, "invalid input", http.StatusBadRequest, err))
		return
	}
	// TODO: implement video update
	_ = videoID
	response.Success(c, nil)
}

func (h *Handler) ToggleAssessmentVideo(c *gin.Context) {
	videoID, _ := strconv.ParseInt(c.Param("videoId"), 10, 64)
	if videoID == 0 {
		response.Fail(c, apperrors.New(apperrors.CodeInvalidArgument, "invalid video id", http.StatusBadRequest, nil))
		return
	}
	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, apperrors.New(apperrors.CodeInvalidArgument, "invalid input", http.StatusBadRequest, err))
		return
	}
	// TODO: implement video toggle
	_ = videoID
	response.Success(c, nil)
}

func (h *Handler) DeleteAssessmentVideo(c *gin.Context) {
	videoID, _ := strconv.ParseInt(c.Param("videoId"), 10, 64)
	if videoID == 0 {
		response.Fail(c, apperrors.New(apperrors.CodeInvalidArgument, "invalid video id", http.StatusBadRequest, nil))
		return
	}
	// TODO: implement video delete
	response.Success(c, nil)
}

func (h *Handler) ListAdminDynamics(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	// Placeholder: return empty list
	response.Success(c, gin.H{"list": []any{}, "total": 0, "page": page, "pageSize": pageSize})
}

func (h *Handler) DeleteDynamic(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if id == 0 {
		response.Fail(c, apperrors.New(apperrors.CodeInvalidArgument, "invalid dynamic id", http.StatusBadRequest, nil))
		return
	}
	// TODO: implement dynamic delete
	response.Success(c, nil)
}

func (h *Handler) BatchDeleteDynamics(c *gin.Context) {
	var ids []int64
	if err := c.ShouldBindJSON(&ids); err != nil {
		response.Fail(c, apperrors.New(apperrors.CodeInvalidArgument, "invalid input", http.StatusBadRequest, err))
		return
	}
	// TODO: implement batch dynamic delete
	response.Success(c, nil)
}




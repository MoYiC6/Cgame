package user

import (
	"net/http"
	"strconv"

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
	user := group.Group("/user")
	if h.authMiddleware != nil {
		user.Use(h.authMiddleware)
	}
	{
		user.GET("/center", h.GetUserCenter)
		user.PUT("/profile", h.UpdateProfile)
		user.GET("/balance", h.GetBalance)
		user.GET("/balance/logs", h.GetBalanceLogs)
		user.GET("/level", h.GetLevel)
		user.GET("/consumption-ranking", h.GetConsumptionRanking)
	}

	balance := group.Group("/balance")
	if h.authMiddleware != nil {
		balance.Use(h.authMiddleware)
	}
	{
		balance.GET("/my-balance", h.GetBalance)
		balance.GET("/my-logs", h.GetBalanceLogs)
	}

	client := group.Group("/client/user")
	if h.authMiddleware != nil {
		client.Use(h.authMiddleware)
	}
	{
		client.GET("/info", h.GetClientUserInfo)
		client.PUT("/info", h.UpdateClientUserInfo)
		client.POST("/complete-profile", h.CompleteProfile)
	}

	clientLevels := group.Group("/client/user-levels")
	{
		clientLevels.GET("", h.GetUserLevels)
	}

	admin := group.Group("/admin/users")
	if h.authMiddleware != nil {
		admin.Use(h.authMiddleware)
	}
	{
		admin.GET("", h.ListUsers)
		admin.GET("/:id", h.GetUserDetail)
		admin.PUT("/:id", h.UpdateUser)
		admin.PUT("/:id/status", h.UpdateUserStatus)
	}

	adminSelect := group.Group("/admin/select")
	if h.authMiddleware != nil {
		adminSelect.Use(h.authMiddleware)
	}
	{
		adminSelect.GET("/user", h.ListUserSelectors)
	}

	adminLogs := group.Group("/admin/logs/user")
	if h.authMiddleware != nil {
		adminLogs.Use(h.authMiddleware)
	}
	{
		adminLogs.GET("", h.ListUserLoginLogs)
		adminLogs.DELETE("/batch", h.DeleteUserLoginLogs)
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

func (h *Handler) GetUserCenter(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Fail(c, apperrors.New(apperrors.CodeForbidden, "unauthorized", http.StatusForbidden, nil))
		return
	}
	info, err := h.service.GetUserCenterInfo(c.Request.Context(), userID)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, info)
}

func (h *Handler) UpdateProfile(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Fail(c, apperrors.New(apperrors.CodeForbidden, "unauthorized", http.StatusForbidden, nil))
		return
	}
	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, apperrors.New(apperrors.CodeInvalidArgument, "invalid input", http.StatusBadRequest, err))
		return
	}
	if err := h.service.UpdateProfile(c.Request.Context(), userID, &req); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *Handler) GetBalance(c *gin.Context) {
	userID, _ := strconv.ParseInt(c.GetString("userID"), 10, 64)
	user, err := h.service.GetUser(c.Request.Context(), userID)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"balance": user.Balance})
}

func (h *Handler) GetBalanceLogs(c *gin.Context) {
	userID, _ := strconv.ParseInt(c.GetString("userID"), 10, 64)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	logs, total, err := h.service.GetBalanceLogs(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"list": logs, "total": total})
}

func (h *Handler) GetLevel(c *gin.Context) {
	userID, _ := strconv.ParseInt(c.GetString("userID"), 10, 64)
	level, err := h.service.GetUserLevel(c.Request.Context(), userID)
	if err != nil {
		response.Fail(c, err)
		return
	}
	if level == nil {
		response.Success(c, gin.H{"level": nil})
		return
	}
	response.Success(c, level)
}

func (h *Handler) GetClientUserInfo(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Fail(c, apperrors.New(apperrors.CodeForbidden, "unauthorized", http.StatusForbidden, nil))
		return
	}
	info, err := h.service.GetUserCenterInfo(c.Request.Context(), userID)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, info)
}

func (h *Handler) UpdateClientUserInfo(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Fail(c, apperrors.New(apperrors.CodeForbidden, "unauthorized", http.StatusForbidden, nil))
		return
	}
	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, apperrors.New(apperrors.CodeInvalidArgument, "invalid input", http.StatusBadRequest, err))
		return
	}
	if err := h.service.UpdateProfile(c.Request.Context(), userID, &req); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *Handler) CompleteProfile(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Fail(c, apperrors.New(apperrors.CodeForbidden, "unauthorized", http.StatusForbidden, nil))
		return
	}
	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, apperrors.New(apperrors.CodeInvalidArgument, "invalid input", http.StatusBadRequest, err))
		return
	}
	if err := h.service.UpdateProfile(c.Request.Context(), userID, &req); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *Handler) GetUserLevels(c *gin.Context) {
	levels, err := h.service.GetUserLevels(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, levels)
}

func (h *Handler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("pageNum", c.DefaultQuery("page", "1")))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	query := UserQuery{
		PageNum:  page,
		PageSize: pageSize,
		Username: c.Query("username"),
		Nickname: c.Query("nickname"),
		Mobile:   c.Query("mobile"),
		Email:    c.Query("email"),
	}
	if statusStr := c.Query("status"); statusStr != "" {
		if v, err := strconv.ParseInt(statusStr, 10, 16); err == nil {
			status := int16(v)
			query.Status = &status
		}
	}
	if isTeacherStr := c.Query("isTeacher"); isTeacherStr != "" {
		if v, err := strconv.ParseInt(isTeacherStr, 10, 16); err == nil {
			isTeacher := int16(v)
			query.IsTeacher = &isTeacher
		}
	}
	users, total, err := h.service.ListUsers(c.Request.Context(), query)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"list": users, "total": total})
}

func (h *Handler) GetUserDetail(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if id == 0 {
		response.Fail(c, apperrors.New(apperrors.CodeInvalidArgument, "invalid user id", http.StatusBadRequest, nil))
		return
	}
	user, err := h.service.GetUserByID(c.Request.Context(), id)
	if err != nil {
		response.Fail(c, err)
		return
	}
	if user == nil {
		response.Fail(c, apperrors.New(apperrors.CodeNotFound, "user not found", http.StatusNotFound, nil))
		return
	}
	response.Success(c, user)
}

func (h *Handler) UpdateUser(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if id == 0 {
		response.Fail(c, apperrors.New(apperrors.CodeInvalidArgument, "invalid user id", http.StatusBadRequest, nil))
		return
	}
	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, apperrors.New(apperrors.CodeInvalidArgument, "invalid input", http.StatusBadRequest, err))
		return
	}
	if err := h.service.UpdateProfile(c.Request.Context(), id, &req); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *Handler) UpdateUserStatus(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if id == 0 {
		response.Fail(c, apperrors.New(apperrors.CodeInvalidArgument, "invalid user id", http.StatusBadRequest, nil))
		return
	}
	var req UpdateUserStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, apperrors.New(apperrors.CodeInvalidArgument, "invalid input", http.StatusBadRequest, err))
		return
	}
	if err := h.service.UpdateUserStatus(c.Request.Context(), id, req.Status); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *Handler) GetConsumptionRanking(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	items, err := h.service.GetConsumptionRanking(c.Request.Context(), limit)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"list": items})
}

func (h *Handler) ListUserSelectors(c *gin.Context) {
	keyword := c.Query("keyword")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	items, err := h.service.ListUserSelectors(c.Request.Context(), keyword, limit)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"list": items})
}

func (h *Handler) ListUserLoginLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("pageNum", c.DefaultQuery("page", "1")))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	var userID *int64
	if uidStr := c.Query("userId"); uidStr != "" {
		if uid, err := strconv.ParseInt(uidStr, 10, 64); err == nil {
			userID = &uid
		}
	}
	logs, total, err := h.service.ListUserLoginLogs(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, gin.H{"list": logs, "total": total})
}

func (h *Handler) DeleteUserLoginLogs(c *gin.Context) {
	var ids []int64
	if err := c.ShouldBindJSON(&ids); err != nil {
		response.Fail(c, apperrors.New(apperrors.CodeInvalidArgument, "invalid input", http.StatusBadRequest, err))
		return
	}
	if err := h.service.DeleteUserLoginLogs(c.Request.Context(), ids); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil)
}

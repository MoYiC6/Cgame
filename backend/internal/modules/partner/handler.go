package partner

import (
	"strconv"

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
	// Partner config admin routes
	config := group.Group("/admin/partner-config")
	if h.authMiddleware != nil {
		config.Use(h.authMiddleware)
	}
	{
		config.GET("", h.ListPartnerConfigs)
		config.POST("", h.CreatePartnerConfig)
		config.PUT("/:id", h.UpdatePartnerConfig)
		config.DELETE("/:id", h.DeletePartnerConfig)
	}

	// Teacher partner admin routes
	partner := group.Group("/admin/teacher/partners")
	if h.authMiddleware != nil {
		partner.Use(h.authMiddleware)
	}
	{
		partner.GET("", h.ListTeacherPartners)
		partner.POST("", h.CreateTeacherPartner)
		partner.PUT("/:id", h.UpdateTeacherPartner)
		partner.DELETE("/:id", h.DeleteTeacherPartner)
		partner.GET("/teacher/:teacherId", h.ListByTeacher)
		partner.GET("/partnered-teachers", h.ListPartneredTeachers)
	}
}

func (h *Handler) ListPartnerConfigs(c *gin.Context) {
	query := PartnerConfigQuery{
		PageNum:     intQuery(c, "pageNum", 1),
		PageSize:    intQuery(c, "pageSize", 10),
		Name:        c.Query("name"),
		Status:      c.Query("status"),
		PartnerType: c.Query("partnerType"),
	}
	result, err := h.service.ListPartnerConfigs(c.Request.Context(), query)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, result)
}

func (h *Handler) CreatePartnerConfig(c *gin.Context) {
	var req PartnerConfigCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, err)
		return
	}
	id, err := h.service.CreatePartnerConfig(c.Request.Context(), req)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, id)
}

func (h *Handler) UpdatePartnerConfig(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, err)
		return
	}
	var req PartnerConfigUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, err)
		return
	}
	if err := h.service.UpdatePartnerConfig(c.Request.Context(), id, req); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *Handler) DeletePartnerConfig(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, err)
		return
	}
	if err := h.service.DeletePartnerConfig(c.Request.Context(), id); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *Handler) ListTeacherPartners(c *gin.Context) {
	query := TeacherPartnerQuery{
		PageNum:  intQuery(c, "pageNum", 1),
		PageSize: intQuery(c, "pageSize", 10),
		Status:   c.Query("status"),
	}
	if raw := c.Query("teacherId"); raw != "" {
		value, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			response.Fail(c, err)
			return
		}
		query.TeacherID = value
	}
	if raw := c.Query("partnerId"); raw != "" {
		value, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			response.Fail(c, err)
			return
		}
		query.PartnerID = value
	}
	result, err := h.service.ListTeacherPartners(c.Request.Context(), query)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, result)
}

func (h *Handler) CreateTeacherPartner(c *gin.Context) {
	var req TeacherPartnerCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, err)
		return
	}
	id, err := h.service.CreateTeacherPartner(c.Request.Context(), req)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, id)
}

func (h *Handler) UpdateTeacherPartner(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, err)
		return
	}
	var req TeacherPartnerUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, err)
		return
	}
	if err := h.service.UpdateTeacherPartner(c.Request.Context(), id, req); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *Handler) DeleteTeacherPartner(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, err)
		return
	}
	if err := h.service.DeleteTeacherPartner(c.Request.Context(), id); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *Handler) ListByTeacher(c *gin.Context) {
	teacherID, err := strconv.ParseInt(c.Param("teacherId"), 10, 64)
	if err != nil {
		response.Fail(c, err)
		return
	}
	query := TeacherPartnerQuery{
		PageNum:   intQuery(c, "pageNum", 1),
		PageSize:  intQuery(c, "pageSize", 10),
		TeacherID: teacherID,
		Status:    c.Query("status"),
	}
	result, err := h.service.ListTeacherPartners(c.Request.Context(), query)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, result)
}

func (h *Handler) ListPartneredTeachers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("pageNum", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	result, err := h.service.ListPartneredTeachers(c.Request.Context(), page, pageSize)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, result)
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

func intQuery(c *gin.Context, key string, fallback int) int {
	value, err := strconv.Atoi(c.DefaultQuery(key, strconv.Itoa(fallback)))
	if err != nil {
		return fallback
	}
	return value
}

package visitor

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
	common := group.Group("/common/visitor")
	{
		common.POST("/track", h.Track)
		common.POST("/batch", h.BatchTrack)
	}

	admin := group.Group("/admin/visitor-stats")
	if h.authMiddleware != nil {
		admin.Use(h.authMiddleware)
	}
	{
		admin.GET("/dashboard", h.Dashboard)
		admin.GET("/trend", h.Trend)
	}
}

// TrackRequest 访客追踪请求
// @Summary 访客追踪
// @Description 记录单次页面访问
// @Tags visitor
// @Accept json
// @Produce json
// @Param request body TrackVisitorRequest true "追踪请求"
// @Success 200 {object} TrackVisitorResponse
// @Router /api/common/visitor/track [post]
func (h *Handler) Track(c *gin.Context) {
	var req TrackVisitorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, apperrors.New(apperrors.CodeInvalidArgument, "invalid request", http.StatusBadRequest, err))
		return
	}
	if req.VisitorID == "" || req.SessionID == "" || req.Page == "" || req.Timestamp == 0 {
		response.Fail(c, apperrors.New(apperrors.CodeInvalidArgument, "missing required fields", http.StatusBadRequest, nil))
		return
	}
	resp, err := h.service.TrackVisitor(c.Request.Context(), &req)
	if err != nil {
		response.Fail(c, apperrors.New(apperrors.CodeInternal, err.Error(), http.StatusInternalServerError, err))
		return
	}
	response.Success(c, resp)
}

// BatchTrackRequest 批量访客追踪请求
// @Summary 批量访客追踪
// @Description 批量记录页面访问
// @Tags visitor
// @Accept json
// @Produce json
// @Param request body []TrackVisitorRequest true "追踪请求数组"
// @Success 200 {object} TrackVisitorResponse
// @Router /api/common/visitor/batch [post]
func (h *Handler) BatchTrack(c *gin.Context) {
	var reqs []*TrackVisitorRequest
	if err := c.ShouldBindJSON(&reqs); err != nil {
		response.Fail(c, apperrors.New(apperrors.CodeInvalidArgument, "invalid request", http.StatusBadRequest, err))
		return
	}
	if len(reqs) == 0 {
		response.Fail(c, apperrors.New(apperrors.CodeInvalidArgument, "empty batch", http.StatusBadRequest, nil))
		return
	}
	resp, err := h.service.BatchTrack(c.Request.Context(), reqs)
	if err != nil {
		response.Fail(c, apperrors.New(apperrors.CodeInternal, err.Error(), http.StatusInternalServerError, err))
		return
	}
	response.Success(c, resp)
}

// DashboardRequest 仪表盘请求
// @Summary 访客仪表盘
// @Description 获取今日/昨日访客统计及趋势
// @Tags visitor
// @Produce json
// @Success 200 {object} DashboardStats
// @Router /api/admin/visitor-stats/dashboard [get]
func (h *Handler) Dashboard(c *gin.Context) {
	stats, err := h.service.GetDashboardStats(c.Request.Context())
	if err != nil {
		response.Fail(c, apperrors.New(apperrors.CodeInternal, err.Error(), http.StatusInternalServerError, err))
		return
	}
	response.Success(c, stats)
}

// TrendRequest 趋势请求
// @Summary 访客趋势
// @Description 获取最近 N 天访客趋势
// @Tags visitor
// @Produce json
// @Param days query int false "天数" default(7) maximum(365)
// @Success 200 {array} TrendData
// @Router /api/admin/visitor-stats/trend [get]
func (h *Handler) Trend(c *gin.Context) {
	daysStr := c.DefaultQuery("days", "7")
	days, _ := strconv.Atoi(daysStr)
	if days <= 0 {
		days = 7
	}
	if days > 365 {
		days = 365
	}
	trend, err := h.service.GetTrend(c.Request.Context(), days)
	if err != nil {
		response.Fail(c, apperrors.New(apperrors.CodeInternal, err.Error(), http.StatusInternalServerError, err))
		return
	}
	response.Success(c, trend)
}

package coupon

import (
	"encoding/json"
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
	client := group.Group("/client/coupon")
	if h.authMiddleware != nil {
		client.Use(h.authMiddleware)
	}
	{
		client.GET("/available", h.ListAvailable)
		client.GET("/my", h.ListMine)
		client.POST("/claim/:id", h.Claim)
	}

	admin := group.Group("/admin/coupon")
	if h.authMiddleware != nil {
		admin.Use(h.authMiddleware)
	}
	{
		admin.GET("", h.ListAdmin)
		admin.POST("", h.Create)
		admin.PUT("/:id", h.Update)
		admin.DELETE("/:id", h.Delete)
		admin.GET("/stats", h.Stats)
	}
}

func (h *Handler) ListAvailable(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Fail(c, apperrors.New(apperrors.CodeForbidden, "unauthorized", http.StatusForbidden, nil))
		return
	}
	result, err := h.service.ListAvailable(c.Request.Context(), userID)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, result)
}

func (h *Handler) ListMine(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Fail(c, apperrors.New(apperrors.CodeForbidden, "unauthorized", http.StatusForbidden, nil))
		return
	}
	var status *int
	if raw := c.Query("status"); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil {
			response.Fail(c, err)
			return
		}
		status = &value
	}
	result, err := h.service.ListMine(c.Request.Context(), userID, status)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, result)
}

func (h *Handler) Claim(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		response.Fail(c, apperrors.New(apperrors.CodeForbidden, "unauthorized", http.StatusForbidden, nil))
		return
	}
	couponID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, err)
		return
	}
	userCouponID, err := h.service.Claim(c.Request.Context(), userID, couponID)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, userCouponID)
}

func (h *Handler) ListAdmin(c *gin.Context) {
	query := CouponQuery{
		PageNum:  intQuery(c, "pageNum", 1),
		PageSize: intQuery(c, "pageSize", 10),
		Name:     c.Query("name"),
		Status:   c.Query("status"),
	}
	if raw := c.Query("type"); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil {
			response.Fail(c, err)
			return
		}
		query.Type = &value
	}
	if raw := c.Query("enabled"); raw != "" {
		value, err := strconv.ParseBool(raw)
		if err != nil {
			response.Fail(c, err)
			return
		}
		query.Enabled = &value
	}
	if raw := c.Query("isPermanent"); raw != "" {
		value, err := strconv.ParseBool(raw)
		if err != nil {
			response.Fail(c, err)
			return
		}
		query.IsPermanent = &value
	}
	result, err := h.service.ListAdmin(c.Request.Context(), query)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, result)
}

func (h *Handler) Create(c *gin.Context) {
	var req CouponCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, err)
		return
	}
	id, err := h.service.Create(c.Request.Context(), req)
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, id)
}

func (h *Handler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, err)
		return
	}
	var req CouponUpdateRequest
	if err := decodeCouponUpdate(c, &req); err != nil {
		response.Fail(c, err)
		return
	}
	if err := h.service.Update(c.Request.Context(), id, req); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *Handler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, err)
		return
	}
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *Handler) Stats(c *gin.Context) {
	stats, err := h.service.Stats(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}
	response.Success(c, stats)
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

func decodeCouponUpdate(c *gin.Context, req *CouponUpdateRequest) error {
	var raw map[string]json.RawMessage
	if err := c.ShouldBindJSON(&raw); err != nil {
		return err
	}
	body, _ := json.Marshal(raw)
	if err := json.Unmarshal(body, req); err != nil {
		return err
	}
	if _, ok := raw["targetLevelIds"]; ok {
		req.TargetLevelIDsSet = true
	}
	if _, ok := raw["restrictedGoodsIds"]; ok {
		req.RestrictedGoodsIDsSet = true
	}
	if _, ok := raw["restrictedCategoryIds"]; ok {
		req.RestrictedCategorySet = true
	}
	return nil
}

func (r *CouponCreateRequest) UnmarshalJSON(data []byte) error {
	type alias CouponCreateRequest
	var presence map[string]json.RawMessage
	if err := json.Unmarshal(data, &presence); err != nil {
		return err
	}
	var raw struct {
		*alias
		StartTime *flexTime `json:"startTime"`
		EndTime   *flexTime `json:"endTime"`
	}
	raw.alias = (*alias)(r)
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if raw.StartTime != nil {
		r.StartTime = (*time.Time)(raw.StartTime)
	}
	if raw.EndTime != nil {
		r.EndTime = (*time.Time)(raw.EndTime)
	}
	if _, ok := presence["distributionMode"]; ok {
		r.DistributionModeSet = true
	}
	return nil
}

func (r *CouponUpdateRequest) UnmarshalJSON(data []byte) error {
	type alias CouponUpdateRequest
	var raw struct {
		*alias
		StartTime *flexTime `json:"startTime"`
		EndTime   *flexTime `json:"endTime"`
	}
	raw.alias = (*alias)(r)
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if raw.StartTime != nil {
		r.StartTime = (*time.Time)(raw.StartTime)
	}
	if raw.EndTime != nil {
		r.EndTime = (*time.Time)(raw.EndTime)
	}
	return nil
}

type flexTime time.Time

func (t *flexTime) UnmarshalJSON(data []byte) error {
	var raw string
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if raw == "" {
		return nil
	}
	layouts := []string{time.RFC3339, "2006-01-02T15:04:05", "2006-01-02 15:04:05"}
	var lastErr error
	for _, layout := range layouts {
		parsed, err := time.Parse(layout, raw)
		if err == nil {
			*t = flexTime(parsed)
			return nil
		}
		lastErr = err
	}
	return lastErr
}

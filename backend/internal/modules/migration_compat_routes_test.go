package modules_test

import (
	"net/http"
	"testing"

	"backend/internal/modules/auth"
	"backend/internal/modules/coupon"
	"backend/internal/modules/external"
	"backend/internal/modules/feedback"
	"backend/internal/modules/file"
	"backend/internal/modules/game"
	"backend/internal/modules/inventory"
	"backend/internal/modules/notification"
	"backend/internal/modules/order"
	"backend/internal/modules/payment"
	"backend/internal/modules/system"
	"backend/internal/modules/teacher"
	"backend/internal/modules/user"
	"github.com/gin-gonic/gin"
)

func TestJavaCompatibleRoutesAreRegisteredForMigratedSlices(t *testing.T) {
	gin.SetMode(gin.TestMode)

	engine := gin.New()
	api := engine.Group("/api")

	auth.NewHandler(nil, auth.HandlerConfig{}).RegisterRoutes(api)
	external.NewHandler(nil, nil).RegisterRoutes(api)
	order.NewHandler(nil).RegisterRoutes(api)
	payment.NewHandler(nil).RegisterRoutes(api)
	inventory.NewHandler(nil, nil).RegisterRoutes(api)
	file.NewHandler(nil, nil).RegisterRoutes(api)
	game.NewHandler(nil, nil).RegisterRoutes(api)
	notification.NewHandler(notification.NewService(notification.NewRepository(nil), nil), nil).RegisterRoutes(api)
	feedback.NewHandler(feedback.NewService(feedback.NewRepository(nil), nil), nil).RegisterRoutes(api)
	coupon.NewHandler(coupon.NewService(coupon.NewRepository(nil), nil), nil).RegisterRoutes(api)
	user.NewHandler(nil, nil).RegisterRoutes(api)
	teacher.NewHandler(nil, nil).RegisterRoutes(api)
	system.NewHandler(nil, nil).RegisterRoutes(api)

	registered := map[string]bool{}
	for _, route := range engine.Routes() {
		registered[route.Method+" "+route.Path] = true
	}

	expected := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/api/auth/info"},
		{http.MethodPost, "/api/client/orders"},
		{http.MethodGet, "/api/client/orders"},
		{http.MethodGet, "/api/client/orders/:orderId"},
		{http.MethodPost, "/api/client/orders/:orderId/cancel"},
		{http.MethodPost, "/api/client/orders/:orderId/confirm"},
		{http.MethodPost, "/api/client/payments"},
		{http.MethodPost, "/api/client/payments/confirm"},
		{http.MethodGet, "/api/client/payments/status"},
		{http.MethodGet, "/api/admin/payments"},
		{http.MethodGet, "/api/admin/payments/stats"},
		{http.MethodGet, "/api/client/goods"},
		{http.MethodGet, "/api/client/goods/:id"},
		{http.MethodGet, "/api/client/goods/detail/:goodsId"},
		{http.MethodPost, "/api/client/goods/sku/check"},
		{http.MethodGet, "/api/client/categories"},
		{http.MethodPost, "/api/upload/file"},
		{http.MethodPost, "/api/upload/base64"},
		{http.MethodGet, "/api/upload/check-hash"},
		{http.MethodGet, "/api/upload/token"},
		{http.MethodPost, "/api/upload/confirm"},
		{http.MethodGet, "/api/client/game-map/list"},
		{http.MethodGet, "/api/client/game-map/goods/:goodsId"},
		{http.MethodGet, "/api/admin/game-map"},
		{http.MethodGet, "/api/admin/game-map/:id"},
		{http.MethodPost, "/api/admin/game-map"},
		{http.MethodPut, "/api/admin/game-map/:id"},
		{http.MethodDelete, "/api/admin/game-map/:id"},
		{http.MethodGet, "/api/admin/game-map/enabled"},
		{http.MethodGet, "/api/admin/bomb-ranking"},
		{http.MethodGet, "/api/balance/my-balance"},
		{http.MethodGet, "/api/balance/my-logs"},
		{http.MethodGet, "/api/client/teachers"},
		{http.MethodGet, "/api/client/teacher/levels"},
		{http.MethodGet, "/api/admin/teacher/levels"},
		{http.MethodGet, "/api/admin/partner-config"},
		{http.MethodPut, "/api/admin/partner-config"},
		{http.MethodGet, "/api/admin/wxpay/config/page"},
		{http.MethodGet, "/api/admin/wxpay/config/:id"},
		{http.MethodPost, "/api/admin/wxpay/config"},
		{http.MethodPut, "/api/admin/wxpay/config/:id"},
		{http.MethodDelete, "/api/admin/wxpay/config/:id"},
		{http.MethodPut, "/api/admin/wxpay/config/:id/status"},
		{http.MethodGet, "/api/admin/wxpay/config/type/:configType"},
		{http.MethodGet, "/api/common/customer-service/config"},
		{http.MethodGet, "/api/admin/notification-inbox"},
		{http.MethodPut, "/api/admin/notification-inbox/:id/read"},
		{http.MethodPut, "/api/admin/notification-inbox/read-all"},
		{http.MethodPost, "/api/client/feedback/submit"},
		{http.MethodGet, "/api/client/feedback/list"},
		{http.MethodGet, "/api/client/feedback/:id"},
		{http.MethodGet, "/api/admin/feedback"},
		{http.MethodGet, "/api/admin/feedback/list"},
		{http.MethodGet, "/api/admin/feedback/:id"},
		{http.MethodPost, "/api/admin/feedback/reply"},
		{http.MethodPost, "/api/admin/feedback/:id/reply"},
		{http.MethodPut, "/api/admin/feedback/status"},
		{http.MethodPut, "/api/admin/feedback/:id/status"},
		{http.MethodDelete, "/api/admin/feedback/:id"},
		{http.MethodGet, "/api/client/coupon/available"},
		{http.MethodGet, "/api/client/coupon/my"},
		{http.MethodPost, "/api/client/coupon/claim/:id"},
		{http.MethodGet, "/api/admin/coupon"},
		{http.MethodPost, "/api/admin/coupon"},
		{http.MethodPut, "/api/admin/coupon/:id"},
		{http.MethodDelete, "/api/admin/coupon/:id"},
		{http.MethodGet, "/api/admin/coupon/stats"},
	}

	for _, route := range expected {
		key := route.method + " " + route.path
		if !registered[key] {
			t.Fatalf("expected Java-compatible route %s to be registered; registered routes: %#v", key, registered)
		}
	}
}

package modules_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"backend/internal/modules/inventory"
	"backend/internal/modules/notification"
	"backend/internal/modules/order"
	"backend/internal/modules/payment"
	"backend/internal/platform/database"
	"backend/internal/platform/observability"
	"backend/internal/platform/response"
	"github.com/gin-gonic/gin"
)

type routeRegistrar interface {
	RegisterRoutes(group *gin.RouterGroup)
}

func TestModulePingHandlersReturnTopLevelTraceIDWithoutPayloadTrace(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name          string
		path          string
		module        string
		newRepository func() any
		newService    func(repo any) any
		newHandler    func(service any) routeRegistrar
	}{
		{
			name:          "order",
			path:          "/api/v1/order/ping",
			module:        "order",
			newRepository: func() any { return order.NewRepository() },
			newService:    func(repo any) any { return order.NewService(repo.(order.Repository), database.NoopTxManager{}) },
			newHandler:    func(service any) routeRegistrar { return order.NewHandler(service.(order.Service)) },
		},
		{
			name:          "payment",
			path:          "/api/v1/payment/ping",
			module:        "payment",
			newRepository: func() any { return payment.NewRepository() },
			newService:    func(repo any) any { return payment.NewService(repo.(payment.Repository), database.NoopTxManager{}) },
			newHandler:    func(service any) routeRegistrar { return payment.NewHandler(service.(payment.Service)) },
		},
		{
			name:          "inventory",
			path:          "/api/v1/inventory/ping",
			module:        "inventory",
			newRepository: func() any { return inventory.NewRepository() },
			newService:    func(repo any) any { return inventory.NewService(repo.(inventory.Repository), database.NoopTxManager{}) },
			newHandler:    func(service any) routeRegistrar { return inventory.NewHandler(service.(inventory.Service)) },
		},
		{
			name:          "notification",
			path:          "/api/v1/notification/ping",
			module:        "notification",
			newRepository: func() any { return notification.NewRepository(nil) },
			newService: func(repo any) any {
				return notification.NewService(repo.(notification.Repository), database.NoopTxManager{})
			},
			newHandler:    func(service any) routeRegistrar { return notification.NewHandler(service.(*notification.Service), nil) },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := tt.newRepository()
			service := tt.newService(repo)
			handler := tt.newHandler(service)

			engine := gin.New()
			api := engine.Group("/api/v1")
			handler.RegisterRoutes(api)

			request := httptest.NewRequest(http.MethodGet, tt.path, nil)
			request = request.WithContext(observability.WithTraceID(context.Background(), "trace-from-context"))
			recorder := httptest.NewRecorder()

			engine.ServeHTTP(recorder, request)

			if recorder.Code != http.StatusOK {
				t.Fatalf("expected 200, got %d", recorder.Code)
			}

			var body response.APIResponse
			if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
				t.Fatalf("json.Unmarshal returned error: %v", err)
			}

			if body.TraceID != "trace-from-context" {
				t.Fatalf("expected top-level trace_id to be populated, got %q", body.TraceID)
			}

			data, ok := body.Data.(map[string]any)
			if !ok {
				t.Fatalf("expected body data to decode into map, got %#v", body.Data)
			}

			if data["module"] != tt.module {
				t.Fatalf("expected module %q, got %#v", tt.module, data["module"])
			}

			if _, exists := data["trace_id"]; exists {
				t.Fatalf("expected payload trace_id to be omitted, got %#v", data["trace_id"])
			}
		})
	}
}

func TestModuleConstructorsExposeSymmetricServiceAndRepositorySeams(t *testing.T) {
	tests := []struct {
		name           string
		service        any
		nilTxService   any
		repository     any
		repositoryType reflect.Type
	}{
		{
			name:           "order",
			service:        order.NewService(order.NewRepository(), database.NoopTxManager{}),
			nilTxService:   order.NewService(order.NewRepository(), nil),
			repository:     order.NewRepository(),
			repositoryType: reflect.TypeFor[order.Repository](),
		},
		{
			name:           "payment",
			service:        payment.NewService(payment.NewRepository(), database.NoopTxManager{}),
			nilTxService:   payment.NewService(payment.NewRepository(), nil),
			repository:     payment.NewRepository(),
			repositoryType: reflect.TypeFor[payment.Repository](),
		},
		{
			name:           "inventory",
			service:        inventory.NewService(inventory.NewRepository(), database.NoopTxManager{}),
			nilTxService:   inventory.NewService(inventory.NewRepository(), nil),
			repository:     inventory.NewRepository(),
			repositoryType: reflect.TypeFor[inventory.Repository](),
		},
		{
			name:           "notification",
			service:        notification.NewService(notification.NewRepository(nil), database.NoopTxManager{}),
			nilTxService:   notification.NewService(notification.NewRepository(nil), nil),
			repository:     notification.NewRepository(nil),
			repositoryType: reflect.TypeFor[notification.Repository](),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serviceType := reflect.TypeOf(tt.service)
			if serviceType.Kind() != reflect.Pointer {
				t.Fatalf("expected pointer service type, got %s", serviceType.Kind())
			}

			serviceElem := serviceType.Elem()
			repoField, ok := serviceElem.FieldByName("repo")
			if !ok {
				t.Fatalf("expected service struct to expose repo field")
			}
			if !repoField.Type.Implements(tt.repositoryType) {
				t.Fatalf("expected repo field to implement %s, got %s", tt.repositoryType, repoField.Type)
			}

			txManagerField, ok := serviceElem.FieldByName("txManager")
			if !ok {
				t.Fatalf("expected service struct to expose txManager field")
			}
			if !txManagerField.Type.Implements(reflect.TypeFor[database.TxManager]()) {
				t.Fatalf("expected txManager field to implement database.TxManager, got %s", txManagerField.Type)
			}

			nilTxManagerField := reflect.ValueOf(tt.nilTxService).Elem().FieldByName("txManager")
			if nilTxManagerField.IsNil() {
				t.Fatalf("expected nil txManager constructor input to fall back to database.NoopTxManager")
			}

			repoType := reflect.TypeOf(tt.repository)
			if repoType.Kind() == reflect.Interface {
				repoType = reflect.TypeOf(reflect.ValueOf(tt.repository).Interface())
			}
			if repoType.Kind() != reflect.Pointer {
				t.Fatalf("expected repository concrete type to be a pointer, got %s", repoType.Kind())
			}

			repoElem := repoType.Elem()
			dbtxField, ok := repoElem.FieldByName("dbtx")
			if !ok {
				t.Fatalf("expected repository struct to expose dbtx field")
			}
			if dbtxField.Type != reflect.TypeFor[database.DBTX]() {
				t.Fatalf("expected dbtx field type %s, got %s", reflect.TypeFor[database.DBTX](), dbtxField.Type)
			}
		})
	}
}

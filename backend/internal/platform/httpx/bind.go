package httpx

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"reflect"

	apperrors "backend/internal/platform/errors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

const invalidArgumentMessage = "invalid argument"

// BindAndValidate binds uri, query, header, and JSON body in a fixed order,
// then validates the final DTO once to avoid repeated source-level validation.
func BindAndValidate[T any](c *gin.Context) (*T, error) {
	payload := new(T)

	if err := bindURI(payload, c.Params); err != nil {
		return nil, invalidArgumentError(err)
	}
	if err := binding.MapFormWithTag(payload, c.Request.URL.Query(), "form"); err != nil {
		return nil, invalidArgumentError(err)
	}
	if err := bindHeader(payload, c.Request.Header); err != nil {
		return nil, invalidArgumentError(err)
	}
	if err := bindJSON(payload, c.Request); err != nil {
		return nil, invalidArgumentError(err)
	}
	if err := binding.Validator.ValidateStruct(payload); err != nil {
		return nil, invalidArgumentError(err)
	}

	return payload, nil
}

func bindURI(target any, params gin.Params) error {
	values := make(map[string][]string, len(params))
	for _, param := range params {
		values[param.Key] = []string{param.Value}
	}

	return binding.MapFormWithTag(target, values, "uri")
}

func bindJSON(target any, request *http.Request) error {
	if request == nil || request.Body == nil {
		return nil
	}
	if request.ContentLength == 0 {
		return nil
	}

	decoder := json.NewDecoder(request.Body)
	if err := decoder.Decode(target); err != nil {
		return err
	}

	request.Body = http.NoBody
	request.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(nil)), nil
	}

	return nil
}

func bindHeader(target any, header http.Header) error {
	values := make(map[string][]string)
	collectHeaderValues(reflect.TypeOf(target), header, values)
	return binding.MapFormWithTag(target, values, "header")
}

func collectHeaderValues(targetType reflect.Type, header http.Header, values map[string][]string) {
	if targetType == nil {
		return
	}
	if targetType.Kind() == reflect.Pointer {
		targetType = targetType.Elem()
	}
	if targetType.Kind() != reflect.Struct {
		return
	}

	for i := range targetType.NumField() {
		field := targetType.Field(i)
		if field.PkgPath != "" && !field.Anonymous {
			continue
		}

		if tag := field.Tag.Get("header"); tag != "" && tag != "-" {
			if entries := header.Values(tag); len(entries) > 0 {
				values[tag] = entries
			}
		}

		fieldType := field.Type
		if fieldType.Kind() == reflect.Pointer {
			fieldType = fieldType.Elem()
		}
		if fieldType.Kind() == reflect.Struct {
			collectHeaderValues(fieldType, header, values)
		}
	}
}

func invalidArgumentError(cause error) *apperrors.AppError {
	return apperrors.New("INVALID_ARGUMENT", invalidArgumentMessage, http.StatusBadRequest, cause)
}

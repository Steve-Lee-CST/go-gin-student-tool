package gin_tool

import (
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
)

/*
HttpRequest:
    1. Most request info can be get from HttpRequest.
    2. Use DecodeRequest before binding request body.
*/
// HttpRequest is a struct to store request info.
type HttpRequest struct {
	Method   string `json:"method"`
	Protocol string `json:"protocol"`
	Host     string `json:"host"`
	URL      string `json:"url"`
	Path     string `json:"path"`
	FullPath string `json:"full_path"`
	IP       string `json:"ip"`

	Header    http.Header    `json:"headers"`
	Cookies   []*http.Cookie `json:"cookies"`
	UserAgent string         `json:"user_agent"`

	// Path params
	PathParams gin.Params `json:"path_params,omitempty"`
	// GET query params
	QueryParams url.Values `json:"query_params,omitempty"`

	// POST content-type
	ContentType string `json:"content_type"`
	// POST body
	RawBody  []byte         `json:"raw_body,omitempty"`
	JsonBody map[string]any `json:"json_body,omitempty"`
	// POST form data
	FormValues url.Values `json:"form_values,omitempty"`
	// POST form files
	FormFiles []string `json:"form_files,omitempty"`

	// time
	ReceivedTime time.Time `json:"received_time"`
	RequestID    string    `json:"request_id,omitempty"`
}

/*
HttpResponse:
    1. Most response info can be get from HttpResponse.
	2. Set c.Writer to IWrappedResponseWriter to copy response body before writing response.
*/
// HttpResponse is a struct to store response info.
type HttpResponse struct {
	// response info
	Status      int    `json:"status"`
	StatusText  string `json:"status_text"`
	Size        int    `json:"size"`
	ContentType string `json:"content_type"`

	// Header And Cookies
	Header  http.Header    `json:"headers,omitempty"`
	Cookies []*http.Cookie `json:"cookies,omitempty"`

	// response body
	Body     []byte         `json:"body,omitempty"`
	JsonBody map[string]any `json:"json_body,omitempty"`

	// error info
	Error      string `json:"error,omitempty"`
	StackTrace string `json:"stack_trace,omitempty"`

	ResponseTime time.Time `json:"response_time"`
	RequestID    string    `json:"request_id,omitempty"`
}

type CommonResponse[T any] struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    *T     `json:"data,omitempty"`
}

func SuccessResponse[T any](data *T) CommonResponse[T] {
	return CommonResponse[T]{
		Code:    0,
		Message: "success",
		Data:    data,
	}
}

func ErrorResponse[T any](code int, message string, data *T) CommonResponse[T] {
	return CommonResponse[T]{
		Code:    code,
		Message: message,
		Data:    data,
	}
}

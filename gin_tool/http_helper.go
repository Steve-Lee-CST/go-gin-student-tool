package gin_tool

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	HttpRequestKey        = "http_request"
	HttpResponseKey       = "http_response"
	HttpResponseWriterKey = "_http_response_writer"
)

// Check if WrappedResponseWriter implements IWrappedResponseWriter
var (
	_ IWrappedResponseWriter = &WrappedResponseWriter{}
)

/*
IWrappedResponseWriter And WrappedResponseWriter:
    1. IWrappedResponseWriter is a interface to wrap gin.ResponseWriter,
        which can be used to get response body.
    2. WrappedResponseWriter is a default implementation of IWrappedResponseWriter,
        it will copy response body to a bytes.Buffer.
*/
// IWrappedResponseWriter is an interface that extends gin.ResponseWriter
type IWrappedResponseWriter interface {
	gin.ResponseWriter
	GetBodyBytes() []byte
}

type WrappedResponseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func NewWrappedResponseWriter(w gin.ResponseWriter) *WrappedResponseWriter {
	return &WrappedResponseWriter{
		ResponseWriter: w,
		body:           &bytes.Buffer{},
	}
}

func (r *WrappedResponseWriter) GetBodyBytes() []byte {
	return r.body.Bytes()
}

func (r *WrappedResponseWriter) Write(b []byte) (int, error) {
	r.body.Write(b)                  // capture response body
	return r.ResponseWriter.Write(b) // write to original response
}

type HttpHelper struct{}

// DecodeRequest decodes the HTTP request data from the gin context.
func (h HttpHelper) DecodeRequest(c *gin.Context) *HttpRequest {
	req := HttpRequest{
		Method:   c.Request.Method,
		Protocol: c.Request.Proto,
		Host:     c.Request.Host,
		URL:      c.Request.URL.String(),
		Path:     c.Request.URL.Path,
		FullPath: c.FullPath(),
		IP:       c.ClientIP(),

		Header:    c.Request.Header.Clone(),
		Cookies:   c.Request.Cookies(),
		UserAgent: c.Request.UserAgent(),

		PathParams:  c.Params,
		QueryParams: c.Request.URL.Query(),
		ContentType: c.ContentType(),

		ReceivedTime: time.Now(),
	}

	// read raw body and write back
	body, _ := c.GetRawData()
	if len(body) > 0 {
		req.RawBody = body
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
	}

	// parse json
	if strings.Contains(req.ContentType, "application/json") && len(body) > 0 {
		_ = json.Unmarshal(body, &req.JsonBody)
	}

	// parse form data
	if strings.Contains(req.ContentType, "multipart/form-data") ||
		strings.Contains(req.ContentType, "application/x-www-form-urlencoded") {
		if err := c.Request.ParseMultipartForm(32 << 20); err == nil {
			req.FormValues = c.Request.PostForm
			if c.Request.MultipartForm != nil {
				for _, files := range c.Request.MultipartForm.File {
					for _, file := range files {
						req.FormFiles = append(req.FormFiles, file.Filename)
					}
				}
			}
		}
	}
	// request ID
	if requestID, exists := GetRequestID(c); exists {
		req.RequestID = requestID
	}

	return &req
}

// DecodeResponse decodes the HTTP response data
// from the gin context and the wrapped response writer.
func (h HttpHelper) DecodeResponse(
	c *gin.Context, writer IWrappedResponseWriter,
) *HttpResponse {
	resp := HttpResponse{
		Status:      c.Writer.Status(),
		StatusText:  http.StatusText(c.Writer.Status()),
		Size:        c.Writer.Size(),
		ContentType: c.Writer.Header().Get("Content-Type"),

		ResponseTime: time.Now(),
	}

	resp.Header = c.Writer.Header()
	resp.Cookies = c.Request.Cookies()
	resp.Body = writer.GetBodyBytes()

	if strings.Contains(resp.ContentType, "application/json") {
		json.Unmarshal(writer.GetBodyBytes(), &resp.JsonBody)
	}

	if requestID, exists := GetRequestID(c); exists {
		resp.RequestID = requestID
	}

	return &resp
}

func (h HttpHelper) SetHttpRequest(c *gin.Context, req *HttpRequest) {
	if req == nil {
		return
	}
	c.Set(HttpRequestKey, req)
}

func (h HttpHelper) SetHttpResponse(c *gin.Context, resp *HttpResponse) {
	if resp == nil {
		return
	}
	c.Set(HttpResponseKey, resp)
}

func (h HttpHelper) SetHttpResponseWriter(c *gin.Context, writer IWrappedResponseWriter) IWrappedResponseWriter {
	if writer == nil {
		return nil
	}
	c.Set(HttpResponseWriterKey, writer)
	c.Writer = writer
	return writer
}

func (h HttpHelper) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Decode request
		httpRequest := h.DecodeRequest(c)
		h.SetHttpRequest(c, httpRequest)

		// Set response writer
		writer := h.SetHttpResponseWriter(c, NewWrappedResponseWriter(c.Writer))

		// Proceed with the request
		c.Next()

		// Decode response
		httpResponse := h.DecodeResponse(c, writer)

		// Set response in context (if needed)
		h.SetHttpResponse(c, httpResponse)
	}
}

func GetHttpRequest(c *gin.Context) (*HttpRequest, bool) {
	httpRequestRaw, ok := c.Get(HttpRequestKey)
	if !ok {
		return nil, false
	}
	httpRequest, ok := httpRequestRaw.(*HttpRequest)
	if !ok {
		return nil, false
	}
	return httpRequest, true
}

func GetHttpResponse(c *gin.Context) (*HttpResponse, bool) {
	httpResponseRaw, ok := c.Get(HttpResponseKey)
	if !ok {
		return nil, false
	}
	httpResponse, ok := httpResponseRaw.(*HttpResponse)
	if !ok {
		return nil, false
	}
	return httpResponse, true
}

func GetHttpResponseWriter(c *gin.Context) (IWrappedResponseWriter, bool) {
	writerRaw, ok := c.Get(HttpResponseWriterKey)
	if !ok {
		return nil, false
	}
	writer, ok := writerRaw.(IWrappedResponseWriter)
	if !ok {
		return nil, false
	}
	return writer, true
}

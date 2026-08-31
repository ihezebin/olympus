package middleware

import (
	"bytes"
	"fmt"
	"io"
	"mime"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/ihezebin/olympus/logger"
)

const maxBodyLen = 1024 * 2

func LoggingRequest() gin.HandlerFunc {
	return generateLoggingRequest(true)
}

func LoggingRequestWithoutHeader() gin.HandlerFunc {
	return generateLoggingRequest(false)
}

func generateLoggingRequest(header bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		fields := map[string]interface{}{
			"method": c.Request.Method,
			"uri":    c.Request.URL.RequestURI(),
			"remote": c.Request.RemoteAddr,
			"body":   requestBody(c),
		}
		if header {
			fields["header"] = c.Request.Header
		}
		logger.WithFields(fields).Info(ctx, "incoming http request")
		c.Next()
	}
}

func requestBody(c *gin.Context) string {
	if c.Request.Body == nil || c.Request.Body == http.NoBody {
		return ""
	}
	// multipart / 音视频等大 body 不 ReadAll，只记 content-length，避免占内存、拖慢请求
	if skipRequestBodyRead(c.Request.Header.Get("Content-Type")) {
		return fmt.Sprintf("[skipped body, content-length=%d]", c.Request.ContentLength)
	}

	bodyData, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return fmt.Sprintf("read request body err: %s", err.Error())
	}
	_ = c.Request.Body.Close()
	c.Request.Body = io.NopCloser(bytes.NewReader(bodyData))

	bodySize := len(bodyData)
	bodySuffix := ""
	if bodySize > maxBodyLen {
		bodySize = maxBodyLen
		bodySuffix = "..."
	}
	return string(bodyData[:bodySize]) + bodySuffix
}

func skipRequestBodyRead(contentType string) bool {
	mediaType, _, _ := mime.ParseMediaType(contentType)
	switch mediaType {
	case "multipart/form-data", "multipart/mixed", "application/octet-stream":
		return true
	}
	if strings.HasPrefix(mediaType, "audio/") ||
		strings.HasPrefix(mediaType, "video/") ||
		strings.HasPrefix(mediaType, "image/") {
		return true
	}
	return false
}

func LoggingResponse() gin.HandlerFunc {
	return generateLoggingResponse(true)
}

func LoggingResponseWithoutHeader() gin.HandlerFunc {
	return generateLoggingResponse(false)
}

func generateLoggingResponse(header bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		rw := &responseWriter{Body: new(bytes.Buffer), ResponseWriter: c.Writer}
		c.Writer = rw
		c.Next()

		fields := map[string]interface{}{
			"status": fmt.Sprintf("%v %s", c.Writer.Status(), http.StatusText(c.Writer.Status())),
			"body":   responseBody(rw),
		}
		if header {
			fields["header"] = c.Writer.Header()
		}

		logger.WithFields(fields).Info(ctx, "outgoing http response")
	}
}

func responseBody(rw *responseWriter) string {
	// multipart / 音视频等媒体响应不记 body 内容，只记 content-length，与请求日志一致
	if skipRequestBodyRead(rw.Header().Get("Content-Type")) {
		return fmt.Sprintf("[skipped body, content-length=%d]", rw.written)
	}

	body := rw.Body.Bytes()
	bodySize := len(body)
	bodySuffix := ""
	if bodySize > maxBodyLen {
		bodySize = maxBodyLen
		bodySuffix = "..."
	}
	return string(body[:bodySize]) + bodySuffix
}

type responseWriter struct {
	gin.ResponseWriter
	Body     *bytes.Buffer
	written  int
	skipBody bool
}

func (w *responseWriter) Write(body []byte) (int, error) {
	w.checkSkipBody()
	n, err := w.ResponseWriter.Write(body)
	if n > 0 {
		w.written += n
	}
	// Write 后 Content-Type 可能被底层 sniff 填充，再检查一次
	if !w.skipBody {
		w.checkSkipBody()
	}
	if !w.skipBody && n > 0 {
		w.Body.Write(body[:n])
	}
	return n, err
}

func (w *responseWriter) WriteString(s string) (int, error) {
	w.checkSkipBody()
	n, err := w.ResponseWriter.WriteString(s)
	if n > 0 {
		w.written += n
	}
	if !w.skipBody {
		w.checkSkipBody()
	}
	if !w.skipBody && n > 0 {
		w.Body.WriteString(s[:n])
	}
	return n, err
}

func (w *responseWriter) checkSkipBody() {
	if w.skipBody {
		return
	}
	if skipRequestBodyRead(w.Header().Get("Content-Type")) {
		w.skipBody = true
		w.Body.Reset()
	}
}

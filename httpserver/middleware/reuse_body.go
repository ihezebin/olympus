package middleware

import (
	"bytes"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

// reusableBody 支持在同一请求链内多次完整读取 body：
// 一次读到 EOF 后，下一次 Read 会自动从头开始。
type reusableBody struct {
	data []byte
	r    *bytes.Reader
	eof  bool
}

func newReusableBody(data []byte) *reusableBody {
	return &reusableBody{
		data: data,
		r:    bytes.NewReader(data),
	}
}

func (b *reusableBody) Read(p []byte) (int, error) {
	if b.eof {
		b.r.Reset(b.data)
		b.eof = false
	}
	n, err := b.r.Read(p)
	if err == io.EOF {
		b.eof = true
	}
	return n, err
}

func (b *reusableBody) Close() error {
	b.r.Reset(b.data)
	b.eof = false
	return nil
}

func ReuseBody() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Body == nil || c.Request.Body == http.NoBody {
			c.Next()
			return
		}

		data, err := io.ReadAll(c.Request.Body)
		_ = c.Request.Body.Close()
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		body := newReusableBody(data)
		c.Request.Body = body
		c.Request.GetBody = func() (io.ReadCloser, error) {
			return newReusableBody(data), nil
		}
		c.Next()
		c.Request.Body = newReusableBody(data)
	}
}

package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"

	"github.com/ihezebin/olympus/logger"
)

// codeInternalServerError 与 httpserver.CodeInternalServerError 保持一致（避免 middleware → httpserver 循环依赖）
const codeInternalServerError = 2

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				ctx := c.Request.Context()
				logger.Errorf(ctx, "panic: %v", err)
				logger.Errorf(ctx, "stack: %s", debug.Stack())

				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"code":    codeInternalServerError,
					"message": "Internal Server Error",
				})
			}
		}()
		c.Next()
	}
}

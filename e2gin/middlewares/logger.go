package middlewares

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/sirupsen/logrus"
)

type Logger struct {
	gin.ResponseWriter
	body bytes.Buffer
}

func (g *Logger) Write(b []byte) (int, error) {
	g.body.Write(b)
	return g.ResponseWriter.Write(b)
}

// RequestLoggingMiddleware /*
func RequestLoggingMiddleware(logger *logrus.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ginBodyLogger := &Logger{
			body:           bytes.Buffer{},
			ResponseWriter: ctx.Writer,
		}
		ctx.Writer = ginBodyLogger
		var req interface{}
		if err := ctx.ShouldBindBodyWith(&req, binding.JSON); err != nil {
			ctx.JSON(http.StatusBadRequest, err.Error())
			return
		}
		data, err := json.Marshal(req)
		if err != nil {
			panic(fmt.Errorf("err while marshaling req msg: %w", err))
		}
		ctx.Next()
		logger.WithFields(logrus.Fields{
			"status":       ctx.Writer.Status(),
			"method":       ctx.Request.Method,
			"path":         ctx.Request.URL.Path,
			"query_params": ctx.Request.URL.Query(),
			"req_body":     string(data),
			"res_body":     ginBodyLogger.body.String(),
		}).Info("request details")
	}
}

func SliceLoggerMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Next()
	}
}

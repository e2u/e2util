package resp

import (
	"fmt"
	"maps"
	"net/http"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const (
	ConsumerHeader         = "X-Api-Consumer"
	ConsumerTypeRADataJSON = "ra-data-json" // ra-data-json-server in react-admin
)

type StatusMessage struct {
	HttpCode int    `json:"-"`
	Code     int    `json:"code"`
	Message  string `json:"message,omitempty"`
	Details  any    `json:"detail,omitempty"`
}

const (
	Success int = 10000 + iota
	Accepted
	Created
	NoContent
)

const (
	BadRequest int = 90000 + iota
	Unauthorized
	Forbidden
	NotFound
	ServerError
	MethodNotAllowed
	NotAcceptable
	NotImplemented
	BadGateway
)

var (
	undefinedError = StatusMessage{HttpCode: http.StatusForbidden, Message: "forbidden"}
	codeMessageMap = map[int]StatusMessage{
		Success:          {HttpCode: http.StatusOK, Message: "success"},
		Accepted:         {HttpCode: http.StatusAccepted, Message: "accepted"},
		Created:          {HttpCode: http.StatusCreated, Message: "created"},
		NoContent:        {HttpCode: http.StatusNoContent, Message: "no_content"},
		BadRequest:       {HttpCode: http.StatusBadRequest, Message: "bad_request"},
		Unauthorized:     {HttpCode: http.StatusUnauthorized, Message: "unauthorized"},
		NotFound:         {HttpCode: http.StatusNotFound, Message: "not_found"},
		ServerError:      {HttpCode: http.StatusInternalServerError, Message: "internal_server_error"},
		Forbidden:        undefinedError,
		MethodNotAllowed: {HttpCode: http.StatusMethodNotAllowed, Message: "method_not_allowed"},
		NotAcceptable:    {HttpCode: http.StatusNotAcceptable, Message: "not_acceptable"},
		NotImplemented:   {HttpCode: http.StatusNotImplemented, Message: "not_implemented"},
		BadGateway:       {HttpCode: http.StatusBadGateway, Message: "bad_gateway"},
	}
)

func getCodeMessage(code int) StatusMessage {
	v, ok := codeMessageMap[code]
	if !ok {
		v = undefinedError
	}
	v.Code = code
	return v
}

type ErrorResponse struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
	Detail  string `json:"detail,omitempty"`
}

func AboutWithJSON(c *gin.Context, code int, detail any) {
	if raDataJSON(c, code, detail) {
		c.Abort()
		return
	}
	v := getCodeMessage(code)
	rv := gin.H{
		"code":    v.Code,
		"message": v.Message,
		"detail":  detail,
	}
	c.AbortWithStatusJSON(v.HttpCode, rv)
}

func SuccessWithJSON(c *gin.Context, code int, data any) {
	if raDataJSON(c, code, data) {
		return
	}
	v := getCodeMessage(code)
	r := gin.H{
		"code":    v.Code,
		"message": v.Message,
	}
	if data == nil {
		c.JSON(v.HttpCode, r)
		return
	}
	switch tv := data.(type) {
	case gin.H:
		maps.Copy(r, tv)
	default:
		maps.Copy(r, gin.H{"data": data})
	}
	c.JSON(v.HttpCode, r)
}

func raDataJSON(c *gin.Context, code int, data any) bool {
	if strings.EqualFold(c.GetHeader(ConsumerHeader), ConsumerTypeRADataJSON) {
		if sl, ok := getSliceLen(data); ok {
			c.Header("X-Total-Count", fmt.Sprintf("%d", sl))
		}
		cm := getCodeMessage(code)
		c.JSON(cm.HttpCode, data)
		return true
	}
	return false
}

func getSliceLen(v any) (int, bool) {
	defer func() {
		if r := recover(); r != nil {
			logrus.Warningf("Recovered in getSliceLen %v", r)
		}
	}()
	rt := reflect.TypeOf(v)
	rv := reflect.ValueOf(v)
	switch rt.Kind() {
	case reflect.Slice, reflect.Array:
		return rv.Len(), true
	default:
		return 0, false
	}
}

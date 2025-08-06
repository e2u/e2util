package e2auth

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type APIResponse struct {
	Status      StatusCode `json:"status,omitempty"`
	ErrCode     ErrCode    `json:"err_code,omitempty"`
	ErrMessage  any        `json:"err_message,omitempty"`
	dynamicData gin.H
}

func (r APIResponse) MarshalJSON() ([]byte, error) {
	data := make(map[string]any)
	data["status"] = r.Status

	if r.ErrMessage != nil {
		data["err_message"] = r.ErrMessage
	}
	if r.ErrCode != "" {
		data["err_code"] = r.ErrCode
	}

	for k, v := range r.dynamicData {
		data[k] = v
	}
	return json.Marshal(data)
}

func errResp(errorCode ErrCode, message any) APIResponse {
	var msg string
	switch v := message.(type) {
	case string:
		msg = v
	case error:
		msg = v.Error()
	default:
		msg = "Unknown error"
	}
	return APIResponse{
		Status:     StatusCodeError,
		ErrCode:    errorCode,
		ErrMessage: msg,
	}
}

func successResp(data gin.H) APIResponse {
	return APIResponse{
		Status:      StatusCodeSuccess,
		dynamicData: data,
	}
}

// func getCSRFSubjectOrAbout(c *gin.Context, secretKey []byte) (*CSRFSubject, error) {
//	subject, err := getVerifiedCSRFSubject(c, secretKey)
//	if err != nil {
//		c.AbortWithStatusJSON(http.StatusForbidden, errResp(ErrCodeInvalidCSRFToken, err))
//		return nil, errors.New("invalid token")
//	}
//	if subject == nil {
//		c.AbortWithStatusJSON(http.StatusForbidden, errResp(ErrCodeInvalidCSRFToken, err))
//		return nil, errors.New("invalid token")
//	}
//	return subject, nil
//}

func getUserOrAbort(c *gin.Context, userId string) (*User, error) {
	user, err := getUserByID(userId)
	if err != nil || user == nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, errResp(ErrCodeUserNotFound, "User does not exist"))
		return nil, err
	}
	return user, nil
}

func getCtxUserOrAbort(c *gin.Context) (*User, error) {
	userId, ok := getCtxUserIdOrAbort(c)
	if !ok {
		return nil, errors.New("user not found in context")
	}
	user, err := getUserOrAbort(c, userId)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func getCtxUserIdOrAbort(c *gin.Context) (string, bool) {
	if userId, ok := c.Get(ctxKeyUserId); ok {
		return userId.(string), true
	}
	c.AbortWithStatusJSON(http.StatusUnauthorized, errResp(ErrCodeUnauthorized, "Unauthorized"))
	return "", false
}

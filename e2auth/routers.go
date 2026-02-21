package e2auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func getProfile(c *gin.Context) {
	user, err := getCtxUserOrAbort(c)
	if err != nil {
		return
	}
	data := gin.H{"id": user.Id, "name": user.Name, "email": user.Email, "roles": user.Roles, "email_verified": user.EmailVerified}
	c.JSON(http.StatusOK, successResp(data))
}

func putProfile(c *gin.Context) {
	var input struct {
		Name  string `json:"name" binding:"omitempty"`
		Email string `json:"email" binding:"omitempty,email"`
	}
	if !bindInput(c, &input) {
		return
	}
	user, err := getCtxUserOrAbort(c)
	if err != nil {
		return
	}
	updates := map[string]any{}
	if input.Name != "" {
		updates["name"] = input.Name
	}
	if input.Email != "" {
		updates["email"] = input.Email
	}
	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, errResp(ErrCodeInvalidInput, "No fields to update"))
		return
	}
	if err = cfg.db.Model(&User{}).Where("id = ?", user.Id).Updates(updates).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, errResp(ErrCodeInternalServerError, err))
		return
	}
	c.JSON(http.StatusOK, successResp(nil))
}

func patchProfile(c *gin.Context) {
	// For simplicity, handle same as putProfile (partial update)
	putProfile(c)
}

func getProfileRoles(c *gin.Context) {
	user, err := getCtxUserOrAbort(c)
	if err != nil {
		return
	}
	c.JSON(http.StatusOK, successResp(gin.H{"roles": user.Roles}))
}

func getSessions(c *gin.Context) {
	user, err := getCtxUserOrAbort(c)
	if err != nil {
		return
	}
	var sessions []Session
	if err = cfg.db.Model(&Session{}).Where("user_id = ?", user.Id).Find(&sessions).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, errResp(ErrCodeInternalServerError, err))
		return
	}
	c.JSON(http.StatusOK, successResp(gin.H{"sessions": sessions}))
}

func deleteSession(c *gin.Context) {
	sessionID := c.Param("id")
	if sessionID == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, errResp(ErrCodeInvalidInput, "missing session id"))
		return
	}
	if err := revokeSession(sessionID); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, errResp(ErrCodeInternalServerError, err))
		return
	}
	c.JSON(http.StatusOK, successResp(nil))
}

func revokeTokens(c *gin.Context) {
	user, err := getCtxUserOrAbort(c)
	if err != nil {
		return
	}
	if err := cfg.db.Model(&Session{}).Where("user_id = ?", user.Id).Update("revoked", true).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, errResp(ErrCodeInternalServerError, err))
		return
	}
	c.JSON(http.StatusOK, successResp(nil))
}

func listUsers(c *gin.Context) {
	var users []User
	if err := cfg.db.Find(&users).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, errResp(ErrCodeInternalServerError, err))
		return
	}
	c.JSON(http.StatusOK, successResp(gin.H{"users": users}))
}

func getUser(c *gin.Context) {
	id := c.Param("id")
	user, err := getUserByID(id)
	if err != nil || user == nil {
		c.AbortWithStatusJSON(http.StatusNotFound, errResp(ErrCodeUserNotFound, "User not found"))
		return
	}
	c.JSON(http.StatusOK, successResp(gin.H{"user": user}))
}

func updateUser(c *gin.Context) {
	id := c.Param("id")
	var input User
	if !bindInput(c, &input) {
		return
	}
	if err := cfg.db.Model(&User{}).Where("id = ?", id).Updates(map[string]any{"name": input.Name, "email": input.Email, "roles": input.Roles}).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, errResp(ErrCodeInternalServerError, err))
		return
	}
	c.JSON(http.StatusOK, successResp(nil))
}

func deleteUser(c *gin.Context) {
	id := c.Param("id")
	if err := cfg.db.Delete(&User{}, "id = ?", id).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, errResp(ErrCodeInternalServerError, err))
		return
	}
	c.JSON(http.StatusOK, successResp(nil))
}

func updateUserRoles(c *gin.Context) {
	id := c.Param("id")
	var payload struct {
		Roles []string `json:"roles"`
	}
	if !bindInput(c, &payload) {
		return
	}
	if err := cfg.db.Model(&User{}).Where("id = ?", id).Update("roles", payload.Roles).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, errResp(ErrCodeInternalServerError, err))
		return
	}
	c.JSON(http.StatusOK, successResp(nil))
}

func searchUsers(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, errResp(ErrCodeInvalidInput, "missing query"))
		return
	}
	var users []User
	if err := cfg.db.Where("name ILIKE ? OR email ILIKE ?", "%"+query+"%", "%"+query+"%").Find(&users).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, errResp(ErrCodeInternalServerError, err))
		return
	}
	c.JSON(http.StatusOK, successResp(gin.H{"users": users}))
}

func getConfig(c *gin.Context) {
	var configs []Configuration
	if err := cfg.db.Find(&configs).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, errResp(ErrCodeInternalServerError, err))
		return
	}
	c.JSON(http.StatusOK, successResp(gin.H{"config": configs}))
}

func updateConfig(c *gin.Context) {
	var input Configuration
	if !bindInput(c, &input) {
		return
	}
	if err := cfg.db.Model(&Configuration{}).Where("key = ?", input.Key).Updates(map[string]any{"value": input.Value, "comment": input.Comment}).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, errResp(ErrCodeInternalServerError, err))
		return
	}
	c.JSON(http.StatusOK, successResp(nil))
}

// Handlers moved to extra_handlers.go

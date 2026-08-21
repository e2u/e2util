package e2auth

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// pageCopy is the UI strings for one language.
type pageCopy struct {
	SignIn                string
	Register              string
	ForgotPassword        string
	ResetPassword         string
	NewPassword           string
	Account               string
	Email                 string
	Password              string
	Username              string
	MFA                   string
	SubmitSignIn          string
	SubmitRegister        string
	PasswordHint          string
	HaveAccount           string
	SendReset             string
	UpdatePassword        string
	BackToSignIn          string
	SignOut               string
	BackToApp             string
	NoticeRegistered      string
	NoticeResetSent       string
	NoticePasswordUpdated string
}

func copyZH() pageCopy {
	return pageCopy{
		SignIn:                "登入",
		Register:              "註冊",
		ForgotPassword:        "忘記密碼",
		ResetPassword:         "重設密碼",
		NewPassword:           "設定新密碼",
		Account:               "帳號",
		Email:                 "電郵",
		Password:              "密碼",
		Username:              "用戶名",
		MFA:                   "MFA 代碼（如已開啟）",
		SubmitSignIn:          "登入",
		SubmitRegister:        "註冊",
		PasswordHint:          "至少 8 位，含大寫、小寫、數字同 !@#$%^&* 其中一個。",
		HaveAccount:           "已有帳號？登入",
		SendReset:             "寄出重設連結",
		UpdatePassword:        "更新密碼",
		BackToSignIn:          "返回登入",
		SignOut:               "登出",
		BackToApp:             "返回應用",
		NoticeRegistered:      "註冊成功，請登入。",
		NoticeResetSent:       "若帳號存在，重設郵件已寄出。",
		NoticePasswordUpdated: "密碼已更新，請重新登入。",
	}
}

func copyEN() pageCopy {
	return pageCopy{
		SignIn:                "Sign in",
		Register:              "Create account",
		ForgotPassword:        "Forgot password",
		ResetPassword:         "Reset password",
		NewPassword:           "Set a new password",
		Account:               "Account",
		Email:                 "Email",
		Password:              "Password",
		Username:              "Username",
		MFA:                   "MFA code (if enabled)",
		SubmitSignIn:          "Sign in",
		SubmitRegister:        "Register",
		PasswordHint:          "At least 8 characters, with uppercase, lowercase, a digit, and one of !@#$%^&*.",
		HaveAccount:           "Already have an account? Sign in",
		SendReset:             "Send reset link",
		UpdatePassword:        "Update password",
		BackToSignIn:          "Back to sign in",
		SignOut:               "Sign out",
		BackToApp:             "Back to app",
		NoticeRegistered:      "Registered. Please sign in.",
		NoticeResetSent:       "If the account exists, a reset email was sent.",
		NoticePasswordUpdated: "Password updated. Please sign in.",
	}
}

func pageCopyFor(lang string) pageCopy {
	if lang == "en" {
		return copyEN()
	}
	return copyZH()
}

func normalizeLang(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	switch {
	case s == "en" || strings.HasPrefix(s, "en-"):
		return "en"
	case s == "zh" || strings.HasPrefix(s, "zh"):
		return "zh"
	default:
		return ""
	}
}

func setLangCookie(c *gin.Context, lang string) {
	secure := c.Request.TLS != nil
	c.SetCookie(langCookie, lang, int((365 * 24 * time.Hour).Seconds()), "/", "", secure, false)
}

func pageLang(c *gin.Context) string {
	if q := normalizeLang(c.Query("lang")); q != "" {
		setLangCookie(c, q)
		return q
	}
	if ck, err := c.Cookie(langCookie); err == nil {
		if q := normalizeLang(ck); q != "" {
			return q
		}
	}
	al := strings.ToLower(c.GetHeader("Accept-Language"))
	if strings.HasPrefix(al, "en") {
		return "en"
	}
	return "zh"
}

func langSwitchURL(c *gin.Context, lang string) string {
	u := *c.Request.URL
	q := u.Query()
	q.Set("lang", lang)
	u.RawQuery = q.Encode()
	if u.Path == "" {
		u.Path = c.Request.URL.Path
	}
	return u.RequestURI()
}

func noticeText(key, lang string) string {
	t := pageCopyFor(lang)
	switch key {
	case "registered":
		return t.NoticeRegistered
	case "reset_sent":
		return t.NoticeResetSent
	case "password_updated":
		return t.NoticePasswordUpdated
	default:
		return ""
	}
}

func pageTitleSuffix(name string, t pageCopy) string {
	switch name {
	case "register":
		return t.Register
	case "forgot":
		return t.ForgotPassword
	case "reset":
		return t.NewPassword
	case "account":
		return t.Account
	default:
		return t.SignIn
	}
}

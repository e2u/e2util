package e2auth

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/e2u/e2util/e2jwt"
	"github.com/gin-gonic/gin"
)

//go:embed html/*.html
var ginPageFS embed.FS

// TemplateFS is the HTML templates for e2gin (names like login.html).
// Use with e2gin.Option.Template.FS and WithGinTemplates().
func TemplateFS() fs.FS {
	sub, err := fs.Sub(ginPageFS, "html")
	if err != nil {
		return ginPageFS
	}
	return sub
}

type pageVM struct {
	Title     string
	AppName   string
	Next      string
	Notice    string
	Error     string
	Token     string
	User      *User
	Lang      string
	HTMLLang  string
	T         pageCopy
	LangURLZH string
	LangURLEN string
}

func pageAppName() string {
	if cfg != nil && cfg.appName != "" {
		return cfg.appName
	}
	return "Account"
}

func prepareVM(c *gin.Context, name string, vm pageVM) pageVM {
	lang := pageLang(c)
	vm.Lang = lang
	vm.HTMLLang = "zh-Hant"
	if lang == "en" {
		vm.HTMLLang = "en"
	}
	vm.T = pageCopyFor(lang)
	if vm.AppName == "" {
		vm.AppName = pageAppName()
	}
	if vm.Title == "" {
		vm.Title = vm.AppName + " · " + pageTitleSuffix(name, vm.T)
	}
	if vm.Notice == "" {
		vm.Notice = noticeText(c.Query("notice"), lang)
	}
	vm.LangURLZH = langSwitchURL(c, "zh")
	vm.LangURLEN = langSwitchURL(c, "en")
	return vm
}

func renderPage(c *gin.Context, name string, vm pageVM) {
	vm = prepareVM(c, name, vm)
	if cfg != nil && cfg.useGinTemplates {
		c.HTML(http.StatusOK, name+".html", vm)
		return
	}
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.Status(http.StatusOK)
	_, _ = c.Writer.WriteString(buildPage(name, vm))
}

func pageLoggedIn(c *gin.Context) bool {
	token := requestSessionToken(c)
	if token == "" {
		return false
	}
	subject, err := e2jwt.VerifyWithEncryptSubject[*SessionSubject](token, getSecretKey())
	if err != nil || subject == nil {
		return false
	}
	session, err := getSessionByID(subject.SessionId)
	if err != nil || session == nil || session.Revoked {
		return false
	}
	c.Set(ctxKeyUserId, subject.UserId)
	c.Set(ctxKeySessionId, subject.SessionId)
	return true
}

func pageLogin(c *gin.Context) {
	if pageLoggedIn(c) {
		c.Redirect(http.StatusFound, safeNext(c.Query("next")))
		return
	}
	renderPage(c, "login", pageVM{
		Next:  safeNext(c.Query("next")),
		Error: c.Query("error"),
	})
}

func pageRegister(c *gin.Context) {
	renderPage(c, "register", pageVM{Error: c.Query("error")})
}

func pageForgot(c *gin.Context) {
	renderPage(c, "forgot", pageVM{Error: c.Query("error")})
}

func pageReset(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.Redirect(http.StatusFound, "/auth/forgot-password")
		return
	}
	renderPage(c, "reset", pageVM{Token: token})
}

func pageAccount(c *gin.Context) {
	if !pageLoggedIn(c) {
		c.Redirect(http.StatusFound, "/auth/login?next=/auth/account")
		return
	}
	user, err := CurrentUser(c)
	if err != nil {
		c.Redirect(http.StatusFound, "/auth/login?next=/auth/account")
		return
	}
	renderPage(c, "account", pageVM{User: user})
}

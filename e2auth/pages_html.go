package e2auth

import (
	h "github.com/e2u/e2util/e2html"
)

const authCSS = `* { box-sizing: border-box; }
body { margin: 0; min-height: 100vh; display: grid; place-items: center;
  font-family: "Iowan Old Style", "Palatino Linotype", Palatino, "Songti SC", serif;
  background: #f3eee4; color: #1c1915; }
main { width: min(26rem, calc(100vw - 2rem)); background: #fffdf8;
  border: 1px solid #d7cbb8; padding: 1.75rem 1.5rem 1.4rem; box-shadow: 6px 6px 0 #1c1915; }
h1 { font-size: 1.35rem; margin: 0 0 .25rem; }
.sub { color: #6b6258; font-size: .85rem; margin: 0 0 1.1rem; }
label { display: block; font-size: .8rem; margin: .7rem 0 .25rem; }
input { width: 100%; padding: .55rem .6rem; border: 1px solid #c9bba6; background: #fff; font: inherit; }
input:focus { outline: 2px solid #1c1915; outline-offset: 1px; }
button, .btn { display: inline-block; margin-top: 1.1rem; width: 100%; text-align: center;
  padding: .65rem; border: 0; background: #1c1915; color: #f3eee4; font: inherit; cursor: pointer; text-decoration: none; }
nav { margin-top: 1rem; font-size: .85rem; }
nav a { color: #1c1915; }
.langs { margin: 0 0 1rem; font-size: .8rem; }
.langs a[aria-current="page"] { font-weight: 700; text-decoration: underline; }
.msg { padding: .5rem .6rem; margin: 0 0 .8rem; font-size: .85rem; border: 1px solid #d7cbb8; background: #f7f0e3; }
.err { border-color: #8a2f1a; background: #f8e7e1; }
.hint { font-size: .75rem; color: #6b6258; margin-top: .35rem; }`

func buildPage(name string, vm pageVM) string {
	var body []h.TAG
	switch name {
	case "register":
		body = registerBody(vm)
	case "forgot":
		body = forgotBody(vm)
	case "reset":
		body = resetBody(vm)
	case "account":
		body = accountBody(vm)
	default:
		body = loginBody(vm)
	}
	return string(h.TS([]h.TAG{
		h.TAG(h.Doctype("html")),
		h.T("html", h.A("lang", vm.HTMLLang),
			h.T("head",
				h.T("meta", h.A("charset", "utf-8")),
				h.T("meta", h.A("name", "viewport"), h.A("content", "width=device-width, initial-scale=1")),
				h.T("title", vm.Title),
				h.TAG("<style>"+authCSS+"</style>"),
			),
			h.T("body", h.T("main", body)),
		),
	}))
}

func langNav(vm pageVM) h.TAG {
	zh := h.Attr{"href": vm.LangURLZH}
	en := h.Attr{"href": vm.LangURLEN}
	if vm.Lang == "zh" {
		zh["aria-current"] = "page"
	}
	if vm.Lang == "en" {
		en["aria-current"] = "page"
	}
	return h.T("p", h.A("class", "langs"),
		h.T("a", zh, "中文"),
		h.TAG(" · "),
		h.T("a", en, "English"),
	)
}

func msgTags(vm pageVM) []h.TAG {
	var out []h.TAG
	if vm.Notice != "" {
		out = append(out, h.T("p", h.A("class", "msg"), vm.Notice))
	}
	if vm.Error != "" {
		out = append(out, h.T("p", h.Attr{"class": "msg err"}, vm.Error))
	}
	return out
}

func labeledInput(id, name, typ, label string, extra h.Attr) []h.TAG {
	attrs := h.Attr{"id": id, "name": name, "type": typ}
	for k, v := range extra {
		attrs[k] = v
	}
	return []h.TAG{
		h.T("label", h.A("for", id), label),
		h.T("input", attrs),
	}
}

func loginBody(vm pageVM) []h.TAG {
	fields := []h.TAG{
		h.T("input", h.Attr{"type": "hidden", "name": "next", "value": vm.Next}),
	}
	t := vm.T
	fields = append(fields, labeledInput("email", "email", "email", t.Email, h.Attr{"required": true, "autocomplete": "username"})...)
	fields = append(fields, labeledInput("password", "password", "password", t.Password, h.Attr{"required": true, "autocomplete": "current-password"})...)
	fields = append(fields, labeledInput("mfa_token", "mfa_token", "text", t.MFA, h.Attr{"inputmode": "numeric", "autocomplete": "one-time-code"})...)
	fields = append(fields, h.T("button", h.A("type", "submit"), t.SubmitSignIn))
	out := []h.TAG{
		langNav(vm),
		h.T("h1", vm.AppName),
		h.T("p", h.A("class", "sub"), t.SignIn),
	}
	out = append(out, msgTags(vm)...)
	out = append(out, h.T("form", h.Attr{"method": "post", "action": "/auth/login"}, fields))
	out = append(out, h.T("nav",
		h.T("a", h.A("href", "/auth/register"), t.Register),
		h.TAG(" · "),
		h.T("a", h.A("href", "/auth/forgot-password"), t.ForgotPassword),
	))
	return out
}

func registerBody(vm pageVM) []h.TAG {
	t := vm.T
	fields := labeledInput("username", "username", "text", t.Username, h.Attr{"required": true, "minlength": "3", "autocomplete": "username"})
	fields = append(fields, labeledInput("email", "email", "email", t.Email, h.Attr{"required": true, "autocomplete": "email"})...)
	fields = append(fields, labeledInput("password", "password", "password", t.Password, h.Attr{"required": true, "minlength": "8", "autocomplete": "new-password"})...)
	fields = append(fields, h.T("p", h.A("class", "hint"), t.PasswordHint))
	fields = append(fields, h.T("button", h.A("type", "submit"), t.SubmitRegister))
	out := []h.TAG{
		langNav(vm),
		h.T("h1", vm.AppName),
		h.T("p", h.A("class", "sub"), t.Register),
	}
	out = append(out, msgTags(vm)...)
	out = append(out, h.T("form", h.Attr{"method": "post", "action": "/auth/register"}, fields))
	out = append(out, h.T("nav", h.T("a", h.A("href", "/auth/login"), t.HaveAccount)))
	return out
}

func forgotBody(vm pageVM) []h.TAG {
	t := vm.T
	fields := labeledInput("email", "email", "email", t.Email, h.Attr{"required": true, "autocomplete": "email"})
	fields = append(fields, h.T("button", h.A("type", "submit"), t.SendReset))
	out := []h.TAG{
		langNav(vm),
		h.T("h1", vm.AppName),
		h.T("p", h.A("class", "sub"), t.ResetPassword),
	}
	out = append(out, msgTags(vm)...)
	out = append(out, h.T("form", h.Attr{"method": "post", "action": "/auth/reset-password"}, fields))
	out = append(out, h.T("nav", h.T("a", h.A("href", "/auth/login"), t.BackToSignIn)))
	return out
}

func resetBody(vm pageVM) []h.TAG {
	fields := []h.TAG{
		h.T("input", h.Attr{"type": "hidden", "name": "token", "value": vm.Token}),
	}
	t := vm.T
	fields = append(fields, labeledInput("password", "password", "password", t.Password, h.Attr{"required": true, "minlength": "8", "autocomplete": "new-password"})...)
	fields = append(fields, h.T("p", h.A("class", "hint"), t.PasswordHint))
	fields = append(fields, h.T("button", h.A("type", "submit"), t.UpdatePassword))
	out := []h.TAG{
		langNav(vm),
		h.T("h1", vm.AppName),
		h.T("p", h.A("class", "sub"), t.NewPassword),
	}
	out = append(out, msgTags(vm)...)
	out = append(out, h.T("form", h.Attr{"method": "post", "action": "/auth/reset-password/confirm"}, fields))
	out = append(out, h.T("nav", h.T("a", h.A("href", "/auth/login"), t.BackToSignIn)))
	return out
}

func accountBody(vm pageVM) []h.TAG {
	t := vm.T
	out := []h.TAG{
		langNav(vm),
		h.T("h1", vm.AppName),
		h.T("p", h.A("class", "sub"), t.Account),
	}
	if vm.User != nil {
		out = append(out,
			h.T("p", h.T("strong", vm.User.Name)),
			h.T("p", vm.User.Email),
			h.T("p", h.A("class", "hint"), "ID: "+vm.User.Id),
		)
	}
	out = append(out,
		h.T("a", h.Attr{"class": "btn", "href": "/auth/logout"}, t.SignOut),
		h.T("nav", h.T("a", h.A("href", "/"), t.BackToApp)),
	)
	return out
}

package middlewares

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
)

type ResourceSrcInterface interface {
	GetSelf() bool
	GetHosts() []string
}

type HostSrc struct {
	Self  bool
	Hosts []string
}

type ScriptStyleSrc struct {
	HostSrc
	UnsafeInline bool
}

type ImgFontMediaSrc struct {
	HostSrc
	Data bool
}

func (h HostSrc) GetSelf() bool {
	return h.Self
}

func (h HostSrc) GetHosts() []string {
	return h.Hosts
}

func (s ScriptStyleSrc) GetSelf() bool {
	return s.Self
}

func (s ScriptStyleSrc) GetHosts() []string {
	return s.Hosts
}

func (i ImgFontMediaSrc) GetSelf() bool {
	return i.Self
}

func (i ImgFontMediaSrc) GetHosts() []string {
	return i.Hosts
}

type SecurityHeadersConfig struct {
	FontSrc                 ImgFontMediaSrc
	ImgSrc                  ImgFontMediaSrc
	ScriptSrc               ScriptStyleSrc
	StyleSrc                ScriptStyleSrc
	ConnectSrc              HostSrc
	MediaSrc                ImgFontMediaSrc
	ObjectSrc               HostSrc
	WorkerSrc               HostSrc
	ManifestSrc             HostSrc
	PrefetchSrc             HostSrc
	FrameSrc                HostSrc
	XFrameOptions           string
	StrictTransportSecurity string
	OtherHeaders            map[string]string
}

func DefaultSecurityHeaders() gin.HandlerFunc {
	return SecurityHeaders(SecurityHeadersConfig{
		FontSrc:     ImgFontMediaSrc{HostSrc: HostSrc{Self: true, Hosts: []string{}}, Data: true},
		ImgSrc:      ImgFontMediaSrc{HostSrc: HostSrc{Self: true, Hosts: []string{}}, Data: true},
		ScriptSrc:   ScriptStyleSrc{HostSrc: HostSrc{Self: true, Hosts: []string{"https://challenges.cloudflare.com"}}},
		StyleSrc:    ScriptStyleSrc{HostSrc: HostSrc{Self: true, Hosts: []string{"https://challenges.cloudflare.com"}}},
		ConnectSrc:  HostSrc{Hosts: []string{"https://challenges.cloudflare.com"}},
		MediaSrc:    ImgFontMediaSrc{HostSrc: HostSrc{Self: true, Hosts: []string{}}, Data: true},
		ObjectSrc:   HostSrc{},
		WorkerSrc:   HostSrc{},
		ManifestSrc: HostSrc{},
		PrefetchSrc: HostSrc{},
		FrameSrc:    HostSrc{Self: false, Hosts: []string{"https://challenges.cloudflare.com"}},
	})
}

func SecurityHeaders(config SecurityHeadersConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		csp := "default-src 'self'; "

		csp += buildResourceSrc("font-src", config.FontSrc)
		csp += buildResourceSrc("img-src", config.ImgSrc)
		csp += buildResourceSrc("script-src", config.ScriptSrc)
		csp += buildResourceSrc("style-src", config.StyleSrc)
		csp += buildResourceSrc("connect-src", config.ConnectSrc)
		csp += buildResourceSrc("media-src", config.MediaSrc)
		csp += buildResourceSrc("object-src", config.ObjectSrc)
		csp += buildResourceSrc("worker-src", config.WorkerSrc)
		csp += buildResourceSrc("manifest-src", config.ManifestSrc)
		csp += buildResourceSrc("prefetch-src", config.PrefetchSrc)
		csp += buildResourceSrc("frame-src", config.FrameSrc)

		if config.XFrameOptions != "" {
			c.Writer.Header().Set("X-Frame-Options", config.XFrameOptions)
		} else {
			c.Writer.Header().Set("X-Frame-Options", "SAMEORIGIN")
		}

		if config.StrictTransportSecurity != "" {
			c.Writer.Header().Set("Strict-Transport-Security", config.StrictTransportSecurity)
		} else {
			c.Writer.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		for key, value := range config.OtherHeaders {
			c.Writer.Header().Set(key, value)
		}

		c.Writer.Header().Set("Content-Security-Policy", strings.TrimSpace(csp))
		c.Next()
	}
}

func buildResourceSrc[T ResourceSrcInterface](directive string, src T) string {
	var parts []string
	switch v := any(src).(type) {
	case HostSrc:
		if v.Self {
			parts = append(parts, "'self'")
		}
		parts = append(parts, v.Hosts...)
	case ScriptStyleSrc:
		if v.Self {
			parts = append(parts, "'self'")
		}
		parts = append(parts, v.Hosts...)
		if v.UnsafeInline {
			parts = append(parts, "'unsafe-inline'")
		}
	case ImgFontMediaSrc:
		if v.Self {
			parts = append(parts, "'self'")
		}
		if v.Data {
			parts = append(parts, "data:")
		}
		parts = append(parts, v.Hosts...)
	}
	if len(parts) > 0 {
		return fmt.Sprintf("%s %s; ", directive, strings.Join(parts, " "))
	}
	return ""
}

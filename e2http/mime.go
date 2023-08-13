package e2http

import (
	"path/filepath"
	"strings"
)

var (
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Basics_of_HTTP/MIME_types/Common_types
	// https://www.iana.org/assignments/media-types/media-types.xhtml
	contentTypeMap = map[string]string{
		"":        "text/html",
		".json":   "application/json",
		".ico":    "image/x-icon",
		".htm":    "text/html",
		".html":   "text/html",
		".png":    "image/png",
		".jpg":    "image/jpeg",
		".jpeg":   "image/jpeg",
		".gif":    "image/gif",
		".txt":    "text/plain",
		".css":    "text/css",
		".map":    "text/html",
		".js":     "application/javascript",
		".aac":    "audio/aac",
		".3g2":    "video/3gpp2",
		".3gp":    "video/3gpp",
		".7z":     "application/x-7z-compressed",
		".abw":    "application/x-abiword",
		".arc":    "application/x-freearc",
		".avi":    "video/x-msvideo",
		".azw":    "application/vnd.amazon.ebook",
		".bin":    "application/octet-stream",
		".bmp":    "image/bmp",
		".bz":     "application/x-bzip",
		".bz2":    "application/x-bzip2",
		".csh":    "application/x-csh",
		".csv":    "text/csv",
		".doc":    "application/msword",
		".docx":   "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		".eot":    "application/vnd.ms-fontobject",
		".epub":   "application/epub+zip",
		".ics":    "text/calendar",
		".jar":    "application/java-archive",
		".jsonld": "application/ld+json",
		".mid":    "audio/midi audio/x-midi",
		".midi":   "audio/midi audio/x-midi",
		".mjs":    "text/javascript",
		".mp3":    "audio/mpeg",
		".mpeg":   "video/mpeg",
		".mpkg":   "application/vnd.apple.installer+xml",
		".odp":    "application/vnd.oasis.opendocument.presentation",
		".ods":    "application/vnd.oasis.opendocument.spreadsheet",
		".odt":    "application/vnd.oasis.opendocument.text",
		".oga":    "audio/ogg",
		".ogv":    "video/ogg",
		".ogx":    "application/ogg",
		".otf":    "font/otf",
		".pdf":    "application/pdf",
		".ppt":    "application/vnd.ms-powerpoint",
		".pptx":   "application/vnd.openxmlformats-officedocument.presentationml.presentation",
		".rar":    "application/x-rar-compressed",
		".rtf":    "application/rtf",
		".sh":     "application/x-sh",
		".svg":    "image/svg+xml",
		".swf":    "application/x-shockwave-flash",
		".tar":    "application/x-tar",
		".tif":    "image/tiff",
		".tiff":   "image/tiff",
		".ttf":    "font/ttf",
		".vsd":    "application/vnd.visio",
		".wav":    "audio/wav",
		".weba":   "audio/webm",
		".webm":   "video/webm",
		".webp":   "image/webp",
		".woff":   "font/woff",
		".woff2":  "font/woff2",
		".xhtml":  "application/xhtml+xml",
		".xls":    "application/vnd.ms-excel",
		".xlsx":   "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		".xml":    "application/xml",
		".xul":    "application/vnd.mozilla.xul+xml",
		".zip":    "application/zip",
	}
)

func GetContentType(filename string) string {
	if v, ok := contentTypeMap[strings.ToLower(filepath.Ext(filename))]; ok {
		return v
	}
	return "application/octet-stream"
}

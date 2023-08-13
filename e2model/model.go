package e2model

import (
	"database/sql"

	"github.com/e2u/e2util/e2array"
)

const (
	HttpPatchOpTest    = "test"
	HttpPatchOpRemove  = "remove"
	HttpPatchOpAdd     = "add"
	HttpPatchOpReplace = "replace"
	HttpPatchOpMove    = "move"
	HttpPatchOpCopy    = "copy"
)

// HttpPatch 操作的数据模型
type HttpPatch struct {
	Op        string      `json:"op,omitempty"`        // must in test,remove,add,replace,move,copy
	From      string      `json:"from,omitempty"`      // use in move,copy op
	Path      string      `json:"path,omitempty"`      // like database table colukn
	Value     interface{} `json:"value,omitempty"`     // values
	Extension interface{} `json:"extension,omitempty"` // extension values
}

// AllowOp 检查当前请求是否在允许操作列表中
func (h *HttpPatch) AllowOp(allows []string) bool {
	return e2array.IncludeString(allows, h.Op)
}

// AllowPath 要修改的属性是否在列表中
func (h *HttpPatch) AllowPath(allows []string) bool {
	return e2array.IncludeString(allows, h.Path)
}

// NullBool bool 類型的擴展
type NullBool struct {
	sql.NullBool
	Error error
}

func NewNullBool(b bool, err error) *NullBool {
	return &NullBool{
		NullBool: sql.NullBool{
			Bool:  b,
			Valid: err == nil,
		},
		Error: err,
	}
}

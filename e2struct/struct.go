package e2struct

import (
	"reflect"
	"strings"
)

func PrepareStruct(v interface{}) {
	// 取得值的 reflect.Value
	val := reflect.ValueOf(v)

	// 取得值的 reflect.Type
	typ := reflect.TypeOf(v)

	// 如果值不是 struct，直接返回
	if typ.Kind() != reflect.Ptr || val.IsNil() || val.Elem().Kind() != reflect.Struct {
		return
	}

	// 遍歷 struct 的所有 field
	for i := 0; i < val.Elem().NumField(); i++ {
		field := val.Elem().Field(i)
		fieldType := typ.Elem().Field(i)

		// 如果 field 是 string，對其值做 strings.TrimSpace 處理
		if fieldType.Type.Kind() == reflect.String {
			field.SetString(strings.TrimSpace(field.String()))
		}

		// 如果 field 是另外一個 struct 的指針，且為 nil，初始化這個指針所指向的 struct
		if fieldType.Type.Kind() == reflect.Ptr && fieldType.Type.Elem().Kind() == reflect.Struct {
			if field.IsNil() {
				fieldValue := reflect.New(fieldType.Type.Elem())
				field.Set(fieldValue)
			}
			// 遞迴調用 PrepareStruct 函數處理
			PrepareStruct(field.Interface())
		}

		// 如果 field 是 struct，遞迴調用 PrepareStruct 函數處理
		if fieldType.Type.Kind() == reflect.Struct {
			PrepareStruct(field.Addr().Interface())
		}
	}
}

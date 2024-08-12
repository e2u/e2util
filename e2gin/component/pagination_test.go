package component

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/e2u/e2util/e2json"
)

type PaginationResult1[T any] struct {
	Items            any    `json:"items"`
	TotalCount       int64  `json:"total_count"`
	TotalPages       int    `json:"total_page"`
	PageSize         int    `json:"page_size"`
	CurrentPageCount int    `json:"current_page_count"`
	PageNumber       int    `json:"page_number"`
	Html             any    `json:"html,omitempty"`
	OrderField       string `json:"order_field,omitempty"`
	OrderDirection   string `json:"order_direction,omitempty"`
}

type Product struct {
	Name string `json:"name"`
}

type Field struct {
	Name string
	Type reflect.Type
	Tag  string
}

func createStructType(structName string, fs []Field) reflect.Type {
	// 创建结构体字段
	fields := make([]reflect.StructField, len(fs))
	for i := 0; i < len(fs); i++ {
		fields[i].Name = fs[i].Name
		fields[i].Type = fs[i].Type
		fields[i].Tag = reflect.StructTag(fmt.Sprintf(`json:"%s"`, fs[i].Tag)) // 可以添加标签
	}
	structType := reflect.StructOf(fields)
	reflect.TypeOf(structType).Name()
	return structType
}

func Test_JsonTag(t *testing.T) {

	ppps := []Product{
		{
			Name: "test",
		},
		{
			Name: "test1",
		},
	}
	prs := createStructType("PPPP", []Field{
		{
			Name: "Name",
			Type: reflect.TypeOf(""),
			Tag:  "nam0e",
		},
		{
			Name: "Items",
			Type: reflect.TypeOf([]Product{}),
			Tag:  "afffe",
		},
	})

	pr := reflect.New(prs).Elem()
	pr.FieldByName("Name").Set(reflect.ValueOf("hello"))
	pr.FieldByName("Items").Set(reflect.ValueOf(ppps))

	e2json.FSprintf(os.Stdout, pr.Interface(), true)
}

package e2var

import (
	"fmt"
	"reflect"
	"testing"
)

func Test_IfElse(t *testing.T) {
	t.Run("should return r1", func(t *testing.T) {
		v1 := "abc"
		v2 := "abc"
		r1 := 123
		r2 := 456
		if IfElse(v1, v2, r1, r2) != r1 {
			t.Fatalf("IfElse(%v, %v, %v, %v) should return %v", v1, v2, r1, r2, r1)
		}
	})

	t.Run("should return r2", func(t *testing.T) {
		v1 := "abc"
		v2 := "abc"
		r1 := "ABC"
		r2 := "EEE"
		if IfElse(v1, v2, r1, r2) != r1 {
			t.Fatalf("IfElse(%v, %v, %v, %v) should return %v", v1, v2, r1, r2, r2)
		}
	})
}

func Test_ExpectOrDefault(t *testing.T) {

	t.Run("should return v2", func(t *testing.T) {
		v1 := 10
		v2 := "abc"
		r, _ := ExpectOrDefault(v1, v2)
		if r != v2 {
			t.Fatalf("Test_ExpectOrDefault failed, expect %v, got %v", v2, r)
		}
	})

	t.Run("should return v1", func(t *testing.T) {
		v1 := "ABCC"
		v2 := "abc"
		r, _ := ExpectOrDefault(v1, v2)
		if r != v1 {
			t.Fatalf("Test_ExpectOrDefault failed, expect %v, got %v", v1, r)
		}
	})

	t.Run("should return v2", func(t *testing.T) {
		var v1 map[string]any
		v2 := "abc"
		r, _ := ExpectOrDefault(v1, v2)
		if r != v2 {
			t.Fatalf("Test_ExpectOrDefault failed, expect %v, got %v", v1, r)
		}
	})
}

func test(m map[string]interface{}) {
	for k, v := range m {
		rt := reflect.TypeOf(v)
		rv := reflect.ValueOf(v)
		switch rt.Kind() {
		case reflect.Slice:
			fmt.Println(k, "is a slice with element type", rt.Elem())
			fmt.Println(k, rv.Len())
		case reflect.Array:
			fmt.Println(k, "is an array with element type", rt.Elem())
			fmt.Println(k, rv.Len())
		default:
			fmt.Println(k, "is something else entirely")
		}
	}
}

type TT struct {
	Name string
	Age  int
}

func Test_typeof(t *testing.T) {
	m := make(map[string]any)
	m["a"] = []string{"a", "b", "c"}
	m["b"] = [4]int{1, 2, 3, 4}
	m["c"] = []TT{{Name: "AAname", Age: 30}, {Name: "Dodd", Age: 50}}
	m["d"] = "hello"
	test(m)
}

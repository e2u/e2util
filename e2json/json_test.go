package e2json

import (
	"fmt"
	"testing"
)

func Test_MustToJSONPString(t *testing.T) {
	t.Run("001", func(t *testing.T) {
		var st = struct {
			A string
			B string
		}{
			A: "hi",
			B: "hello",
		}
		t.Log(MustToJSONPString(st))
	})

	t.Run("002", func(t *testing.T) {
		var i = 100
		t.Log(MustToJSONString(i))
	})

	t.Run("003", func(t *testing.T) {
		jstr := `{"age":"10","pi":"3.14","done":"F"}`
		var st = struct {
			Age  FlexInt64   `json:"age"`
			Pi   FlexFloat64 `json:"pi"`
			Done FlexBool    `json:"done"`
		}{}
		err := MustFromJSONString(jstr, &st)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("%+v", st)
		if !st.Age.Valid && st.Age.Int64 != 10 {
			t.Fatal(st.Age)
		}
		if !st.Pi.Valid && st.Pi.Float64 != 3.14 {
			t.Fatal(st.Pi)
		}
		if !st.Done.Valid && st.Done.Bool == true {
			t.Fatal(st.Done)
		}

	})

}

func Test_parse(t *testing.T) {

	fmt.Println(parse[int64]([]byte("a")))
}

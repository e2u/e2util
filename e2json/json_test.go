package e2json

import (
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

}

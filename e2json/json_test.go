package e2json

import (
	"testing"
)

func Test_MustToJSONPString(t *testing.T) {
	var st = struct {
		A string
		B string
	}{
		A: "hi",
		B: "hello",
	}
	t.Log(MustToJSONPString(st))
}

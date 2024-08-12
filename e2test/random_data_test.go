package e2test

import (
	"testing"
)

func Test_RandomWord(t *testing.T) {
	t.Log(RandomWord())
}

func Test_RandomWords(t *testing.T) {
	t.Log(RandomWords(5, 20))
}

func Test_RandomPhrase(t *testing.T) {
	t.Log(RandomPhrase(5, 20))
}

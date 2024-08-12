package e2test

import (
	"context"
	_ "embed"
	"os"
	"path/filepath"
	"strings"

	"github.com/e2u/e2util/e2crypto"
	"github.com/e2u/e2util/e2db"
	"github.com/e2u/e2util/e2exec"
	"github.com/e2u/e2util/e2http"
	"github.com/e2u/e2util/e2json"
	"github.com/e2u/e2util/e2os"
	"github.com/sirupsen/logrus"
)

type Type string

const (
	TypeString Type = "string"
	TypeInt    Type = "int"
	TypeBool   Type = "bool"
	TypeFloat  Type = "float"
)

var AllTypes = []Type{TypeString, TypeInt, TypeBool, TypeFloat}

// data source
// https://random-word-api.herokuapp.com/all
var words []string

func InitWords() error {
	localFile := filepath.Join(os.TempDir(), "e2u_test_all_en_words.json")
	logrus.Infof("cache local words to: %s", localFile)
	if e2os.FileExists(localFile) {
		if data, err := os.ReadFile(localFile); err == nil {
			if err = e2json.MustFromJSONByte(data, &words); err == nil {
				return nil
			}
		}
	}

	dataUrl := "https://raw.githubusercontent.com/e2u/words/main/all_en_words.json"
	logrus.Infof("load data from %s", dataUrl)
	h := e2http.Builder(context.TODO()).URL(dataUrl)
	h.Do()
	data := h.Body()

	if err := e2json.MustFromJSONByte(data, &words); err != nil {
		return err
	}
	return os.WriteFile(localFile, data, os.ModePerm)
}

func RandomWord() string {
	if len(words) == 0 {
		e2exec.SilentError(InitWords())
	}
	ri := e2crypto.RandomNumber(0, len(words))
	return words[ri]
}

func RandomWords(minNumber, maxNumber int64) []string {
	if len(words) == 0 {
		e2exec.SilentError(InitWords())
	}

	var ws []string
	number := e2crypto.RandomNumber(minNumber, maxNumber)
	for i := int64(0); i < number; i++ {
		ws = append(ws, RandomWord())
	}
	return ws
}

func RandomPhrase(minWords, maxWords int64) string {
	s := strings.Join(RandomWords(minWords, maxWords), " ")

	return s
}

func RandomValue(t Type, min, max int64) any {
	rn := e2crypto.RandomNumber(min, max)
	switch t {
	case "int":
		return e2crypto.RandomNumber(min, max)
	case "bool":
		return rn%2 == 0
	case "string":
		return e2crypto.RandomString(int(e2crypto.RandomNumber(min, max)))
	case "float":
		return e2crypto.RandomFloat(float64(min), float64(max))
	}

	return nil
}

func RandomJSONBArray(minNumber, maxNumber, minLen, maxLen int64) e2db.JSONBArray {
	var rs e2db.JSONBArray
	number := e2crypto.RandomNumber(minNumber, maxNumber)
	for i := int64(0); i < number; i++ {
		rs = append(rs, RandomValue(e2exec.Must(e2crypto.RandomElement(AllTypes)), minLen, maxLen))
	}
	return rs
}

func RandomJSONBMap(minNumber, maxNumber, minKeyLen, maxKeyLen, minValueLen, maxValueLen int64) e2db.JSONBMap {
	rs := make(e2db.JSONBMap)
	number := e2crypto.RandomNumber(minNumber, maxNumber)

	for i := int64(0); i < number; i++ {
		rs[e2crypto.RandomString(int(e2crypto.RandomNumber(minKeyLen, maxKeyLen)))] = RandomValue(e2exec.Must(e2crypto.RandomElement(AllTypes)), minValueLen, maxValueLen)
	}

	return rs
}

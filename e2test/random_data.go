package e2test

import (
	"context"
	_ "embed"
	"os"
	"path/filepath"
	"strings"

	"github.com/e2u/e2util/e2crypto"
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
		if data, err := os.ReadFile(filepath.Clean(localFile)); err == nil {
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
	ri, _ := e2crypto.RandomNumber(0, len(words))
	return words[ri]
}

func RandomWords(minNumber, maxNumber int64) []string {
	if len(words) == 0 {
		e2exec.SilentError(InitWords())
	}

	var ws []string
	number, _ := e2crypto.RandomNumber(minNumber, maxNumber)
	for range number {
		ws = append(ws, RandomWord())
	}
	return ws
}

func RandomPhrase(minWords, maxWords int64) string {
	s := strings.Join(RandomWords(minWords, maxWords), " ")

	return s
}

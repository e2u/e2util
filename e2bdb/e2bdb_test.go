package e2bdb

import (
	"bytes"
	"strings"
	"testing"

	"github.com/dgraph-io/badger/v4"
	"github.com/e2u/e2util/e2crypto"
	"github.com/e2u/e2util/e2json"
	"github.com/e2u/e2util/e2test"
)

func Test_FileStruct(t *testing.T) {
	f := File{
		Key:          "id-12345678",
		Name:         "hello.jpg",
		Size:         100,
		Type:         "image/jpeg",
		LastModified: 1234566,
		Hash:         "1234567890",
		Reader:       strings.NewReader("hello"),
	}

	t.Log(e2json.MustToJSONString(f, true))
}

func Test_BDB(t *testing.T) {
	bdb, err := New("/tmp/test-bdb")
	if err != nil {
		t.Fatal(err)
	}
	defer bdb.Close()

	t.Run("test exists", func(t *testing.T) {
		isExists, err := bdb.Exists("abcdeffg")
		if err != nil {
			t.Fatal(err)
		}
		t.Log(isExists)
	})

	t.Run("test storage and load content", func(t *testing.T) {

		key := "a"
		content := []byte("hello")
		err := bdb.bdb.Update(func(txn *badger.Txn) error {
			t.Logf(">>>>>> key=%v", key)
			sErr := bdb.storageContent(txn, key, content)
			if sErr != nil {
				t.Errorf(">>>>> storage error: %v", sErr)
				return sErr
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
		err = bdb.bdb.View(func(txn *badger.Txn) error {
			bc, err := bdb.loadContent(txn, key)
			if err != nil {
				t.Errorf(">>>>> load error: %v", err)
				return err
			}
			t.Logf(">>>>> loaded content: %v", string(bc))
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("test storage and load", func(t *testing.T) {
		id := e2crypto.RandomString(8)
		err := bdb.StorageFile(&File{
			Key:          id,
			Name:         e2test.RandomWord(),
			Size:         100,
			Type:         "file/data",
			LastModified: 1234567890,
			Reader:       bytes.NewBufferString("hello"),
		}, DefaultOptions())
		if err != nil {
			t.Fatalf(">>>>>>>>> storage error=%v", err)
		}

		f, err := bdb.LoadFile(id)
		if err != nil {
			t.Fatalf(">>>>>>>>> load error=%v", err)
		}
		t.Log(f)
	})

}

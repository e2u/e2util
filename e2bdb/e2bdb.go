package e2bdb

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"io"
	"strings"

	"github.com/dgraph-io/badger/v4"
	"github.com/e2u/e2util/e2hash"
	"github.com/e2u/e2util/e2json"
	"github.com/e2u/e2util/e2slice"
	"github.com/sirupsen/logrus"
)

const (
	mateKeyPrefix    = "$M$"
	contentKeyPrefix = "$C$"
)

type BDB struct {
	bdb *badger.DB
}

type Options struct {
	ForceOverwrite bool
}

func DefaultOptions() Options {
	return Options{
		ForceOverwrite: false,
	}
}

func (opt Options) WithForceOverwrite(val bool) Options {
	opt.ForceOverwrite = val
	return opt
}

func checkAndSetOptions(opt *Options) error {
	return nil
}

func New(path string, opts ...badger.Options) (*BDB, error) {
	opt := badger.DefaultOptions(path)
	if len(opts) > 0 {
		opt = opts[0]
	}
	opt.WithCompactL0OnClose(true).
		WithVLogPercentile(0.5)
	b, err := badger.Open(opt)
	if err != nil {
		return nil, err
	}
	return &BDB{
		bdb: b,
	}, nil
}

type File struct {
	Key          string `json:"id,omitempty"`
	Name         string `json:"name,omitempty"`
	Size         int    `json:"size,omitempty"`
	Type         string `json:"type,omitempty"`
	LastModified int64  `json:"last_modified,omitempty"`
	Hash         string `json:"hash,omitempty"`
	io.Reader    `json:"-"`
	content      []byte `json:"-"`
}

func (f *File) Prepare() error {
	if f.Key == "" {
		return errors.New("missing key")
	}

	bc, err := io.ReadAll(f.Reader)
	if err != nil {
		logrus.WithField("scope", "Prepare").Errorf("read File.Reader failed, error=%v", err)
		return err
	}
	f.content = e2slice.Copy(bc)
	f.Hash = contentHash(f.content)
	f.Size = len(bc)

	return nil
}

func (f *File) Content() []byte {
	return f.content
}

func (b *BDB) formatMateKey(key string) []byte {
	if !strings.HasPrefix(key, mateKeyPrefix) {
		key = mateKeyPrefix + key
	}
	return []byte(key)
}

func (b *BDB) formatContentKey(key string) []byte {
	if !strings.HasPrefix(key, contentKeyPrefix) {
		key = contentKeyPrefix + key
	}
	return []byte(key)
}

func (b *BDB) Close() error {
	return b.bdb.Close()
}

func (b *BDB) StorageFile(f *File, opts ...Options) error {
	opt := DefaultOptions()
	if len(opts) > 0 {
		opt = opts[0]
	}
	if err := checkAndSetOptions(&opt); err != nil {
		return err
	}

	if err := f.Prepare(); err != nil {
		return err
	}

	return b.bdb.Update(func(txn *badger.Txn) error {
		if err := b.storageContent(txn, f.Key, f.Content()); err != nil {
			return err
		}
		if err := b.storageMate(txn, f); err != nil {
			return err
		}
		return nil
	})
}

func (b *BDB) LoadFile(key string) (*File, error) {
	var file *File
	if iErr := b.bdb.View(func(txn *badger.Txn) error {
		content, err := b.loadContent(txn, key)
		if err != nil {
			return err
		}
		file, err = b.loadMate(txn, key)
		if err != nil {
			return err
		}
		// file.content = make([]byte, len(content))
		// copy(file.content, content)
		file.content = e2slice.Copy(content)
		file.Reader = bytes.NewReader(content)
		return nil
	}); iErr != nil {
		return nil, iErr
	}
	return file, nil
}

func (b *BDB) Delete(key string) error {
	return b.bdb.Update(func(txn *badger.Txn) error {
		if err := b.deleteMate(txn, key); err != nil {
			return err
		}
		if err := b.deleteContent(txn, key); err != nil {
			return err
		}
		return nil
	})
}

func (b *BDB) Exists(key string) (bool, error) {
	var isExists bool
	err := b.bdb.View(func(txn *badger.Txn) error {
		var err error
		isExists, err = b.exists(txn, key)
		if err != nil {
			return err
		}
		return nil
	})
	return isExists, err
}

func (b *BDB) exists(txn *badger.Txn, key string) (bool, error) {
	mate, err := b.loadMate(txn, key)
	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			return false, nil
		}
		return false, err
	}
	content, err := b.loadContent(txn, key)
	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			return false, nil
		}
		return false, err
	}
	return mate.Hash == contentHash(content), nil
}

func (b *BDB) loadMate(txn *badger.Txn, key string) (*File, error) {
	item, err := txn.Get(b.formatMateKey(key))
	if err != nil {
		return nil, err
	}
	file := &File{}
	err = item.Value(func(val []byte) error {
		return e2json.MustFromJSONByte(val, file)
	})
	return file, err
}

func (b *BDB) loadContent(txn *badger.Txn, key string) ([]byte, error) {
	var content []byte
	item, err := txn.Get(b.formatContentKey(key))
	if err != nil {
		return nil, err
	}
	err = item.Value(func(val []byte) error {
		content = e2slice.Copy(val)
		return nil
	})
	return content, err
}

func (b *BDB) storageMate(txn *badger.Txn, f *File) error {
	logrus.Infof("e2bdb - storage mate key=%v, mate=%v", f.Key, e2json.MustToJSONString(f))
	if err := txn.Set(b.formatMateKey(f.Key), e2json.MustToJSONByte(f)); err != nil {
		logrus.Errorf("e2bdb - storage mate error=%v, key=%v", err, f.Key)
		return err
	}
	return nil
}

func (b *BDB) storageContent(txn *badger.Txn, key string, content []byte) error {
	logrus.Infof("e2bdb - storage content key=%v", key)
	if err := txn.Set(b.formatContentKey(key), content); err != nil {
		logrus.Errorf("e2bdb - storage file content error=%v, key=%v, size=%d", err, key, len(content))
		return err
	}
	return nil
}

func (b *BDB) deleteMate(txn *badger.Txn, key string) error {
	return txn.Delete(b.formatMateKey(key))
}

func (b *BDB) deleteContent(txn *badger.Txn, key string) error {
	return txn.Delete(b.formatContentKey(key))
}

func contentHash(data []byte) string {
	return e2hash.HashHex(data, sha256.New)
}

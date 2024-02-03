package e2crypto

import (
	"crypto/rand"
	"encoding/base64"
	"math/big"
	"strings"
)

// RandomString 返回一个随机字符串,base64 范围,移除 / 和 b
func RandomString(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	rs := base64.StdEncoding.WithPadding(base64.NoPadding).EncodeToString(b)
	rs = strings.ReplaceAll(rs, "/", "a")
	rs = strings.ReplaceAll(rs, "+", "b")
	if len(rs) > n {
		return rs[:n]
	}
	return rs
}

// RandomBytes 返回随机字节数组
func RandomBytes(n int) []byte {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return []byte("")
	}
	return b
}

// RandomUint 生成一個介於 min 和 max 之間的隨機數
func RandomUint(min, max int64) int64 {
	nb, err := rand.Int(rand.Reader, big.NewInt(max-min+1))
	if err != nil {
		return 0
	}
	return nb.Int64() + min
}

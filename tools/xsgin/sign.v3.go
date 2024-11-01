package xsign

import (
	"crypto/md5"
	"encoding/hex"
	"hash"
	"io"
)

type ISigner interface {
	io.Writer
	CheckSum() string
}

type SignV3 struct {
	md5     hash.Hash
	signKey string
}

func NewSignV3(signKey string) *SignV3 {
	return &SignV3{
		md5:     md5.New(),
		signKey: signKey,
	}
}

func (x *SignV3) Write(p []byte) (n int, err error) {
	return x.md5.Write(p)
}

func (x *SignV3) CheckSum() string {
	x.md5.Write([]byte(x.signKey))
	return hex.EncodeToString(x.md5.Sum(nil))
}

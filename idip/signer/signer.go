// Package signer provides functionality for signing IDIP requests
package signer

import (
	"crypto/md5"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// Header represents the signing headers used in requests
type Header struct {
	AppId   string `json:"idip-sign-appid"`
	Time    string `json:"idip-sign-time"`
	Rnd     string `json:"idip-sign-rnd"`
	Sign    string `json:"idip-sign"`
	Version string `json:"idip-sign-version"`
}

// ISigner defines the interface for signing operations
type ISigner interface {
	Sign(body, tag string) string
	GenSignedHeader(body, tag string, version string) Header
	GetRnd() int
	GetTimestamp() int64
}

// Signer implements the ISigner interface
type Signer struct {
	appId     string
	appSecret string
	timestamp int64
	rnd       int
}

// NewSigner creates a new Signer instance with the given credentials
func NewSigner(appId, secret string) (*Signer, error) {
	if appId == "" {
		return nil, fmt.Errorf("appId can not empty")
	}

	if secret == "" {
		return nil, fmt.Errorf("secret can not empty")
	}

	return &Signer{
		appSecret: secret,
		appId:     appId,
		timestamp: time.Now().Unix(),
		rnd:       rand.Intn(1000),
	}, nil
}

// Sign generates a signature for the given body and tag
func (s Signer) Sign(body, tag string) string {
	signString := fmt.Sprintf("%s%s%d%d%s", tag, body, s.timestamp, s.rnd, s.appSecret)
	hash := md5.Sum([]byte(signString))
	return fmt.Sprintf("%x", hash)
}

// GenSignedHeader generates a complete Header with signature
func (s Signer) GenSignedHeader(body, tag string, version string) Header {
	if version == "" {
		version = "v1"
	}

	signStr := s.Sign(body, tag)
	return Header{
		AppId:   s.appId,
		Time:    fmt.Sprintf("%d", s.timestamp),
		Rnd:     fmt.Sprintf("%d", s.rnd),
		Sign:    strings.ToUpper(strings.ReplaceAll(signStr, "-", "")),
		Version: version,
	}
}

// Header2Map converts a Header struct to a map[string]string
func (h Header) Header2Map() map[string]string {
	return map[string]string{
		"idip-sign-appid":   h.AppId,
		"idip-sign-time":    h.Time,
		"idip-sign-rnd":     h.Rnd,
		"idip-sign":         h.Sign,
		"idip-sign-version": h.Version,
	}
}

// GetRnd returns the random number used in signing
func (s Signer) GetRnd() int {
	return s.rnd
}

// GetTimestamp returns the timestamp used in signing
func (s Signer) GetTimestamp() int64 {
	return s.timestamp
}

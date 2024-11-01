package xtls

import (
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"fmt"
)

type SSLPinningCheckFunc func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error

func SSLPinningChecker(exceptedHash string) SSLPinningCheckFunc {
	return func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
		// 如果没有设置预期哈希，则不检查
		if exceptedHash == "" {
			return nil
		}
		if len(rawCerts) == 0 {
			return fmt.Errorf("no certificates provided")
		}

		// 只检查第一个证书（服务器证书）
		cert, err := x509.ParseCertificate(rawCerts[0])
		if err != nil {
			return fmt.Errorf("failed to parse certificate: %v", err)
		}

		actualHash, _ := CalCertPubHash(cert)
		// 比较哈希
		if actualHash != exceptedHash {
			return fmt.Errorf("certificate hash does not match expected hash")
		}

		return nil
	}
}

func CalCertPubHash(cert *x509.Certificate) (string, error) {
	// 计算公钥的 SHA256 哈希
	pubKeyBytes, err := x509.MarshalPKIXPublicKey(cert.PublicKey)
	if err != nil {
		return "", fmt.Errorf("failed to marshal public key: %v", err)
	}
	hash := sha256.Sum256(pubKeyBytes)
	return base64.StdEncoding.EncodeToString(hash[:]), nil
}

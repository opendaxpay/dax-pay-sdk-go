package daxpay

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
)

// 加载私钥 PEM（PKCS#8 优先，回退 PKCS#1）— 对照后端 RsaSignUtil#loadPrivateKeyFromPem
func parsePrivateKey(pemStr string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, errors.New("私钥 PEM 解析失败：无效的 PEM 格式")
	}
	if key, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
		rsaKey, ok := key.(*rsa.PrivateKey)
		if !ok {
			return nil, errors.New("私钥非 RSA 类型")
		}
		return rsaKey, nil
	}
	return x509.ParsePKCS1PrivateKey(block.Bytes)
}

// 加载公钥 PEM（X.509 优先，回退 PKCS#1）— 对照后端 RsaSignUtil#loadPublicKeyFromPem
func parsePublicKey(pemStr string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, errors.New("公钥 PEM 解析失败：无效的 PEM 格式")
	}
	if key, err := x509.ParsePKIXPublicKey(block.Bytes); err == nil {
		rsaKey, ok := key.(*rsa.PublicKey)
		if !ok {
			return nil, errors.New("公钥非 RSA 类型")
		}
		return rsaKey, nil
	}
	return x509.ParsePKCS1PublicKey(block.Bytes)
}

// ValidatePrivateKeyPEM 校验商户私钥 PEM 是否可解析（PKCS#8 优先，回退 PKCS#1）
// 联调页面保存配置时做即时校验用；解析失败返回带原因的 error
func ValidatePrivateKeyPEM(pemStr string) error {
	_, err := parsePrivateKey(pemStr)
	return err
}

// ValidatePublicKeyPEM 校验平台公钥 PEM 是否可解析（X.509 优先，回退 PKCS#1）
func ValidatePublicKeyPEM(pemStr string) error {
	_, err := parsePublicKey(pemStr)
	return err
}

// RsaSign 用商户私钥对 data 进行 SHA256withRSA 签名，返回 Base64
// 对照后端 RsaSignUtil#sign（UTF-8 字节，PKCS1v15 确定性签名）
func RsaSign(data string, privateKeyPem string) (string, error) {
	key, err := parsePrivateKey(privateKeyPem)
	if err != nil {
		return "", fmt.Errorf("签名失败: %w", err)
	}
	hashed := sha256.Sum256([]byte(data)) // UTF-8 字节
	sig, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, hashed[:])
	if err != nil {
		return "", fmt.Errorf("签名失败: %w", err)
	}
	return base64.StdEncoding.EncodeToString(sig), nil
}

// RsaVerify 用平台公钥验签（SHA256withRSA）— 对照后端 RsaSignUtil#verify
func RsaVerify(data string, signB64 string, publicKeyPem string) (bool, error) {
	key, err := parsePublicKey(publicKeyPem)
	if err != nil {
		return false, fmt.Errorf("验签失败: %w", err)
	}
	sig, err := base64.StdEncoding.DecodeString(signB64)
	if err != nil {
		return false, fmt.Errorf("签名 Base64 解码失败: %w", err)
	}
	hashed := sha256.Sum256([]byte(data))
	return rsa.VerifyPKCS1v15(key, crypto.SHA256, hashed[:], sig) == nil, nil
}

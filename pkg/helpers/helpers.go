package helpers

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"github.com/edimarlnx/secure-templates/pkg/config"
	"log"
	"os"
	"strings"
)

func GetEnv(name, defaultValue string) string {
	value := os.Getenv(name)
	if strings.TrimSpace(value) != "" {
		return value
	}
	return defaultValue
}

func ParseConfig(filename string) config.SecureTemplateConfig {
	data, err := os.ReadFile(filename)
	if err != nil {
		log.Fatalf("Error on parse config file: %s", filename)
	}
	var cfg config.SecureTemplateConfig
	err = json.Unmarshal(data, &cfg)
	if err != nil {
		log.Fatalf("Error on parse config file: %s", filename)
	}
	return cfg
}

func ExportRsaPrivateKeyAsPemStr(privKey *rsa.PrivateKey) string {
	privKeyBytes := x509.MarshalPKCS1PrivateKey(privKey)
	privKeyPem := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: privKeyBytes,
		},
	)
	return base64.StdEncoding.EncodeToString(privKeyPem)
}

func ParseRsaPrivateKeyFromPemStr(privKeyBase64 string) (*rsa.PrivateKey, error) {
	data, err := base64.StdEncoding.DecodeString(privKeyBase64)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode([]byte(data))
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the key")
	}
	privKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return privKey, nil
}

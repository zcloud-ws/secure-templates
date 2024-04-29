package helpers

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"github.com/joho/godotenv"
	"github.com/zcloud-ws/secure-templates/pkg/config"
	"github.com/zcloud-ws/secure-templates/pkg/logging"
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

func ParseConfig(filename string, isConfigRequired bool) config.SecureTemplateConfig {
	if filename == "" && !isConfigRequired {
		logging.Log.Warnf("Using print key as secret connector.\n")
		return config.SecureTemplateConfig{
			SecretEngine: config.SecretEnginePrintKeys,
		}
	}
	data, err := os.ReadFile(filename)
	if err != nil {
		logging.Log.Fatalf("Error on parse config file: %s\n", filename)
	}
	var cfg config.SecureTemplateConfig
	err = json.Unmarshal(data, &cfg)
	if err != nil {
		logging.Log.Fatalf("Error on parse config file: %s\n", filename)
	}
	cfg.ExpandEnvVars()
	return cfg
}

func GenRsaPrivateKey(pwd string) ([]byte, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	block := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}
	if pwd != "" {
		blk, err := x509.EncryptPEMBlock(rand.Reader, block.Type, block.Bytes, []byte(pwd), x509.PEMCipherAES256)
		if err != nil {
			return nil, err
		}
		block = blk
	}
	privKeyPem := pem.EncodeToMemory(block)
	return privKeyPem, nil
}

func ParseRsaPrivateKeyFromPemStr(privKeyBase64, pwd string) (*rsa.PrivateKey, error) {
	data, err := base64.StdEncoding.DecodeString(privKeyBase64)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the key")
	}
	var certData []byte
	if pwd != "" && x509.IsEncryptedPEMBlock(block) {
		certData, err = x509.DecryptPEMBlock(block, []byte(pwd))
		if err != nil {
			return nil, err
		}
	} else {
		certData = block.Bytes
	}
	privKey, err := x509.ParsePKCS1PrivateKey(certData)
	if err != nil {
		return nil, err
	}

	return privKey, nil
}

func ParseEnvFileAsKeyValue(envFile string) (map[string]string, error) {
	data, err := godotenv.Read(envFile)
	if err != nil {
		return nil, err
	}
	return data, nil
}

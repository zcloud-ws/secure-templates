package helpers

import (
	"bufio"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
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
	file, err := os.Open(envFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.New(fmt.Sprintf("Env file not found: %s", envFile))
		}
		return nil, err
	}
	buf := bufio.NewScanner(file)
	data := map[string]string{}
	for buf.Scan() {
		line := strings.TrimSpace(buf.Text())
		key, value, found := strings.Cut(line, "=")
		if value == "" || key == "" || !found {
			continue
		}
		data[key] = value
	}
	return data, nil
}

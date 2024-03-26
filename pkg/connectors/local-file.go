package connectors

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"github.com/edimarlnx/secure-templates/pkg/config"
	"github.com/edimarlnx/secure-templates/pkg/helpers"
	"github.com/go-jose/go-jose/v3/json"
	"log"
	"os"
)

type LocalFileConnector struct {
	Connector
	secretFile string
	secrets    map[string]map[string]string
	encPrivKey *rsa.PrivateKey
	encPubKey  *rsa.PublicKey
}

func (v *LocalFileConnector) Init(secTplConfig config.SecureTemplateConfig) error {
	v.secretFile = secTplConfig.LocalFileConfig.Filename
	v.secrets = map[string]map[string]string{}
	encPrivKey := helpers.GetEnv("LOCAL_SECRET_PRIVATE_KEY", secTplConfig.LocalFileConfig.EncPrivKey)
	if encPrivKey != "" {
		passphrase := helpers.GetEnv("LOCAL_SECRET_PRIVATE_KEY_PASSPHRASE", secTplConfig.LocalFileConfig.Passphrase)
		privKey, err := helpers.ParseRsaPrivateKeyFromPemStr(encPrivKey, passphrase)
		if err != nil {
			return err
		}
		v.encPrivKey = privKey
		v.encPubKey = &privKey.PublicKey
	}
	return v.loadFromFile()
}

func (v *LocalFileConnector) Secret(secretName, keyName string) any {
	secret := v.secrets[secretName]
	if secret == nil {
		log.Fatalf("secret not exists '%s'", secretName)
		return keyName
	}
	if keyName != "" {
		value, ok := secret[keyName]
		if !ok {
			log.Fatalf("unable to load value for key %s", keyName)
			return keyName
		}
		encData, err := v.decrypt(value)
		if err != nil {
			log.Fatalf("unable to decrypt value for key %s", keyName)
		}
		return encData
	}
	data := map[string]interface{}{}
	for k, vl := range secret {
		encData, err := v.decrypt(vl)
		if err != nil {
			log.Fatalf("unable to decrypt value for key %s", keyName)
		}
		data[k] = encData
	}
	return data
}

func (v *LocalFileConnector) Finalize() {
	if err := v.saveToFile(); err != nil {
		log.Fatal(err)
	}
}

func (v *LocalFileConnector) WriteKey(secretName, keyName, keyValue string) error {
	return v.WriteKeys(secretName, map[string]string{keyName: keyValue})
}

func (v *LocalFileConnector) WriteKeys(secretName string, keyValue map[string]string) error {
	secret := v.secrets[secretName]
	if secret == nil {
		v.secrets[secretName] = map[string]string{}
		secret = v.secrets[secretName]
	}
	for key, value := range keyValue {
		encData, err := v.encrypt(value)
		if err != nil {
			return err
		}
		secret[key] = encData
	}
	return v.saveToFile()
}

func (v *LocalFileConnector) loadFromFile() error {
	data, err := os.ReadFile(v.secretFile)
	if os.IsNotExist(err) {
		v.secrets = map[string]map[string]string{}
		return nil
	}
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, &v.secrets)
	return err
}

func (v *LocalFileConnector) saveToFile() error {
	data, err := json.MarshalIndent(v.secrets, "", "  ")
	if err != nil {
		return err
	}
	err = os.WriteFile(v.secretFile, data, 0700)
	return err
}

func (v *LocalFileConnector) encrypt(str string) (string, error) {
	if v.encPubKey == nil {
		return str, nil
	}
	encData, err := rsa.EncryptOAEP(
		sha256.New(),
		rand.Reader,
		v.encPubKey,
		[]byte(str),
		nil)
	if err != nil {
		return str, err
	}
	return base64.StdEncoding.EncodeToString(encData), nil
}

func (v *LocalFileConnector) decrypt(str string) (string, error) {
	if v.encPrivKey == nil {
		return str, nil
	}
	data, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return str, err
	}
	decData, err := rsa.DecryptOAEP(
		sha256.New(),
		rand.Reader,
		v.encPrivKey,
		data,
		nil)
	if err != nil {
		return str, err
	}
	return string(decData), nil
}

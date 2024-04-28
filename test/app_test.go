package test

import (
	"fmt"
	"os"
	"testing"
)

func Test_initApp(t *testing.T) {
	workdir, err := os.Getwd()
	if err != nil {
		workdir = os.TempDir()
	}
	configFile := "local-file-cfg-test.json"
	tests := []DataTest{
		{
			Name: "init-config",
			Args: []string{
				"secure-templates",
				"init-config",
				"-o",
				configFile,
				"-secret-file",
				fmt.Sprintf("%s/local-file-secret-test.json", workdir),
				"-private-key-passphrase",
				"test-pwd",
			},
			RequiredStrings: []string{},
			Envs: map[string]string{
				"SEC_TPL_CONFIG": configFile,
			},
		},
		{
			Name: "put app user",
			Args: []string{
				"secure-templates",
				"manage-secret",
				"put",
				"core",
				"app_user",
				"dev_user",
			},
			RequiredStrings: []string{
				"saved on secret",
			},
			Envs: map[string]string{
				"SEC_TPL_CONFIG": configFile,
			},
		},
		{
			Name: "put app passwd",
			Args: []string{
				"secure-templates",
				"manage-secret",
				"put",
				"core",
				"app_passwd",
				"2dabe3d7c66fb75f751202fdab19266b",
			},
			RequiredStrings: []string{
				"saved on secret",
			},
			Envs: map[string]string{
				"SEC_TPL_CONFIG": configFile,
			},
		},
		{
			Name: "put client app user",
			Args: []string{
				"secure-templates",
				"manage-secret",
				"put",
				"client",
				"app_user",
				"dev_user",
			},
			RequiredStrings: []string{
				"saved on secret",
			},
			Envs: map[string]string{
				"SEC_TPL_CONFIG": configFile,
			},
		},
		{
			Name: "put client app passwd",
			Args: []string{
				"secure-templates",
				"manage-secret",
				"put",
				"client",
				"app_passwd",
				"2dabe3d7c66fb75f751202fdab19266b",
			},
			RequiredStrings: []string{
				"saved on secret",
			},
			Envs: map[string]string{
				"SEC_TPL_CONFIG": configFile,
			},
		},
		{
			Name: "update secrets from .env file",
			Args: []string{
				"secure-templates",
				"manage-secret",
				"import",
				"test",
				"secrets.env",
			},
			RequiredStrings: []string{
				"8 keys saved on secret",
			},
			Envs: map[string]string{
				"SEC_TPL_CONFIG": configFile,
			},
		},
		{
			Name: "Env file",
			Args: []string{
				"secure-templates",
				"samples/.env",
			},
			RequiredStrings: []string{
				"APP_USER=dev_user",
				"APP_PASSWORD=2dabe3d7c66fb75f751202fdab19266b",
				"CLIENT_APP_USER=dev_user",
				"CLIENT_APP_PASSWORD=2dabe3d7c66fb75f751202fdab19266b",
			},
			Envs: map[string]string{
				"SEC_TPL_CONFIG": configFile,
			},
		},
		{
			Name: "k8s secret yaml",
			Args: []string{
				"secure-templates",
				"samples/k8s-secret.yaml",
			},
			RequiredStrings: []string{
				"name: st-secret",
				"namespace: dev-ns",
				"APP_USER: ZGV2X3VzZXI=",
				"APP_PASSWORD: MmRhYmUzZDdjNjZmYjc1Zjc1MTIwMmZkYWIxOTI2NmI=",
				"CLIENT_APP_USER: \"dev_user\"",
				"CLIENT_APP_PASSWORD: \"2dabe3d7c66fb75f751202fdab19266b\"",
			},
			Envs: map[string]string{
				"SEC_TPL_CONFIG":   configFile,
				"SECRET_NAME":      "st-secret",
				"SECRET_NAMESPACE": "dev-ns",
			},
		},
		{
			Name: "k8s secret yaml - print keys",
			Args: []string{
				"secure-templates",
				"-print-keys",
				"samples/k8s-secret.yaml",
			},
			RequiredStrings: []string{
				"Template keys:",
				"core.app_user",
				"core.app_passwd",
				"client.app_user",
				"client.app_passwd",
			},
			Envs: map[string]string{
				"SEC_TPL_CONFIG":   configFile,
				"SECRET_NAME":      "st-secret",
				"SECRET_NAMESPACE": "dev-ns",
			},
		},
		{
			Name: "Imported secrets from env file",
			Args: []string{
				"secure-templates",
				"samples/secrets.env",
			},
			RequiredStrings: []string{
				"app1_secret=12345",
				"app2_secret=67890",
				"app3_secret=12345",
				"app4_secret=67\"8\"90",
				"app5_secret=67890",
				"app6_secret=\"67890\"",
				"app7_secret=678`90",
				"app8_secret=áçõ",
			},
			Envs: map[string]string{
				"SEC_TPL_CONFIG":   configFile,
				"SECRET_NAME":      "st-secret",
				"SECRET_NAMESPACE": "dev-ns",
			},
		},
		{
			Name: "Use secret range values",
			Args: []string{
				"secure-templates",
				"samples/secrets-list.env",
			},
			RequiredStrings: []string{
				"app1_secret:12345",
				"app2_secret:67890",
				"app3_secret:12345",
				"app4_secret:67\"8\"90",
				"app5_secret:67890",
				"app6_secret:\"67890\"",
				"app7_secret:678`90",
				"app8_secret:áçõ",
			},
			Envs: map[string]string{
				"SEC_TPL_CONFIG":   configFile,
				"SECRET_NAME":      "st-secret",
				"SECRET_NAMESPACE": "dev-ns",
			},
		},
	}

	SuiteTest(t, configFile, tests)
}

package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"testing"
)

type testData struct {
	name            string
	args            []string
	requiredStrings []string
	envs            map[string]string
}

func Test_initApp(t *testing.T) {
	workdir, err := os.Getwd()
	if err != nil {
		workdir = os.TempDir()
	}
	configFile := "test/local-file-cfg-test.json"
	tests := []testData{
		{
			name: "init-config",
			args: []string{
				"secure-templates",
				"init-config",
				"-o",
				configFile,
				"-secret-file",
				fmt.Sprintf("%s/test/local-file-secret-test.json", workdir),
				"-private-key-passphrase",
				"test-pwd",
			},
			requiredStrings: []string{},
			envs: map[string]string{
				"SEC_TPL_CONFIG": configFile,
			},
		},
		{
			name: "put app user",
			args: []string{
				"secure-templates",
				"manage-secret",
				"put",
				"core",
				"app_user",
				"dev_user",
			},
			requiredStrings: []string{
				"saved on secret",
			},
			envs: map[string]string{
				"SEC_TPL_CONFIG": configFile,
			},
		},
		{
			name: "put app passwd",
			args: []string{
				"secure-templates",
				"manage-secret",
				"put",
				"core",
				"app_passwd",
				"2dabe3d7c66fb75f751202fdab19266b",
			},
			requiredStrings: []string{
				"saved on secret",
			},
			envs: map[string]string{
				"SEC_TPL_CONFIG": configFile,
			},
		},
		{
			name: "put client app user",
			args: []string{
				"secure-templates",
				"manage-secret",
				"put",
				"client",
				"app_user",
				"dev_user",
			},
			requiredStrings: []string{
				"saved on secret",
			},
			envs: map[string]string{
				"SEC_TPL_CONFIG": configFile,
			},
		},
		{
			name: "put client app passwd",
			args: []string{
				"secure-templates",
				"manage-secret",
				"put",
				"client",
				"app_passwd",
				"2dabe3d7c66fb75f751202fdab19266b",
			},
			requiredStrings: []string{
				"saved on secret",
			},
			envs: map[string]string{
				"SEC_TPL_CONFIG": configFile,
			},
		},
		{
			name: "update secrets from .env file",
			args: []string{
				"secure-templates",
				"manage-secret",
				"import",
				"test",
				"test/secrets.env",
			},
			requiredStrings: []string{
				"8 keys saved on secret",
			},
			envs: map[string]string{
				"SEC_TPL_CONFIG": configFile,
			},
		},
		{
			name: "Env file",
			args: []string{
				"secure-templates",
				"test/samples/.env",
			},
			requiredStrings: []string{
				"APP_USER=dev_user",
				"APP_PASSWORD=2dabe3d7c66fb75f751202fdab19266b",
				"CLIENT_APP_USER=dev_user",
				"CLIENT_APP_PASSWORD=2dabe3d7c66fb75f751202fdab19266b",
			},
			envs: map[string]string{
				"SEC_TPL_CONFIG": configFile,
			},
		},
		{
			name: "k8s secret yaml",
			args: []string{
				"secure-templates",
				"test/samples/k8s-secret.yaml",
			},
			requiredStrings: []string{
				"name: st-secret",
				"namespace: dev-ns",
				"APP_USER: ZGV2X3VzZXI=",
				"APP_PASSWORD: MmRhYmUzZDdjNjZmYjc1Zjc1MTIwMmZkYWIxOTI2NmI=",
				"CLIENT_APP_USER: \"dev_user\"",
				"CLIENT_APP_PASSWORD: \"2dabe3d7c66fb75f751202fdab19266b\"",
			},
			envs: map[string]string{
				"SEC_TPL_CONFIG":   configFile,
				"SECRET_NAME":      "st-secret",
				"SECRET_NAMESPACE": "dev-ns",
			},
		},
		{
			name: "k8s secret yaml - print keys",
			args: []string{
				"secure-templates",
				"-print-keys",
				"test/samples/k8s-secret.yaml",
			},
			requiredStrings: []string{
				"Template keys:",
				"core.app_user",
				"core.app_passwd",
				"client.app_user",
				"client.app_passwd",
			},
			envs: map[string]string{
				"SEC_TPL_CONFIG":   configFile,
				"SECRET_NAME":      "st-secret",
				"SECRET_NAMESPACE": "dev-ns",
			},
		},
		{
			name: "Imported secrets from env file",
			args: []string{
				"secure-templates",
				"test/samples/secrets.env",
			},
			requiredStrings: []string{
				"app1_secret=12345",
				"app2_secret=67890",
				"app3_secret=\"12345\"",
				"app4_secret=\"67\\\"8\\\"90\"",
				"app5_secret='67890'",
				"app6_secret='\"67890\"'",
				"app7_secret=678`90",
				"app8_secret=áçõ",
			},
			envs: map[string]string{
				"SEC_TPL_CONFIG":   configFile,
				"SECRET_NAME":      "st-secret",
				"SECRET_NAMESPACE": "dev-ns",
			},
		},
		{
			name: "Use secret range values",
			args: []string{
				"secure-templates",
				"test/samples/secrets-list.env",
			},
			requiredStrings: []string{
				"app1_secret:12345",
				"app2_secret:67890",
				"app3_secret:\"12345\"",
				"app4_secret:\"67\\\"8\\\"90\"",
				"app5_secret:'67890'",
				"app6_secret:'\"67890\"'",
				"app7_secret:678`90",
				"app8_secret:áçõ",
			},
			envs: map[string]string{
				"SEC_TPL_CONFIG":   configFile,
				"SECRET_NAME":      "st-secret",
				"SECRET_NAMESPACE": "dev-ns",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for key, value := range tt.envs {
				t.Setenv(key, value)
			}
			buf := bytes.Buffer{}
			initApp(tt.args, &buf)
			str := buf.String()
			for _, requiredString := range tt.requiredStrings {
				if !strings.Contains(str, requiredString) {
					t.Fatalf("Required '%s' string not found.", requiredString)
				}
			}
		})
	}
}

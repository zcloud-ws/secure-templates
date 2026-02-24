package test

import (
	"os"
	"testing"
)

func Test_ociVaultApp(t *testing.T) {
	// Skip if OCI credentials are not configured
	if os.Getenv("OCI_COMPARTMENT_OCID") == "" {
		t.Skip("Skipping OCI Vault tests: OCI_COMPARTMENT_OCID not set")
	}

	configFile := "configs/oci-vault-cfg.json"
	tests := []DataTest{
		{
			Name: "oci-vault put secret key",
			Args: []string{
				"secure-templates",
				"manage-secret",
				"put",
				"oci-test-secret",
				"app_user",
				"oci_dev_user",
			},
			RequiredStrings: []string{
				"saved on secret",
			},
			Envs: map[string]string{
				"SEC_TPL_CONFIG": configFile,
			},
		},
		{
			Name: "oci-vault put second key",
			Args: []string{
				"secure-templates",
				"manage-secret",
				"put",
				"oci-test-secret",
				"app_passwd",
				"oci_secret_password_123",
			},
			RequiredStrings: []string{
				"saved on secret",
			},
			Envs: map[string]string{
				"SEC_TPL_CONFIG": configFile,
			},
		},
		{
			Name: "oci-vault import from env file",
			Args: []string{
				"secure-templates",
				"manage-secret",
				"import",
				"oci-test-import",
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
			Name: "oci-vault render env template",
			Args: []string{
				"secure-templates",
				"samples/.env",
			},
			RequiredStrings: []string{
				"APP_USER=oci_dev_user",
				"APP_PASSWORD=oci_secret_password_123",
			},
			Envs: map[string]string{
				"SEC_TPL_CONFIG": configFile,
			},
		},
		{
			Name: "oci-vault print keys",
			Args: []string{
				"secure-templates",
				"-print-keys",
				"samples/k8s-secret.yaml",
			},
			RequiredStrings: []string{
				"Template keys:",
				"core.app_user",
				"core.app_passwd",
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

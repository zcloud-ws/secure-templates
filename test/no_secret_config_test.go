package test

import (
	"testing"
)

func Test_no_secret_config(t *testing.T) {
	tests := []DataTest{
		{
			Name: "Env file",
			Args: []string{
				"secure-templates",
				"samples/template-without-secret.json",
			},
			RequiredStrings: []string{
				"\"userFromEnv\": \"dev_user\"",
				"\"base64value\": \"YjY0LXN0cg==\"",
				"\"no-secret\": \"no-secret\"",
			},
			RequiredErrStrings: []string{
				"Using print key as secret connector",
			},
			Envs: map[string]string{
				"SP_USERNAME": "dev_user",
			},
		},
	}
	SuiteTest(t, "", tests)
}

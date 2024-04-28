package test

import (
	"bytes"
	"github.com/edimarlnx/secure-templates/pkg/app"
	"strings"
	"testing"
)

type DataTest struct {
	Name               string
	Args               []string
	RequiredStrings    []string
	RequiredErrStrings []string
	Envs               map[string]string
}

func SuiteTest(t *testing.T, configFile string, data []DataTest) {
	for _, tt := range data {
		t.Run(tt.Name, func(t *testing.T) {
			for key, value := range tt.Envs {
				t.Setenv(key, value)
			}
			buf := bytes.Buffer{}
			bufErr := bytes.Buffer{}
			app.InitApp(tt.Args, &buf, &bufErr)
			str := buf.String()
			strErr := bufErr.String()
			for _, requiredString := range tt.RequiredStrings {
				if !strings.Contains(str, requiredString) {
					t.Fatalf("Required '%s' string not found.", requiredString)
				}
			}
			for _, requiredErrString := range tt.RequiredErrStrings {
				if !strings.Contains(strErr, requiredErrString) {
					t.Fatalf("Required error '%s' string not found.", requiredErrString)
				}
			}
		})
	}
}

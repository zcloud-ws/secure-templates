package render

import (
	"github.com/edimarlnx/secure-templates/pkg/config"
	"github.com/edimarlnx/secure-templates/pkg/connectors"
	"github.com/edimarlnx/secure-templates/pkg/helpers"
)

func RegisterSecret(connector connectors.Connector) func(args ...string) any {
	return func(args ...string) any {
		if len(args) == 1 {
			return connector.Secret(args[0], "")
		}
		return connector.Secret(args[0], args[1])
	}
}

func RegisterEnvVar(cfgOptions config.SecureTemplateConfigOptions) func(string) string {
	return func(envName string) string {
		if cfgOptions.EnvShowNameAsValueIfEmpty {
			return helpers.GetEnv(envName, envName)
		}
		return helpers.GetEnv(envName, "")
	}
}

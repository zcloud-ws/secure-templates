package render

import (
	"github.com/zcloud-ws/secure-templates/pkg/config"
	"github.com/zcloud-ws/secure-templates/pkg/connectors"
	"github.com/zcloud-ws/secure-templates/pkg/envs"
	"github.com/zcloud-ws/secure-templates/pkg/helpers"
	"github.com/zcloud-ws/secure-templates/pkg/logging"
	"regexp"
	"slices"
)

var restrictedEnvVars []string
var allowEnvVarsRegex *regexp.Regexp

func RegisterSecret(connector connectors.Connector) func(args ...string) any {
	return func(args ...string) any {
		if len(args) == 1 {
			return connector.Secret(args[0], "")
		}
		return connector.Secret(args[0], args[1])
	}
}

func RegisterEnvVar(cfgOptions config.SecureTemplateConfigOptions) func(string) string {
	if !cfgOptions.EnvAllowAccessToSecureTemplateEnvs {
		restrictedEnvVars = []string{
			envs.LocalSecretPrivateKeyPassphraseEnv,
			envs.SecTplConfigEnv,
			envs.SecTplOutputEnv,
			envs.LocalSecretPrivateKeyEnv,
			envs.VaultAddrEnv,
			envs.VaultTokenEnv,
			envs.VaultSecretEngineEnv,
			envs.VaultNsEnv,
		}
	}
	if cfgOptions.EnvRestrictedNameRegex != "" {
		regex, err := regexp.Compile(cfgOptions.EnvRestrictedNameRegex)
		if err != nil {
			logging.Log.Warnf("Error on parse regex: %s\n", cfgOptions.EnvRestrictedNameRegex)
		} else {
			allowEnvVarsRegex = regex
		}
	}

	return func(envName string) string {
		if slices.Contains(restrictedEnvVars, envName) || (allowEnvVarsRegex != nil && !allowEnvVarsRegex.MatchString(envName)) {
			logging.Log.Warnf("'%s' is a restricted variable name.\n", envName)
			return ""
		}
		if cfgOptions.EnvShowNameAsValueIfEmpty {
			return helpers.GetEnv(envName, envName)
		}
		return helpers.GetEnv(envName, "")
	}
}

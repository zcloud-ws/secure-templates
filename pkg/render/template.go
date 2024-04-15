package render

import (
	"github.com/Masterminds/sprig/v3"
	"github.com/edimarlnx/secure-templates/pkg/config"
	"github.com/edimarlnx/secure-templates/pkg/connectors"
	"io"
	"os"
	"path/filepath"
	"text/template"
)

func funcMap(cfgOptions config.SecureTemplateConfigOptions, connector connectors.Connector) template.FuncMap {
	funcMaps := template.FuncMap{}
	for k, v := range sprig.FuncMap() {
		funcMaps[k] = v
	}
	funcMaps["env"] = RegisterEnvVar(cfgOptions)
	funcMaps["secret"] = RegisterSecret(connector)
	return funcMaps
}

func ParseFile(cfgOptions config.SecureTemplateConfigOptions, file *os.File, connector connectors.Connector, output io.Writer) error {
	tpl, err := template.New(filepath.Base(file.Name())).
		Funcs(funcMap(cfgOptions, connector)).
		ParseFiles(file.Name())
	if err != nil {
		return err
	}
	return processTemplate(tpl, output)
}

func processTemplate(tpl *template.Template, output io.Writer) error {
	err := tpl.Execute(output, nil)
	if err != nil {
		return err
	}
	return nil
}

package render

import (
	"github.com/Masterminds/sprig/v3"
	"github.com/edimarlnx/secure-templates/pkg/connectors"
	"io"
	"os"
	"path/filepath"
	"text/template"
)

func funcMap(connector connectors.Connector) template.FuncMap {
	funcMaps := template.FuncMap{}
	for k, v := range sprig.FuncMap() {
		funcMaps[k] = v
	}
	funcMaps["env"] = EnvVar
	funcMaps["secret"] = RegisterSecret(connector)
	return funcMaps
}

func ParseFile(file *os.File, connector connectors.Connector, output io.Writer) error {
	tpl, err := template.New(filepath.Base(file.Name())).
		Funcs(funcMap(connector)).
		ParseFiles(file.Name())
	if err != nil {
		return err
	}
	return ProcessTemplate(tpl, output)
}

func ProcessTemplate(tpl *template.Template, output io.Writer) error {
	err := tpl.Execute(output, nil)
	if err != nil {
		return err
	}
	return nil
}

package render

import (
	"github.com/edimarlnx/secure-templates/pkg/connectors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

func funcMap(connector connectors.Connector) template.FuncMap {
	return template.FuncMap{
		"base64Encode": Base64Encode,
		"base64Decode": Base64Decode,
		"env":          EnvVar,
		"secret":       RegisterSecret(connector),
		"toUpper":      strings.ToUpper,
		"toLower":      strings.ToLower,
		"trimSpace":    strings.TrimSpace,
	}
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

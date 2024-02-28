package template

import (
	"bytes"
	"fmt"
	"gopkg.in/yaml.v3"
	"strings"
	"text/template"
)

func Parse(name string, contents string, templateData any) (string, error) {
	if templateData == nil {
		return "", fmt.Errorf("template data not provided")
	}

	funcs := template.FuncMap{"join": strings.Join, "toYaml": toYAML}

	tmpl, err := template.New(name).Funcs(funcs).Parse(contents)
	if err != nil {
		return "", fmt.Errorf("parsing contents: %w", err)
	}

	var buff bytes.Buffer
	if err = tmpl.Execute(&buff, templateData); err != nil {
		return "", fmt.Errorf("applying template: %w", err)
	}

	return buff.String(), nil
}

func toYAML(v interface{}) string {
	var buf bytes.Buffer
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)

	if err := encoder.Encode(v); err != nil {
		return ""
	}
	encoder.Close()

	return strings.TrimSuffix(buf.String(), "\n")
}

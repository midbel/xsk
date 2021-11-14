package maestro

import (
	"strings"
	"text/template"
)

const helptext = `
{{if .Help -}}
{{.Help}}
{{- end}}

Available commands:
{{range $k, $cs := .Commands}}
{{$k}}:
{{repeat "-" $k}}-
{{- range $cs}}
  - {{printf "%-20s %s" .Name .Short -}}
{{end}}
{{end}}
{{.File}} ({{.Version}})
`

const cmdhelp = `
{{.Command}}: {{.About}}

{{.Desc}}

{{.Usage}}

tag(s): {{join .Tags ", "}}
`

func renderTemplate(name string, ctx interface{}) (string, error) {
	t, err := template.New("template").Funcs(funcmap).Parse(name)
	if err != nil {
		return "", err
	}
	var str strings.Builder
	if err := t.Execute(&str, ctx); err != nil {
		return "", err
	}
	return str.String(), nil
}

var funcmap = template.FuncMap{
	"repeat": repeat,
	"join":   strings.Join,
}

func repeat(char string, value interface{}) string {
	var n int
	switch v := value.(type) {
	case string:
		n = len(v)
	case int:
		n = v
	default:
		return ""
	}
	return strings.Repeat(char, n)
}
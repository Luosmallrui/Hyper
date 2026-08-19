package utils

import (
	"bytes"
	"html/template"
)

func RenderTemplate(text []byte, data any) (string, error) {
	tmpl, err := template.New("tmpl").Parse(string(text))
	if err != nil {
		return "", err
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		return "", err
	}

	return body.String(), nil
}

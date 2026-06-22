package email

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
)

//go:embed templates/*.html
var templateFS embed.FS

var templates = template.Must(
	template.ParseFS(templateFS, "templates/*.html"),
)

type templateData struct {
	ActionURL string
	Timestamp string
}

// render executes the named template with the given data and returns the
// resulting HTML string.
func render(name string, data any) (string, error) {
	var buf bytes.Buffer
	if err := templates.ExecuteTemplate(&buf, name, data); err != nil {
		return "", fmt.Errorf("render email template %q: %w", name, err)
	}

	return buf.String(), nil
}

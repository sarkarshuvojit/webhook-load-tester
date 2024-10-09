package templates

import (
	"embed"
	"os"
)

var _ embed.FS

//go:embed default-test-template.yml
var templateContent string

func CreateTemplate(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	f.WriteString(templateContent)

	return nil
}

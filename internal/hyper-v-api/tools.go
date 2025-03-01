package hypervapi

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"text/template"

	"github.com/masterzen/winrm"
)

func renderTemplate(templateFile string, data any) (*bytes.Buffer, error) {
	tmpl, err := template.ParseFS(templatesFS, "templates/"+templateFile)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template %s: %w", templateFile, err)
	}

	var script bytes.Buffer
	if err := tmpl.Execute(&script, data); err != nil {
		return nil, fmt.Errorf("failed to execute template %s: %w", templateFile, err)
	}

	return &script, nil
}

func runRemoteCommand(ctx context.Context, client *winrm.Client, script string) (string, string, int, error) {
	// TODO: Move error logging to this level for all scripts
	log.Println("Running Remote command")
	stdout, stderr, exitCode, err := client.RunPSWithContext(ctx, script)
	log.Println(script)
	if err != nil {
		return "", "", 0, fmt.Errorf("failed to run remote command: %w", err)
	}

	if exitCode != 0 {
		return "", stderr, exitCode, nil
	}

	return stdout, stderr, exitCode, nil
}

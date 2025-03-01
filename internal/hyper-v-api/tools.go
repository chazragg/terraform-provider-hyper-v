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

func runRemoteCommand(ctx context.Context, client *winrm.Client, script string) (string, error) {
	log.Println("Running Remote command")
	stdout, stderr, exitCode, err := client.RunPSWithContext(ctx, script)
	if err != nil {
		return "", fmt.Errorf("failed to run remote command: %w", err)
	}

	if exitCode != 0 {
		if stderr != "" {
			return "", fmt.Errorf("PowerShell script error: %s", stderr)
		}
		return "", fmt.Errorf("PowerShell script exited with code %d", exitCode)
	}

	return stdout, nil
}

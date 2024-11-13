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
		log.Fatalf("Error Parsing FS: %v", err)
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	var script bytes.Buffer
	if err := tmpl.Execute(&script, data); err != nil {
		log.Fatalf("Error Executing Template: %v", err)
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	return &script, nil
}

func runRemoteCommand(ctx context.Context, client *winrm.Client, script string) (string, string, int, error) {
	log.Println("Running Remote command")
	stdout, stderr, exitCode, err := client.RunPSWithContext(ctx, script)
	if err != nil {
		return "", "", 0, fmt.Errorf("failed to run remote command: %w", err)
	}

	return stdout, stderr, exitCode, nil
}

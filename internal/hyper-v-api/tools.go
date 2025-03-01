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
	log.Println("Executing Remote PowerShell command via WinRM...")

	stdout, stderr, exitCode, err := client.RunPSWithContext(ctx, script)

	if err != nil {
		log.Printf("Command execution failed: %v\n", err)
		log.Printf("Exit Code: %d\nStderr: %s\n", exitCode, stderr)
		return "", "", 0, fmt.Errorf("failed to run remote command: %w", err)
	}

	log.Printf("Command executed successfully with exit code %d\n", exitCode)
	if stderr != "" {
		log.Printf("Command stderr: %s\n", stderr)
	}

	return stdout, stderr, exitCode, nil
}

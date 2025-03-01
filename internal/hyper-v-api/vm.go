package hypervapi

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
)

type VMModel struct {
	VMId               string `json:"VMId,omitempty"`
	Name               string `json:"VMName,omitempty"`
	Generation         int64  `json:"Generation,omitempty"`
	MemoryStartupBytes int64  `json:"MemoryStartup,omitempty"`
	Path               string `json:"Path,omitempty"`
	BootDevice         string `json:"BootDevice,omitempty"`
}

func (c *Client) CreateVM(ctx context.Context, data VMModel) (*VMModel, error) {
	// Ensure we have a connected WinRM client
	if err := c.Connect(); err != nil {
		return nil, fmt.Errorf("failed to connect to WinRM: %w", err)
	}

	// Double-check that c.winrmClient is not nil after connection
	if c.winrmClient == nil {
		return nil, fmt.Errorf("WinRM client is nil after connection")
	}
	log.Println("Created new Client")

	// Render PowerShell script
	script, err := renderTemplate("CreateVirtualMachine.ps1.tmpl", data)
	if err != nil {
		return nil, fmt.Errorf("failed to render PowerShell script: %w", err)
	}
	log.Println("Rendered template")

	// Run command on remote system using c.winrmClient
	outWriter, errWrite, exitCode, err := runRemoteCommand(ctx, c.winrmClient, script.String())
	if err != nil {
		return nil, fmt.Errorf("failed to execute remote command: %w", err)
	}
	if errWrite != "" {
		return nil, fmt.Errorf("PowerShell script error: %s", errWrite)
	}
	if exitCode != 0 {
		return nil, fmt.Errorf("PowerShell script exited with code %d", exitCode)
	}

	// Parse command output to VMModel
	var vmResult VMModel
	outWriter = strings.TrimSpace(outWriter) // Remove trailing newlines
	if err := json.Unmarshal([]byte(outWriter), &vmResult); err != nil {
		return nil, fmt.Errorf("failed to parse command output: %w", err)
	}

	return &vmResult, nil
}

func (c *Client) DeleteVM(ctx context.Context, data VMModel) error {
	if err := c.Connect(); err != nil {
		return fmt.Errorf("failed to connect to WinRM: %w", err)
	}

	if c.winrmClient == nil {
		return fmt.Errorf("WinRM client is nil after connection")
	}

	log.Println("Created new Client")

	// Render PowerShell script
	script, err := renderTemplate("DeleteVirtualMachine.ps1.tmpl", data)
	if err != nil {
		log.Fatalf("Error Rendering tempalte: %v", err)
		log.Println(script)
		return fmt.Errorf("failed to render PowerShell script: %w", err)
	}
	log.Println("Rendered template")
	log.Print(script.String())

	// Run command on remote system using c.winrmClient
	outWriter, errWrite, exitCode, err := runRemoteCommand(ctx, c.winrmClient, script.String())
	if err != nil {
		log.Fatalln(err)
		log.Fatalln(errWrite)
		return fmt.Errorf("failed to execute remote command: %w", err)
	}
	if errWrite != "" {
		log.Fatalln(errWrite)
		return fmt.Errorf("failed to execute remote command: %s", errWrite)
	}
	if exitCode != 0 {
		log.Fatalf("Exit Code: exitCode")
		return fmt.Errorf("script exited with code %d", exitCode)
	}

	log.Print(outWriter)
	return nil
}

func (c *Client) GetVM(ctx context.Context, data VMModel) (*VMModel, error) {
	if err := c.Connect(); err != nil {
		return nil, fmt.Errorf("failed to connect to WinRM: %w", err)
	}

	if c.winrmClient == nil {
		return nil, fmt.Errorf("WinRM client is nil afetr connection")
	}

	log.Println("Created new client")

	script, err := renderTemplate("GetVirtualMachine.ps1.tmpl", data)
	if err != nil {
		return nil, fmt.Errorf("failed to execute remote command: %w", err)
	}

	outWriter, _, _, err := runRemoteCommand(ctx, c.winrmClient, script.String())
	if err != nil {
		return nil, fmt.Errorf("failed to execute remote command: %w", err)
	}

	// Parse command output to VMModel
	var vmResult VMModel
	if err := json.Unmarshal([]byte(outWriter), &vmResult); err != nil {
		return nil, fmt.Errorf("failed to parse command output: %w", err)
	}
	return &vmResult, nil
}

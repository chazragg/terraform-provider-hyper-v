package hypervapi

import (
	"context"
	"encoding/json"
	"fmt"
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

	// Render PowerShell script
	script, err := renderTemplate("CreateVirtualMachine.ps1.tmpl", data)
	if err != nil {
		return nil, fmt.Errorf("failed to render PowerShell script: %w", err)
	}

	// Run command on remote system using c.winrmClient
	outWriter, err := runRemoteCommand(ctx, c.winrmClient, script.String())
	if err != nil {
		return nil, fmt.Errorf("failed to execute remote command: %w", err)
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
	// Ensure we have a connected WinRM client
	if err := c.Connect(); err != nil {
		return fmt.Errorf("failed to connect to WinRM: %w", err)
	}

	// Double-check that c.winrmClient is not nil after connection
	if c.winrmClient == nil {
		return fmt.Errorf("WinRM client is nil after connection")
	}

	// Render PowerShell script
	script, err := renderTemplate("DeleteVirtualMachine.ps1.tmpl", data)
	if err != nil {
		return fmt.Errorf("failed to render PowerShell script: %w", err)
	}

	// Run command on remote system using c.winrmClient
	_, err = runRemoteCommand(ctx, c.winrmClient, script.String())
	if err != nil {
		return fmt.Errorf("failed to execute remote command: %w", err)
	}

	return nil
}

func (c *Client) GetVM(ctx context.Context, data VMModel) (*VMModel, error) {
	if err := c.Connect(); err != nil {
		return nil, fmt.Errorf("failed to connect to WinRM: %w", err)
	}

	if c.winrmClient == nil {
		return nil, fmt.Errorf("WinRM client is nil afetr connection")
	}

	script, err := renderTemplate("GetVirtualMachine.ps1.tmpl", data)
	if err != nil {
		return nil, fmt.Errorf("failed to execute remote command: %w", err)
	}

	outWriter, err := runRemoteCommand(ctx, c.winrmClient, script.String())
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

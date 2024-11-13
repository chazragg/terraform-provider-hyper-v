package hypervapi

import (
	"fmt"
	"time"

	"github.com/masterzen/winrm"
)

type Client struct {
	Host          string
	Port          int
	Username      string
	Password      string
	HTTPS         bool
	Insecure      bool
	TLSServerName string
	CACert        []byte
	CAKey         []byte
	Cert          []byte
	Timeout       time.Duration

	winrmClient *winrm.Client // Add this field to store the initialized WinRM client
}

func (c *Client) Connect() error {
	endpoint := winrm.NewEndpoint(c.Host, c.Port, c.HTTPS, c.Insecure, c.CACert, c.Cert, c.CAKey, c.Timeout)
	client, err := winrm.NewClient(endpoint, c.Username, c.Password)
	if err != nil {
		return fmt.Errorf("failed to create WinRM client: %w", err)
	}
	c.winrmClient = client
	return nil
}

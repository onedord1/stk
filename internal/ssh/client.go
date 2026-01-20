package ssh

import (
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

// Client manages SSH connections
type Client struct {
	connections map[string]*ssh.Client
	mu          sync.RWMutex
	timeout     time.Duration
}

// NewClient creates a new SSH client manager
func NewClient(timeout time.Duration) *Client {
	return &Client{
		connections: make(map[string]*ssh.Client),
		timeout:     timeout,
	}
}

// Connect establishes an SSH connection to a host
func (c *Client) Connect(host HostEntry) (*ssh.Client, error) {
	key := fmt.Sprintf("%s:%d", host.Hostname, host.Port)

	// Check for existing connection
	c.mu.RLock()
	if conn, ok := c.connections[key]; ok {
		c.mu.RUnlock()
		// Test if connection is still alive
		if _, _, err := conn.SendRequest("keepalive@systask", true, nil); err == nil {
			return conn, nil
		}
		// Connection dead, remove it
		c.mu.Lock()
		delete(c.connections, key)
		c.mu.Unlock()
	} else {
		c.mu.RUnlock()
	}

	// Build auth methods
	authMethods := c.getAuthMethods(host)
	if len(authMethods) == 0 {
		return nil, fmt.Errorf("no authentication methods available")
	}

	// Determine user
	user := host.User
	if user == "" {
		user = os.Getenv("USER")
	}

	config := &ssh.ClientConfig{
		User:            user,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // TODO: Implement proper host key verification
		Timeout:         c.timeout,
	}

	addr := fmt.Sprintf("%s:%d", host.Hostname, host.Port)
	if host.Hostname == "" {
		addr = fmt.Sprintf("%s:%d", host.Name, host.Port)
	}

	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", addr, err)
	}

	// Cache the connection
	c.mu.Lock()
	c.connections[key] = client
	c.mu.Unlock()

	return client, nil
}

// getAuthMethods returns available authentication methods
func (c *Client) getAuthMethods(host HostEntry) []ssh.AuthMethod {
	var methods []ssh.AuthMethod

	// Try SSH agent first
	if agentAuth := c.getAgentAuth(); agentAuth != nil {
		methods = append(methods, agentAuth)
	}

	// Try specific key file
	if host.KeyFile != "" {
		if keyAuth := c.getKeyAuth(host.KeyFile); keyAuth != nil {
			methods = append(methods, keyAuth)
		}
	}

	// Try default key files
	homeDir, _ := os.UserHomeDir()
	defaultKeys := []string{
		filepath.Join(homeDir, ".ssh", "id_ed25519"),
		filepath.Join(homeDir, ".ssh", "id_rsa"),
		filepath.Join(homeDir, ".ssh", "id_ecdsa"),
	}

	for _, keyPath := range defaultKeys {
		if keyAuth := c.getKeyAuth(keyPath); keyAuth != nil {
			methods = append(methods, keyAuth)
		}
	}

	return methods
}

// getAgentAuth returns SSH agent authentication
func (c *Client) getAgentAuth() ssh.AuthMethod {
	socket := os.Getenv("SSH_AUTH_SOCK")
	if socket == "" {
		return nil
	}

	conn, err := net.Dial("unix", socket)
	if err != nil {
		return nil
	}

	agentClient := agent.NewClient(conn)
	return ssh.PublicKeysCallback(agentClient.Signers)
}

// getKeyAuth returns key-based authentication
func (c *Client) getKeyAuth(keyPath string) ssh.AuthMethod {
	key, err := os.ReadFile(keyPath)
	if err != nil {
		return nil
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		// Key might be encrypted, skip for now
		return nil
	}

	return ssh.PublicKeys(signer)
}

// Disconnect closes a specific connection
func (c *Client) Disconnect(host HostEntry) error {
	key := fmt.Sprintf("%s:%d", host.Hostname, host.Port)

	c.mu.Lock()
	defer c.mu.Unlock()

	if conn, ok := c.connections[key]; ok {
		delete(c.connections, key)
		return conn.Close()
	}
	return nil
}

// DisconnectAll closes all connections
func (c *Client) DisconnectAll() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for key, conn := range c.connections {
		conn.Close()
		delete(c.connections, key)
	}
}

// IsConnected checks if a host is connected
func (c *Client) IsConnected(host HostEntry) bool {
	key := fmt.Sprintf("%s:%d", host.Hostname, host.Port)

	c.mu.RLock()
	conn, ok := c.connections[key]
	c.mu.RUnlock()

	if !ok {
		return false
	}

	// Test connection
	_, _, err := conn.SendRequest("keepalive@systask", true, nil)
	return err == nil
}

// GetSession creates a new SSH session for a host
func (c *Client) GetSession(host HostEntry) (*ssh.Session, error) {
	client, err := c.Connect(host)
	if err != nil {
		return nil, err
	}
	return client.NewSession()
}

// RunCommand executes a command and returns output
func (c *Client) RunCommand(host HostEntry, command string) (string, error) {
	session, err := c.GetSession(host)
	if err != nil {
		return "", err
	}
	defer session.Close()

	output, err := session.CombinedOutput(command)
	return string(output), err
}

// RunCommandStream executes a command with streaming output
func (c *Client) RunCommandStream(host HostEntry, command string, stdout, stderr io.Writer) error {
	session, err := c.GetSession(host)
	if err != nil {
		return err
	}
	defer session.Close()

	session.Stdout = stdout
	session.Stderr = stderr

	return session.Run(command)
}

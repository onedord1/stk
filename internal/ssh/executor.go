package ssh

import (
	"bytes"
	"context"
	"fmt"
	"sync"
)

// ExecutionResult holds the result of a command execution
type ExecutionResult struct {
	Host    HostEntry
	Output  string
	Error   error
	Success bool
}

// Executor handles batch command execution
type Executor struct {
	client      *Client
	maxParallel int
}

// NewExecutor creates a new batch executor
func NewExecutor(client *Client, maxParallel int) *Executor {
	return &Executor{
		client:      client,
		maxParallel: maxParallel,
	}
}

// ExecuteOnHosts runs a command on multiple hosts in parallel
func (e *Executor) ExecuteOnHosts(ctx context.Context, hosts []HostEntry, command string) []ExecutionResult {
	results := make([]ExecutionResult, len(hosts))

	// Semaphore for limiting parallelism
	sem := make(chan struct{}, e.maxParallel)
	var wg sync.WaitGroup

	for i, host := range hosts {
		wg.Add(1)
		go func(idx int, h HostEntry) {
			defer wg.Done()

			// Acquire semaphore
			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-ctx.Done():
				results[idx] = ExecutionResult{
					Host:    h,
					Error:   ctx.Err(),
					Success: false,
				}
				return
			}

			// Execute command
			output, err := e.client.RunCommand(h, command)
			results[idx] = ExecutionResult{
				Host:    h,
				Output:  output,
				Error:   err,
				Success: err == nil,
			}
		}(i, host)
	}

	wg.Wait()
	return results
}

// ExecuteWithCallback runs a command with real-time result callback
func (e *Executor) ExecuteWithCallback(ctx context.Context, hosts []HostEntry, command string, callback func(ExecutionResult)) {
	sem := make(chan struct{}, e.maxParallel)
	var wg sync.WaitGroup

	for _, host := range hosts {
		wg.Add(1)
		go func(h HostEntry) {
			defer wg.Done()

			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-ctx.Done():
				callback(ExecutionResult{
					Host:    h,
					Error:   ctx.Err(),
					Success: false,
				})
				return
			}

			output, err := e.client.RunCommand(h, command)
			callback(ExecutionResult{
				Host:    h,
				Output:  output,
				Error:   err,
				Success: err == nil,
			})
		}(host)
	}

	wg.Wait()
}

// StreamResult holds streaming output result
type StreamResult struct {
	Host   HostEntry
	Stdout *bytes.Buffer
	Stderr *bytes.Buffer
	Error  error
}

// ExecuteStream runs a command with streaming output buffers
func (e *Executor) ExecuteStream(ctx context.Context, host HostEntry, command string) *StreamResult {
	result := &StreamResult{
		Host:   host,
		Stdout: new(bytes.Buffer),
		Stderr: new(bytes.Buffer),
	}

	result.Error = e.client.RunCommandStream(host, command, result.Stdout, result.Stderr)
	return result
}

// TestConnection tests if a host is reachable
func (e *Executor) TestConnection(host HostEntry) error {
	_, err := e.client.RunCommand(host, "echo ok")
	return err
}

// GetSystemInfo retrieves basic system information
func (e *Executor) GetSystemInfo(host HostEntry) (map[string]string, error) {
	info := make(map[string]string)

	// Get hostname
	if output, err := e.client.RunCommand(host, "hostname"); err == nil {
		info["hostname"] = trimOutput(output)
	}

	// Get OS info
	if output, err := e.client.RunCommand(host, "cat /etc/os-release 2>/dev/null | grep PRETTY_NAME | cut -d= -f2 | tr -d '\"'"); err == nil {
		info["os"] = trimOutput(output)
	}

	// Get kernel
	if output, err := e.client.RunCommand(host, "uname -r"); err == nil {
		info["kernel"] = trimOutput(output)
	}

	// Get uptime
	if output, err := e.client.RunCommand(host, "uptime -p 2>/dev/null || uptime"); err == nil {
		info["uptime"] = trimOutput(output)
	}

	return info, nil
}

func trimOutput(s string) string {
	return string(bytes.TrimSpace([]byte(s)))
}

// DetectPackageManager detects the system's package manager
func (e *Executor) DetectPackageManager(host HostEntry) (string, error) {
	checks := []struct {
		command string
		manager string
	}{
		{"which apt 2>/dev/null", "apt"},
		{"which dnf 2>/dev/null", "dnf"},
		{"which yum 2>/dev/null", "yum"},
		{"which zypper 2>/dev/null", "zypper"},
		{"which pacman 2>/dev/null", "pacman"},
		{"which apk 2>/dev/null", "apk"},
	}

	for _, check := range checks {
		if output, err := e.client.RunCommand(host, check.command); err == nil && len(output) > 0 {
			return check.manager, nil
		}
	}

	return "", fmt.Errorf("no supported package manager found")
}

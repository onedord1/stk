package ssh

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// HostEntry represents a parsed SSH host
type HostEntry struct {
	Name      string
	Hostname  string
	Port      int
	User      string
	KeyFile   string
	ProxyJump string
	Source    string // "config" or "known_hosts"
}

// ParseSSHConfig parses ~/.ssh/config file
func ParseSSHConfig(configPath string) ([]HostEntry, error) {
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var hosts []HostEntry
	var currentHost *HostEntry

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse key-value pairs
		parts := strings.SplitN(line, " ", 2)
		if len(parts) < 2 {
			parts = strings.SplitN(line, "\t", 2)
		}
		if len(parts) < 2 {
			continue
		}

		key := strings.ToLower(strings.TrimSpace(parts[0]))
		value := strings.TrimSpace(parts[1])

		switch key {
		case "host":
			// Skip wildcards
			if strings.Contains(value, "*") {
				currentHost = nil
				continue
			}

			if currentHost != nil {
				hosts = append(hosts, *currentHost)
			}
			currentHost = &HostEntry{
				Name:   value,
				Port:   22,
				Source: "config",
			}
		case "hostname":
			if currentHost != nil {
				currentHost.Hostname = value
			}
		case "port":
			if currentHost != nil {
				if port, err := strconv.Atoi(value); err == nil {
					currentHost.Port = port
				}
			}
		case "user":
			if currentHost != nil {
				currentHost.User = value
			}
		case "identityfile":
			if currentHost != nil {
				// Expand ~ to home directory
				if strings.HasPrefix(value, "~") {
					home, _ := os.UserHomeDir()
					value = filepath.Join(home, value[1:])
				}
				currentHost.KeyFile = value
			}
		case "proxyjump":
			if currentHost != nil {
				currentHost.ProxyJump = value
			}
		}
	}

	// Don't forget the last host
	if currentHost != nil {
		hosts = append(hosts, *currentHost)
	}

	return hosts, scanner.Err()
}

// ParseKnownHosts parses ~/.ssh/known_hosts file
func ParseKnownHosts(knownHostsPath string) ([]HostEntry, error) {
	file, err := os.Open(knownHostsPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var hosts []HostEntry
	seen := make(map[string]bool)

	// Regex to match host entries (handles [host]:port and plain host)
	hostPortRegex := regexp.MustCompile(`^\[([^\]]+)\]:(\d+)`)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "@") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}

		hostField := fields[0]
		// Handle comma-separated hosts (e.g., "host1,host2")
		hostnames := strings.Split(hostField, ",")

		for _, hostname := range hostnames {
			hostname = strings.TrimSpace(hostname)

			// Skip hashed entries
			if strings.HasPrefix(hostname, "|1|") {
				continue
			}

			var host HostEntry
			host.Source = "known_hosts"
			host.Port = 22

			// Check for [host]:port format
			if matches := hostPortRegex.FindStringSubmatch(hostname); matches != nil {
				host.Hostname = matches[1]
				if port, err := strconv.Atoi(matches[2]); err == nil {
					host.Port = port
				}
			} else {
				host.Hostname = hostname
			}

			// Try to determine if it's an IP or hostname
			host.Name = host.Hostname
			if net.ParseIP(host.Hostname) != nil {
				// It's an IP, use as-is
				host.Name = fmt.Sprintf("host-%s", host.Hostname)
			}

			// Skip duplicates
			key := fmt.Sprintf("%s:%d", host.Hostname, host.Port)
			if seen[key] {
				continue
			}
			seen[key] = true

			hosts = append(hosts, host)
		}
	}

	return hosts, scanner.Err()
}

// DiscoverHosts discovers all SSH hosts from config and known_hosts
func DiscoverHosts(configPath, knownHostsPath string) ([]HostEntry, error) {
	var allHosts []HostEntry
	seen := make(map[string]bool)

	// Parse SSH config (higher priority)
	configHosts, err := ParseSSHConfig(configPath)
	if err == nil {
		for _, host := range configHosts {
			key := host.Name
			if !seen[key] {
				allHosts = append(allHosts, host)
				seen[key] = true
			}
		}
	}

	// Parse known_hosts
	knownHosts, err := ParseKnownHosts(knownHostsPath)
	if err == nil {
		for _, host := range knownHosts {
			key := host.Hostname
			if !seen[key] {
				allHosts = append(allHosts, host)
				seen[key] = true
			}
		}
	}

	return allHosts, nil
}

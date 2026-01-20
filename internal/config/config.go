package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config holds all application configuration
type Config struct {
	Theme           string      `yaml:"theme"`
	RefreshInterval int         `yaml:"refresh_interval"` // seconds
	SSHTimeout      int         `yaml:"ssh_timeout"`      // seconds
	MaxParallel     int         `yaml:"max_parallel"`     // max parallel SSH connections
	SSHConfigPath   string      `yaml:"ssh_config_path"`
	KnownHostsPath  string      `yaml:"known_hosts_path"`
	LogLevel        string      `yaml:"log_level"`
	Hosts           []Host      `yaml:"hosts"`    // manually configured hosts
	Groups          []HostGroup `yaml:"groups"`   // host groups
	KeysDir         string      `yaml:"keys_dir"` // directory for PEM keys
}

// Host represents a server configuration
type Host struct {
	Name     string   `yaml:"name"`
	Hostname string   `yaml:"hostname"`
	Port     int      `yaml:"port"`
	User     string   `yaml:"user"`
	KeyPath  string   `yaml:"key_path"`           // Path to PEM/key file
	Password string   `yaml:"password,omitempty"` // Optional password
	Tags     []string `yaml:"tags"`
	Provider string   `yaml:"provider"`  // aws, azure, gcp, local, custom
	Group    string   `yaml:"group"`     // Group name
	AuthType string   `yaml:"auth_type"` // key, password, agent
}

// HostGroup represents a group of hosts
type HostGroup struct {
	Name     string `yaml:"name"`
	Icon     string `yaml:"icon"`     // Emoji icon
	Color    string `yaml:"color"`    // Hex color
	Expanded bool   `yaml:"expanded"` // UI state
}

// Default returns default configuration
func Default() *Config {
	homeDir, _ := os.UserHomeDir()
	return &Config{
		Theme:           "catppuccin",
		RefreshInterval: 2,
		SSHTimeout:      10,
		MaxParallel:     5,
		SSHConfigPath:   filepath.Join(homeDir, ".ssh", "config"),
		KnownHostsPath:  filepath.Join(homeDir, ".ssh", "known_hosts"),
		LogLevel:        "info",
		Hosts:           []Host{},
		Groups: []HostGroup{
			{Name: "Production", Icon: "🔴", Expanded: true},
			{Name: "Staging", Icon: "🟡", Expanded: true},
			{Name: "Development", Icon: "🟢", Expanded: true},
			{Name: "Cloud", Icon: "☁️", Expanded: true},
			{Name: "Local", Icon: "🖥️", Expanded: true},
		},
		KeysDir: filepath.Join(homeDir, ".ssh"),
	}
}

// Load reads configuration from file
func Load() (*Config, error) {
	cfg := Default()

	configDir, err := os.UserConfigDir()
	if err != nil {
		return cfg, err
	}

	configPath := filepath.Join(configDir, "systask", "config.yaml")

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			_ = createSampleConfig(configPath)
			return cfg, nil
		}
		return cfg, err
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}

// Save writes configuration to file
func (c *Config) Save() error {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return err
	}

	systaskDir := filepath.Join(configDir, "systask")
	if err := os.MkdirAll(systaskDir, 0755); err != nil {
		return err
	}

	configPath := filepath.Join(systaskDir, "config.yaml")

	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

// createSampleConfig creates a sample configuration file
func createSampleConfig(path string) error {
	sample := `# SysTask Configuration
# Theme: catppuccin, dracula, nord, gruvbox, solarized, tokyo-night
theme: catppuccin

# SSH settings
ssh_timeout: 10
max_parallel: 5
keys_dir: ~/.ssh

# Host Groups (customize as needed)
groups:
  - name: Production
    icon: "🔴"
    expanded: true
  - name: Staging
    icon: "🟡"
    expanded: true
  - name: Development
    icon: "🟢"
    expanded: true
  - name: Cloud
    icon: "☁️"
    expanded: true

# Manually configured hosts
hosts:
  # SSH Key Authentication (recommended)
  # - name: my-server
  #   hostname: 192.168.1.100
  #   port: 22
  #   user: admin
  #   key_path: ~/.ssh/id_rsa
  #   group: Production
  #   auth_type: key

  # Cloud VM with PEM key
  # - name: aws-web-server
  #   hostname: ec2-xx-xx-xx-xx.compute-1.amazonaws.com
  #   user: ec2-user
  #   key_path: ~/.ssh/my-key.pem
  #   provider: aws
  #   group: Cloud
  #   auth_type: key

  # Password authentication (not recommended)
  # - name: legacy-server
  #   hostname: 10.0.0.5
  #   user: root
  #   password: secret123
  #   group: Development
  #   auth_type: password
`

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(path, []byte(sample), 0644)
}

// AddHost adds a new host to the configuration
func (c *Config) AddHost(host Host) {
	c.Hosts = append(c.Hosts, host)
}

// RemoveHost removes a host by name
func (c *Config) RemoveHost(name string) {
	for i, h := range c.Hosts {
		if h.Name == name {
			c.Hosts = append(c.Hosts[:i], c.Hosts[i+1:]...)
			return
		}
	}
}

// GetHost returns a host by name
func (c *Config) GetHost(name string) *Host {
	for i := range c.Hosts {
		if c.Hosts[i].Name == name {
			return &c.Hosts[i]
		}
	}
	return nil
}

// GetHostsByGroup returns hosts in a specific group
func (c *Config) GetHostsByGroup(group string) []Host {
	var hosts []Host
	for _, h := range c.Hosts {
		if h.Group == group {
			hosts = append(hosts, h)
		}
	}
	return hosts
}

// AddGroup adds a new host group
func (c *Config) AddGroup(group HostGroup) {
	c.Groups = append(c.Groups, group)
}

// GetGroup returns a group by name
func (c *Config) GetGroup(name string) *HostGroup {
	for i := range c.Groups {
		if c.Groups[i].Name == name {
			return &c.Groups[i]
		}
	}
	return nil
}

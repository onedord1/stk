# SysTask Installation Guide

## Requirements

- **Operating System**: Linux (tested on Ubuntu 20.04+, Debian 10+, RHEL/CentOS 8+)
- **Terminal**: Any modern terminal with Unicode support (recommended: Alacritty, Kitty, iTerm2)
- **SSH Access**: SSH keys or credentials for remote servers

## Installation Methods

### Method 1: Binary Download (Recommended)

Download the pre-built binary for your platform:

```bash
# Linux AMD64 (most common)
curl -LO https://github.com/yourusername/systask/releases/latest/download/systask-linux-amd64
chmod +x systask-linux-amd64
sudo mv systask-linux-amd64 /usr/local/bin/systask

# Linux ARM64 (Raspberry Pi, ARM servers)
curl -LO https://github.com/yourusername/systask/releases/latest/download/systask-linux-arm64
chmod +x systask-linux-arm64
sudo mv systask-linux-arm64 /usr/local/bin/systask

# macOS AMD64
curl -LO https://github.com/yourusername/systask/releases/latest/download/systask-darwin-amd64
chmod +x systask-darwin-amd64
sudo mv systask-darwin-amd64 /usr/local/bin/systask

# macOS ARM64 (Apple Silicon)
curl -LO https://github.com/yourusername/systask/releases/latest/download/systask-darwin-arm64
chmod +x systask-darwin-arm64
sudo mv systask-darwin-arm64 /usr/local/bin/systask
```

### Method 2: Build from Source

Requires Go 1.21 or higher.

```bash
# Install Go (if not already installed)
# Visit https://go.dev/doc/install

# Clone repository
git clone https://github.com/yourusername/systask.git
cd systask

# Build
go build -o systask ./main.go

# Optional: Install globally
sudo mv systask /usr/local/bin/
```

### Method 3: Using Makefile

```bash
git clone https://github.com/yourusername/systask.git
cd systask
make build
make install  # Installs to $GOPATH/bin
```

## Post-Installation

### 1. Verify Installation

```bash
systask --version
```

### 2. Create Configuration (Optional)

SysTask works out-of-the-box by reading `~/.ssh/config`. For custom hosts:

```bash
mkdir -p ~/.config/systask
cp /path/to/systask/config.yaml.example ~/.config/systask/config.yaml
```

### 3. Run SysTask

```bash
systask
```

## Troubleshooting

### "command not found"
Ensure the binary is in your PATH:
```bash
export PATH=$PATH:/usr/local/bin
# Add to ~/.bashrc or ~/.zshrc for persistence
```

### Unicode characters not displaying
Use a terminal with Unicode and emoji support. Install a Nerd Font:
```bash
# Example: Install Hack Nerd Font
wget https://github.com/ryanoasis/nerd-fonts/releases/download/v3.0.0/Hack.zip
unzip Hack.zip -d ~/.local/share/fonts
fc-cache -fv
```

### SSH connection issues
1. Verify SSH works directly: `ssh user@host`
2. Check key permissions: `chmod 600 ~/.ssh/id_rsa`
3. Ensure ssh-agent is running: `eval $(ssh-agent)`

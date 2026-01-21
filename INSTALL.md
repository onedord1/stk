# Installation Guide

SysTask can be installed on Linux systems using your distribution's package manager or from source.

## Quick Install

### Debian/Ubuntu
```bash
sudo apt update
sudo apt install systask
```

### Fedora/RHEL/CentOS
```bash
sudo dnf install systask
# or for older systems
sudo yum install systask
```

### Arch Linux
```bash
sudo pacman -S systask
```

### Alpine Linux
```bash
sudo apk add systask
```

---

## Manual Installation from Source

### Prerequisites
- Go 1.16 or later
- OpenSSH client
- Git

### Build and Install
```bash
# Clone the repository
git clone https://github.com/onedord1/stk.git
cd stk

# Build the binary
go build -o systask ./main.go

# Install to system
sudo install -Dm755 systask /usr/local/bin/systask

# Create config directory
mkdir -p ~/.config/systask
cp config.yaml.example ~/.config/systask/config.yaml
```

### Run
```bash
systask
```

---

## Package Details

### Debian Package
- **File**: `systask_1.0.2_amd64.deb`
- **Installs to**: `/usr/bin/systask`
- **Config**: `/etc/systask/config.yaml.example`
- **Depends**: openssh-client

### RPM Package
- **File**: `systask-1.0.2-1.x86_64.rpm`
- **Installs to**: `/usr/bin/systask`
- **Config**: `/etc/systask/config.yaml.example`
- **Depends**: openssh-clients

### Arch Package (AUR)
- **Package**: `systask`
- **Install**: `sudo pacman -S systask`
- **Installs to**: `/usr/bin/systask`
- **Config**: `/etc/systask/config.yaml.example`

---

## Configuration

After installation, create your config file:

```bash
mkdir -p ~/.config/systask
cp /etc/systask/config.yaml.example ~/.config/systask/config.yaml
```

Edit `~/.config/systask/config.yaml` to add your SSH hosts.

---

## Uninstall

### Debian/Ubuntu
```bash
sudo apt remove systask
```

### Fedora/RHEL/CentOS
```bash
sudo dnf remove systask
# or
sudo yum remove systask
```

### Arch Linux
```bash
sudo pacman -R systask
```

### Manual
```bash
sudo rm /usr/local/bin/systask
```

---

## Troubleshooting

### Command not found
Ensure the installation directory is in your PATH:
```bash
which systask
# Should output: /usr/bin/systask or /usr/local/bin/systask
```

### Permission denied
Make sure the binary has execute permissions:
```bash
ls -l /usr/bin/systask
# Should show: -rwxr-xr-x
```

### SSH connection issues
Verify SSH is installed and configured:
```bash
ssh -V
cat ~/.ssh/config
```

---

## Support

For issues and feature requests, visit: https://github.com/onedord1/stk/issues

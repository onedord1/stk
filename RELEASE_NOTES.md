# SysTask v1.0.0 Release Notes

**Release Date**: January 20, 2024

## 🎉 Initial Release

SysTask is a powerful Terminal User Interface (TUI) for managing multiple Linux servers via SSH. This initial release includes a comprehensive feature set for server administration.

---

## ✨ Features

### 🖥️ 10 Management Modules

| Module | Description |
|--------|-------------|
| **Health** | Real-time CPU, Memory, Disk, Network metrics |
| **Services** | Systemd service management |
| **Processes** | Process viewer with kill functionality |
| **Logs** | Real-time log viewer with filtering |
| **Disks** | Disk usage and mount information |
| **Batch** | Multi-server command execution |
| **Users** | User account management |
| **Docker** | Container management |
| **Installer** | Package installation |
| **SFTP** | Dual-pane file manager |

### 🔌 Terminal Mode
Fullscreen SSH terminal access - press `t` after connecting for an immersive shell experience.

### 🎨 12 Color Themes
- catppuccin (default)
- dracula
- nord
- gruvbox
- solarized
- tokyo-night
- monokai
- one-dark
- cyberpunk
- forest
- ocean
- sunset

### 🔐 Security Features
- **AES-256-GCM encryption** for stored passwords
- **PBKDF2 key derivation** with unique salt
- SSH key and agent authentication support
- No plaintext credential storage

### 📁 Host Management
- Add, edit, delete hosts via form UI
- Auto-discovery from `~/.ssh/config`
- Host grouping (Production, Staging, etc.)
- AWS/Azure/GCP PEM key support

---

## 📥 Installation

### Binary Download
```bash
# Linux AMD64
curl -LO https://github.com/yourusername/systask/releases/download/v1.0.0/systask-linux-amd64
chmod +x systask-linux-amd64
./systask-linux-amd64
```

### Build from Source
```bash
git clone https://github.com/yourusername/systask.git
cd systask
go build -o systask ./main.go
./systask
```

---

## ⌨️ Quick Start

1. **Launch**: `./systask`
2. **Add server**: Press `a`
3. **Connect**: Select host, press `Enter`
4. **Switch modules**: Press `1-9` or `F`
5. **Terminal mode**: Press `t`
6. **Change theme**: Type `:theme dracula`
7. **Get help**: Press `?`

---

## 📋 Requirements

- Linux (tested on Ubuntu 20.04+, Debian 10+, RHEL 8+)
- Terminal with Unicode support
- SSH access to remote servers

---

## 🔗 Links

- **Documentation**: See README.md
- **Configuration**: See config.yaml.example
- **Issues**: [GitHub Issues](https://github.com/yourusername/systask/issues)

---

## 🙏 Acknowledgments

Built with:
- [tview](https://github.com/rivo/tview) - Terminal UI library
- [tcell](https://github.com/gdamore/tcell) - Terminal cell library
- [golang.org/x/crypto](https://pkg.go.dev/golang.org/x/crypto) - SSH and encryption

---

## 📄 License

MIT License - see LICENSE file for details.

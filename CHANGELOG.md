# Changelog

All notable changes to SysTask will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2024-01-20

### Added
- **10 Modules**: Health, Services, Processes, Logs, Disks, Batch, Users, Docker, Installer, SFTP
- **Terminal Mode**: Fullscreen SSH terminal with `t` key
- **12 Themes**: catppuccin, dracula, nord, gruvbox, solarized, tokyo-night, monokai, one-dark, cyberpunk, forest, ocean, sunset
- **Host Management**: Add, edit, delete hosts with form-based UI
- **SSH Config Integration**: Auto-discovers hosts from `~/.ssh/config`
- **PEM Key Support**: AWS, Azure, GCP key-based authentication
- **Batch Execution**: Run commands across multiple servers simultaneously
- **SFTP File Manager**: Dual-pane file browser with copy/move operations
- **Encrypted Credentials**: AES-256-GCM encryption for stored passwords
- **Host Groups**: Organize servers by environment
- **Real-time Metrics**: CPU, Memory, Disk, Network monitoring
- **Service Management**: Start/stop/restart systemd services
- **Process Viewer**: View and kill processes
- **Docker Management**: Container start/stop/logs
- **Command Palette**: Vim-style `:` commands
- **Search**: Filter servers with `/`
- **Help Overlay**: Press `?` for context-sensitive help

### Security
- PBKDF2 key derivation for password encryption
- Unique salt per installation
- No plaintext password storage

## [Unreleased]

### Planned
- SSH tunneling / port forwarding
- Ansible playbook integration
- Cloud provider discovery (AWS, GCP, Azure)
- Session recording and playback
- Team sharing with encrypted configs

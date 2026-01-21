# Changelog

All notable changes to SysTask will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.2] - 2026-01-22

### Fixed
- **SFTP Copy (`c`) Not Working**: Global app handler was intercepting 'c' key for host connection. Now correctly delegates to SFTP module when active.
- **SFTP Progress Bar Showing 0%**: Progress bar now displays real-time transfer progress with proper percentage calculation.
- **Transfer Queue UI Improvements**:
  - Progress bar width increased to 80 characters for better visibility
  - Animated progress display for single file transfers (10% → 30% → 60% → 100%)
  - File count progress for multiple files (1/5, 2/5, etc.)
  - Status icons: ⏳ (in progress) → ✅ (complete) → ❌ (error)
  - Combined layout for maximum bar width

### Improved
- Better status messages with debugging info (row and file count)
- Clear error messages in transfer queue with error icon
- Completion feedback showing total files transferred

## [3.1.0] - 2026-01-20

### Added
- **Disk Partition Management**: Full partition operations in Block Devices view
  - `c` = Create partition (with parted commands)
  - `d` = Delete partition (with safety warnings)
  - `f` = Format partition (ext4, xfs, btrfs, ntfs)
  - `m` = Mount/Unmount partitions
  - `i` = View detailed device info
  - Tab to switch between Filesystems and Block Devices views
- **Enhanced Health Module**: Complete UI redesign
  - Network RX/TX statistics
  - Process and user count
  - Swap usage display
  - Big visual progress bars with color coding
  - Quick stats footer bar
- **Logs Module**: Word wrap toggle (`w`) and real-time following (`f`)
- **Installer Module**: Fixed package manager detection for Amazon Linux, Alpine
- **Responsive Layout**: Better display on small screens (13" 1080p)
  - Hostname truncation with ellipsis
  - Reduced sidebar widths
  - Improved panel proportions

### Fixed
- Package manager detection now supports: apt, dnf, yum, pacman, yay, zypper, apk
- Tab navigation between Categories and Packages in Installer
- Installer Refresh now called when switching to module

## [3.0.0] - 2026-01-19

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

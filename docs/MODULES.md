# SysTask Modules Guide

This guide provides detailed information about each SysTask module.

---

## Health Module (Key: 1)

Real-time system health monitoring.

### Metrics Displayed
- **CPU**: Usage percentage, load average, per-core stats
- **Memory**: Used/Total, percentage, swap usage
- **Disk**: Root partition usage, I/O stats
- **Network**: Bytes sent/received, active connections

### Keyboard Shortcuts
| Key | Action |
|-----|--------|
| `r` | Refresh metrics |
| `ESC` | Return to home |

---

## Services Module (Key: 2)

Manage systemd services.

### Features
- View all services with status (running/stopped/failed)
- Start, stop, restart services
- Enable/disable on boot
- View service logs

### Keyboard Shortcuts
| Key | Action |
|-----|--------|
| `Enter` | Select service for actions |
| `s` | Start service |
| `x` | Stop service |
| `r` | Restart service |
| `e` | Enable on boot |
| `d` | Disable on boot |
| `l` | View logs |
| `/` | Filter services |

---

## Processes Module (Key: 3)

Process management like `htop`.

### Features
- View running processes with CPU/Memory usage
- Sort by various columns
- Kill processes
- Search/filter

### Keyboard Shortcuts
| Key | Action |
|-----|--------|
| `k` | Kill selected process |
| `K` | Kill with SIGKILL |
| `c` | Sort by CPU |
| `m` | Sort by Memory |
| `p` | Sort by PID |
| `/` | Filter processes |

---

## Logs Module (Key: 4)

Real-time log viewer using journalctl.

### Features
- View system logs in real-time
- Filter by unit/service
- Search within logs
- Tail mode (follow)

### Keyboard Shortcuts
| Key | Action |
|-----|--------|
| `f` | Toggle follow mode |
| `/` | Search logs |
| `u` | Filter by unit |
| `n` | Next match |
| `N` | Previous match |

---

## Disks Module (Key: 5)

Disk and storage information.

### Features
- View all mount points
- Disk usage with visual bars
- Filesystem type
- Mount options

### Displayed Information
- Device name
- Mount point
- Total/Used/Available space
- Usage percentage
- Filesystem type

---

## Batch Module (Key: 6)

Execute commands on multiple servers.

### Workflow
1. Select hosts with `Space`
2. Press `Tab` to enter command
3. Type your command
4. Press `Enter` to execute
5. View results per host

### Keyboard Shortcuts
| Key | Action |
|-----|--------|
| `Space` | Toggle host selection |
| `a` | Select all hosts |
| `n` | Deselect all |
| `Tab` | Switch to command input |
| `Enter` | Execute command |

---

## Users Module (Key: 7)

User account management.

### Features
- List all users
- View user details (groups, home, shell)
- Add users (requires root)
- Delete users
- Modify user properties

### Displayed Information
- Username
- UID
- Primary group
- Home directory
- Shell
- Last login

---

## Docker Module (Key: 8)

Docker container management.

### Features
- List all containers (running and stopped)
- Start/stop/restart containers
- View container logs
- View container stats
- Execute commands in container

### Keyboard Shortcuts
| Key | Action |
|-----|--------|
| `Enter` | Select container |
| `s` | Start container |
| `x` | Stop container |
| `r` | Restart container |
| `l` | View logs |
| `e` | Exec into container |

---

## Installer Module (Key: 9)

Package installation across servers.

### Features
- Search packages
- Install packages
- Remove packages
- View installed packages

### Supported Package Managers
- apt (Debian/Ubuntu)
- yum/dnf (RHEL/CentOS/Fedora)
- pacman (Arch)

---

## SFTP Module (Key: F)

Dual-pane file manager with transfer capabilities.

### Layout
- Left pane: Source (can be local or remote)
- Right pane: Destination (can be local or remote)

### Features
- Navigate directories
- Copy files between hosts
- Move/rename files
- Create directories
- Delete files
- View file properties

### Keyboard Shortcuts
| Key | Action |
|-----|--------|
| `Tab` | Switch panels |
| `Space` | Select file |
| `c` | Copy selected files |
| `m` | Move selected files |
| `d` | Delete selected |
| `h` | Change host |
| `n` | New directory |
| `Enter` | Open directory / view file |

---

## Terminal Mode (Key: t)

Fullscreen interactive SSH terminal.

### Features
- Direct shell access
- Full terminal emulation
- No UI overhead

### Usage
1. Connect to a server
2. Press `t`
3. Use normally
4. Press `Ctrl+D` to exit

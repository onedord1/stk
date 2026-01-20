# SysTask v3.1.0 Release Notes

**Release Date**: January 20, 2026

## 🚀 Major Update: Disk Partitioning & Visual Overhaul

This release introduces powerful disk management capabilities and a completely redesigned Health module for better visibility into your system's status.

---

## ✨ New Features

### 🔧 Partition Management
The Disk module now includes comprehensive partition operations. Switch to the **Block Devices** view (Tab) to manage your disks.

- **Create Partitions** (`c`): Generate `parted` commands to create new partitions.
- **Delete Partitions** (`d`): Safely remove partitions with confirmation warnings.
- **Format Partitions** (`f`): Generate commands to format partitions (ext4, xfs, btrfs, ntfs, vfat).
- **Mount/Unmount** (`m`): Quickly toggle mount status of partitions.
- **Device Info** (`i`): View detailed `fdisk` information for any device.

### 🏥 Redesigned Health Module
A complete visual overhaul of the Health dashboard for instant status awareness.

- **Enhanced Visuals**: Big, colorful progress bars with gradient effects.
- **New Metrics**:
  - **Network Traffic**: Real-time RX (Receive) and TX (Transmit) stats.
  - **Counts**: Active Process count and Logged-in User count.
  - **Swap Usage**: Dedicated visualization for swap memory.
- **Smart Coloring**: Indicators change color (Green/Yellow/Red) based on load.
- **Quick Stats**: A handy footer bar summarizing key metrics.

### 📱 Responsive Layout
Improved scaling for smaller screens (e.g., 13" laptops at 1080p).
- **Smart Truncation**: Long hostnames are elegantly truncated (`host-name...`).
- **Compact Sidebars**: Optimized panel widths for better screen utilization.

### 🛠️ Improvements
- **Logs Module**: 
  - Toggle word wrap with `w`.
  - Follow real-time logs with `f`.
- **Installer Module**: 
  - Fixed package manager detection for Amazon Linux (yum/dnf) and Alpine (apk).
  - Calling Refresh manually is no longer needed when switching tabs.
- **General**: Added keyboard shortcut hints to various modules.

---

## 📥 Installation

```bash
# Update existing installation
git pull
go build -o systask ./main.go
./systask
```

---

## ⚠️ Important Notes
- Partition operations are **destructive**. Always verify commands before running them.
- Sudo privileges are required for partition management.

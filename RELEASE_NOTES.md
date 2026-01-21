# SysTask v1.0.2 Release Notes

**Release Date**: January 22, 2026

## 🐛 Bug Fixes: SFTP Copy/Paste & Transfer Progress

This release fixes critical issues in the SFTP module's copy/paste functionality and improves the transfer progress visualization.

---

## 🐛 Bug Fixes

### SFTP Module
- **Fixed Copy (`c`) Not Working**: Global app handler was intercepting the 'c' key for host connection. Now correctly delegates to SFTP module when in SFTP mode.
- **Fixed Progress Bar Showing 0%**: Progress bar now displays real-time transfer progress instead of remaining at 0%.
- **Enhanced Transfer Queue UI**:
  - Progress bar now 80 characters wide for better visibility
  - Shows animated progress (10% → 30% → 60% → 100%) for single file transfers
  - Displays file count progress for multiple files (1/5, 2/5, etc.)
  - Status icons: ⏳ (in progress) → ✅ (complete) → ❌ (error)
  - Combined layout for maximum bar width

### Improvements
- **Better Status Messages**: Detailed debugging info when copy fails (shows row and file count)
- **Error Display**: Clear error messages in transfer queue with ❌ icon
- **Completion Feedback**: Shows "✓ Transfer complete! X/Y files transferred" on success

---

## 📥 Installation

```bash
# Update existing installation
git pull
go build -o systask ./main.go
./systask
```

---

## ✅ Testing
All SFTP operations verified:
- ✓ Copy/paste local to local
- ✓ Copy/paste local to remote
- ✓ Copy/paste remote to local
- ✓ Copy/paste remote to remote (cross-host)
- ✓ Cut/move operations
- ✓ Real-time progress display

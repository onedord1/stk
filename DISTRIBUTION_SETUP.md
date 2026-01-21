# SysTask Distribution Setup Guide

This guide explains how to set up global distribution for SysTask across Ubuntu/Debian (PPA), Fedora/RHEL (Copr), and Arch Linux (AUR).

## Overview

| Distribution | Method | Installation | Status |
|---|---|---|---|
| **Ubuntu/Debian** | Launchpad PPA | `sudo apt install systask` | 📋 Setup needed |
| **Fedora/RHEL/CentOS** | Copr | `sudo dnf install systask` | 📋 Setup needed |
| **Arch Linux** | AUR | `yay -S systask` | 📋 Setup needed |

## Quick Start

### Automated Setup
```bash
./scripts/setup-distributions.sh
```

### Manual Setup

## 1. Launchpad PPA (Ubuntu/Debian)

### Prerequisites
- Launchpad account: https://launchpad.net/
- GPG key for signing packages
- `dput` and `devscripts` installed

### Setup Steps

1. **Create Launchpad Account**
   ```bash
   # Go to https://launchpad.net/ and sign up
   ```

2. **Generate/Upload GPG Key**
   ```bash
   # Generate key (if needed)
   gpg --gen-key
   
   # Get your key ID
   gpg --list-keys
   
   # Upload to keyserver
   gpg --send-keys --keyserver keyserver.ubuntu.com YOUR_KEY_ID
   ```

3. **Create PPA**
   - Go to https://launchpad.net/~yourusername/+create-ppa
   - Name: `systask`
   - Description: "Terminal system manager for Linux servers"
   - Click "Create PPA"

4. **Install Build Tools**
   ```bash
   sudo apt-get install dput devscripts ubuntu-dev-tools
   ```

5. **Build and Upload**
   ```bash
   cd ~/CascadeProjects/stk
   
   # Build source package
   debuild -S -sa
   
   # Upload to PPA
   dput ppa:yourusername/systask ../systask_1.0.3-1_source.changes
   ```

6. **Verify**
   - Go to https://launchpad.net/~yourusername/+archive/ubuntu/systask
   - Wait for builds to complete (5-30 minutes)

### User Installation
```bash
sudo add-apt-repository ppa:yourusername/systask
sudo apt update
sudo apt install systask
```

**Detailed Guide**: See `packaging/debian/PPA_SETUP.md`

---

## 2. Fedora Copr (Fedora/RHEL/CentOS)

### Prerequisites
- Copr account: https://copr.fedorainfracloud.org/
- GitHub repository (already have: onedord1/stk)
- copr-cli installed

### Setup Steps

1. **Create Copr Account**
   - Go to https://copr.fedorainfracloud.org/
   - Sign in with Fedora Account or GitHub
   - Authorize Copr

2. **Create Copr Project**
   - Go to https://copr.fedorainfracloud.org/coprs/create/
   - Project name: `systask`
   - Description: "Terminal system manager for Linux servers"
   - Select chroots:
     - ✓ Fedora 39 (x86_64, aarch64)
     - ✓ Fedora 40 (x86_64, aarch64)
     - ✓ EPEL 8 (x86_64, aarch64)
     - ✓ EPEL 9 (x86_64, aarch64)
   - Click "Create"

3. **Enable GitHub Webhook**
   - Go to project settings
   - Enable "Build from GitHub"
   - Enable "GitHub webhook"
   - Repository: `onedord1/stk`
   - Spec file: `packaging/rpm/systask.spec`

4. **Trigger First Build**
   ```bash
   # Option 1: Via web interface
   # Go to project → New Build → Build from GitHub → stable branch
   
   # Option 2: Via CLI
   sudo dnf install copr-cli
   copr-cli config
   copr-cli build-package systask \
     --git-url https://github.com/onedord1/stk \
     --git-branch stable \
     --spec packaging/rpm/systask.spec
   ```

5. **Verify**
   - Go to https://copr.fedorainfracloud.org/coprs/yourusername/systask/
   - Check "Builds" tab
   - Wait for all builds to complete

### User Installation
```bash
sudo dnf copr enable yourusername/systask
sudo dnf install systask
```

**Detailed Guide**: See `packaging/rpm/COPR_SETUP.md`

---

## 3. Arch Linux AUR

### Prerequisites
- AUR account: https://aur.archlinux.org/
- SSH key for AUR access

### Setup Steps

1. **Create AUR Account**
   - Go to https://aur.archlinux.org/register/
   - Create account

2. **Generate SSH Key**
   ```bash
   ssh-keygen -t ed25519 -f ~/.ssh/aur -C "aur@systask"
   ```

3. **Add SSH Key to AUR Account**
   - Log in to https://aur.archlinux.org
   - Account Settings
   - Paste public key from `~/.ssh/aur.pub`

4. **Configure SSH**
   ```bash
   cat >> ~/.ssh/config << 'EOF'
   Host aur.archlinux.org
       IdentityFile ~/.ssh/aur
       User aur
   EOF
   chmod 600 ~/.ssh/config
   ```

5. **Submit Package**
   ```bash
   # Clone AUR repository
   git clone ssh://aur@aur.archlinux.org/systask.git
   cd systask
   
   # Copy PKGBUILD
   cp ../packaging/arch/PKGBUILD .
   
   # Generate .SRCINFO
   makepkg --printsrcinfo > .SRCINFO
   
   # Commit and push
   git add PKGBUILD .SRCINFO
   git commit -m "Initial commit for systask v1.0.3"
   git push
   ```

6. **Verify**
   - Go to https://aur.archlinux.org/packages/systask
   - Package should appear within 24 hours

### User Installation
```bash
# Using yay
yay -S systask

# Using paru
paru -S systask

# Manual
git clone https://aur.archlinux.org/systask.git
cd systask
makepkg -si
```

**Detailed Guide**: See `packaging/arch/AUR_SUBMISSION.md`

---

## Installation Commands Summary

Once all distributions are set up, users can install with:

### Ubuntu/Debian
```bash
sudo add-apt-repository ppa:yourusername/systask
sudo apt update
sudo apt install systask
```

### Fedora/RHEL/CentOS
```bash
sudo dnf copr enable yourusername/systask
sudo dnf install systask
```

### Arch Linux
```bash
yay -S systask
# or
paru -S systask
```

### From Source (Any Linux)
```bash
git clone https://github.com/onedord1/stk.git
cd stk
./scripts/install.sh
```

---

## Maintenance

### Update to New Version

1. **Update version in files**
   - `CHANGELOG.md`
   - `RELEASE_NOTES.md`
   - `packaging/debian/control`
   - `packaging/rpm/systask.spec`
   - `packaging/arch/PKGBUILD`

2. **Commit and push**
   ```bash
   git add -A
   git commit -m "Update to v1.0.4"
   git push origin stable
   ```

3. **Create GitHub Release**
   ```bash
   git tag -a v1.0.4 -m "Release v1.0.4"
   git push origin v1.0.4
   gh release create v1.0.4 --notes-file RELEASE_NOTES.md
   ```

4. **Distributions Update Automatically**
   - PPA: Manual rebuild needed
   - Copr: Automatic via GitHub webhook
   - AUR: Manual update needed

---

## Troubleshooting

### PPA Build Fails
- Check build logs at https://launchpad.net/~yourusername/+archive/ubuntu/systask
- Common issues: missing dependencies, Go version, incorrect paths

### Copr Build Fails
- Check build logs at https://copr.fedorainfracloud.org/coprs/yourusername/systask/
- Verify `.spec` file path is correct
- Check Go version requirement

### AUR Submission Fails
- Verify SSH key is added to AUR account
- Check PKGBUILD syntax: `bash -n PKGBUILD`
- Ensure `.SRCINFO` is generated correctly

---

## References

- [Launchpad PPA Guide](https://help.launchpad.net/Packaging/PPA)
- [Copr Documentation](https://docs.pagure.org/copr.copr/)
- [AUR Submission Guidelines](https://wiki.archlinux.org/title/AUR_submission_guidelines)
- [Debian Packaging Guide](https://www.debian.org/doc/manuals/debian-faq/pkg-basics.en.html)
- [RPM Packaging Guide](https://rpm-packaging-guide.github.io/)

# Packaging Guide

This document explains how to build and distribute SysTask packages for different Linux distributions.

## Overview

SysTask is packaged for:
- **Debian/Ubuntu** - `.deb` packages
- **Fedora/RHEL/CentOS** - `.rpm` packages
- **Arch Linux** - AUR (Arch User Repository)
- **Alpine Linux** - APK packages (planned)

## Building Packages Locally

### Prerequisites
```bash
# Debian/Ubuntu
sudo apt-get install build-essential debhelper devscripts rpm golang-go

# Fedora/RHEL
sudo dnf install rpm-build golang

# Arch Linux
sudo pacman -S base-devel go
```

### Build All Packages
```bash
./scripts/build-packages.sh
```

Output will be in `dist/` directory:
- `systask_1.0.2_amd64.deb` - Debian package
- `systask-1.0.2-1.x86_64.rpm` - RPM package
- `PKGBUILD` - Arch Linux build file

## Debian Package Distribution

### Local Installation
```bash
sudo dpkg -i dist/systask_1.0.2_amd64.deb
```

### PPA (Personal Package Archive)
To distribute via PPA (Ubuntu):

1. Create a Launchpad account
2. Create a PPA at https://launchpad.net/~yourname/+create-ppa
3. Install `dput`:
   ```bash
   sudo apt-get install dput
   ```

4. Configure `~/.dput.cf`:
   ```ini
   [ppa]
   fqdn = ppa.launchpad.net
   method = sftp
   incoming = ~yourname/ppa/ubuntu
   login = yourname
   allow_unsigned_uploads = 0
   ```

5. Build source package:
   ```bash
   debuild -S -sa
   ```

6. Upload to PPA:
   ```bash
   dput ppa:yourname/ppa ../systask_1.0.2-1_source.changes
   ```

### Debian Repository
To create your own Debian repository:

```bash
# Create repository structure
mkdir -p debian-repo/pool/main/s/systask
mkdir -p debian-repo/dists/focal/main/binary-amd64

# Copy package
cp dist/systask_1.0.2_amd64.deb debian-repo/pool/main/s/systask/

# Create Packages file
cd debian-repo/dists/focal/main/binary-amd64
apt-ftparchive packages ../../../pool/main > Packages
gzip -k Packages

# Create Release file
cd debian-repo/dists/focal
apt-ftparchive release . > Release

# Sign Release (optional)
gpg --clearsign -o InRelease Release
```

## RPM Package Distribution

### Local Installation
```bash
sudo dnf install dist/systask-1.0.2-1.x86_64.rpm
# or
sudo yum install dist/systask-1.0.2-1.x86_64.rpm
```

### Copr (Community Projects)
To distribute via Copr (Fedora):

1. Create account at https://copr.fedorainfracloud.org
2. Create a new project
3. Upload SRPM or connect GitHub repository
4. Enable builds for desired distributions

### Custom RPM Repository
```bash
# Create repository
mkdir -p rpm-repo/RPMS/x86_64
cp dist/systask-1.0.2-1.x86_64.rpm rpm-repo/RPMS/x86_64/

# Create repository metadata
createrepo rpm-repo/

# Sign packages (optional)
gpg --detach-sign --armor rpm-repo/RPMS/x86_64/systask-1.0.2-1.x86_64.rpm
```

## Arch Linux (AUR)

### Submit to AUR
1. Create account at https://aur.archlinux.org
2. Generate SSH key for AUR
3. Clone AUR repository:
   ```bash
   git clone ssh://aur@aur.archlinux.org/systask.git
   ```

4. Copy PKGBUILD:
   ```bash
   cp dist/PKGBUILD systask/
   ```

5. Update `.SRCINFO`:
   ```bash
   cd systask
   makepkg --printsrcinfo > .SRCINFO
   ```

6. Commit and push:
   ```bash
   git add PKGBUILD .SRCINFO
   git commit -m "Update to v1.0.2"
   git push
   ```

### Local AUR Installation
```bash
git clone https://aur.archlinux.org/systask.git
cd systask
makepkg -si
```

## GitHub Actions Automation

The `.github/workflows/release.yml` workflow automatically:

1. **Builds binaries** for Linux AMD64 and ARM64
2. **Creates Debian package** (.deb)
3. **Creates RPM package** (.rpm)
4. **Publishes to AUR** (requires SSH key setup)
5. **Uploads to GitHub Release**

### Setup GitHub Actions

#### For AUR Publishing
1. Generate SSH key:
   ```bash
   ssh-keygen -t ed25519 -f aur-key -C "systask-bot"
   ```

2. Add public key to AUR account settings

3. Add private key to GitHub Secrets:
   - Go to Settings → Secrets and variables → Actions
   - Create `AUR_SSH_PRIVATE_KEY` with the private key content

## Release Process

### Manual Release
```bash
# Build packages
./scripts/build-packages.sh

# Create git tag
git tag -a v1.0.2 -m "Release v1.0.2"
git push origin v1.0.2

# Create GitHub release
gh release create v1.0.2 \
  --title "v1.0.2 - SFTP Copy/Paste & Progress Bar Fixes" \
  --notes-file RELEASE_NOTES.md \
  dist/systask_1.0.2_amd64.deb \
  dist/systask-1.0.2-1.x86_64.rpm
```

### Automated Release
1. Push tag to GitHub
2. GitHub Actions automatically:
   - Builds all packages
   - Uploads to release
   - Publishes to AUR
   - Updates package repositories

## Distribution Checklist

- [ ] Update version in `CHANGELOG.md`
- [ ] Update version in `RELEASE_NOTES.md`
- [ ] Build packages: `./scripts/build-packages.sh`
- [ ] Test packages on target distributions
- [ ] Create git tag: `git tag -a vX.Y.Z`
- [ ] Push tag: `git push origin vX.Y.Z`
- [ ] Create GitHub release with packages
- [ ] Update AUR PKGBUILD
- [ ] Update Debian/RPM repositories
- [ ] Announce release on social media

## Testing Packages

### Docker Testing
```bash
# Test Debian
docker run -it ubuntu:latest bash
apt-get update && apt-get install -y ./systask_1.0.2_amd64.deb
systask --help

# Test Fedora
docker run -it fedora:latest bash
dnf install -y ./systask-1.0.2-1.x86_64.rpm
systask --help

# Test Arch
docker run -it archlinux:latest bash
pacman -U ./systask-1.0.2-1-x86_64.pkg.tar.zst
systask --help
```

## Troubleshooting

### Debian Package Issues
```bash
# Check package contents
dpkg -c dist/systask_1.0.2_amd64.deb

# Verify dependencies
dpkg -I dist/systask_1.0.2_amd64.deb

# Install with verbose output
sudo apt-get install -y ./dist/systask_1.0.2_amd64.deb -o Debug::pkgProblemResolver=yes
```

### RPM Package Issues
```bash
# Check package contents
rpm -qlp dist/systask-1.0.2-1.x86_64.rpm

# Verify dependencies
rpm -qp --requires dist/systask-1.0.2-1.x86_64.rpm

# Install with verbose output
sudo dnf install -v ./dist/systask-1.0.2-1.x86_64.rpm
```

## References

- [Debian Packaging Guide](https://www.debian.org/doc/manuals/debian-faq/pkg-basics.en.html)
- [RPM Packaging Guide](https://rpm-packaging-guide.github.io/)
- [Arch Linux Packaging Standards](https://wiki.archlinux.org/title/Arch_packaging_standards)
- [AUR Submission Guidelines](https://wiki.archlinux.org/title/AUR_submission_guidelines)

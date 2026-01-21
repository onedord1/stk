# Copr Setup Guide for SysTask

## What is Copr?

Copr (Community Projects) is Fedora's community build system that automatically builds RPM packages from source and hosts them in a repository. Users can install with a single command.

## Prerequisites

1. **Copr Account**
   - Go to https://copr.fedorainfracloud.org/
   - Sign in with Fedora Account or GitHub
   - Authorize Copr to access your account

2. **GitHub Repository**
   - Your project should be on GitHub (already done: https://github.com/onedord1/stk)
   - Repository should have a `.spec` file in the root or `packaging/rpm/` directory

## Create Copr Project

### Option 1: Web Interface (Easiest)

1. Go to https://copr.fedorainfracloud.org/coprs/create/
2. Fill in the form:
   - **Project name**: `systask`
   - **Description**: "Terminal system manager for Linux servers"
   - **Instructions**: "Install with: sudo dnf copr enable onedord1/systask && sudo dnf install systask"
   - **Chroots**: Select target distributions:
     - ✓ Fedora 39 (x86_64, aarch64)
     - ✓ Fedora 40 (x86_64, aarch64)
     - ✓ EPEL 8 (x86_64, aarch64)
     - ✓ EPEL 9 (x86_64, aarch64)
3. Click "Create"

### Option 2: Command Line

```bash
# Install copr-cli
sudo dnf install copr-cli

# Configure
copr-cli config

# Create project
copr-cli create systask \
  --chroot fedora-39-x86_64 \
  --chroot fedora-39-aarch64 \
  --chroot fedora-40-x86_64 \
  --chroot fedora-40-aarch64 \
  --chroot epel-8-x86_64 \
  --chroot epel-8-aarch64 \
  --chroot epel-9-x86_64 \
  --chroot epel-9-aarch64
```

## Connect GitHub Repository

1. Go to your Copr project: https://copr.fedorainfracloud.org/coprs/yourusername/systask/
2. Click "Settings"
3. Under "Build Options", enable:
   - **Build from GitHub**: ON
   - **GitHub webhook**: ON
4. Enter GitHub repository: `onedord1/stk`
5. Specify `.spec` file location: `packaging/rpm/systask.spec`
6. Click "Save"

## Trigger First Build

### Option 1: Web Interface
1. Go to project page
2. Click "New Build"
3. Select "Build from GitHub"
4. Enter branch: `stable` or `main`
5. Click "Build"

### Option 2: Command Line
```bash
copr-cli build-package systask \
  --git-url https://github.com/onedord1/stk \
  --git-branch stable \
  --spec packaging/rpm/systask.spec
```

### Option 3: Automatic on GitHub Release
Once webhook is enabled, Copr automatically builds when you:
1. Push a tag: `git tag -a v1.0.3 && git push origin v1.0.3`
2. Create a GitHub Release
3. Copr detects the release and builds automatically

## Users Installation

Once Copr project is set up and builds complete:

```bash
# Add Copr repository
sudo dnf copr enable onedord1/systask

# Install
sudo dnf install systask

# Or in one command
sudo dnf install -y 'dnf-command(copr)' && sudo dnf copr enable onedord1/systask && sudo dnf install systask
```

## Verify Build Status

1. Go to https://copr.fedorainfracloud.org/coprs/onedord1/systask/
2. Check "Builds" tab
3. View build logs for each architecture
4. Once all builds show "succeeded", package is ready

## Troubleshooting

### Build Fails
- Click on failed build
- Check "Build Log" tab
- Common issues:
  - Missing dependencies in `.spec` file
  - Go version too old
  - Incorrect `.spec` file path

### GitHub Webhook Not Working
```bash
# Check webhook configuration
# Go to GitHub repo Settings → Webhooks
# Verify Copr webhook is listed and active

# Manual rebuild
copr-cli build-package systask \
  --git-url https://github.com/onedord1/stk \
  --git-branch stable
```

### Package Not Installing
```bash
# Verify Copr is enabled
dnf copr list

# Check available versions
dnf search systask

# View available packages
dnf info systask --all
```

## Maintenance

### Update to New Version

1. Update `.spec` file with new version
2. Commit and push to GitHub
3. Create GitHub Release with tag
4. Copr automatically builds new version

### Delete Old Builds

1. Go to project page
2. Click on old build
3. Click "Delete"

### Disable Copr Project

1. Go to project settings
2. Click "Delete this project"

## Supported Distributions

| Distribution | Versions | Architectures |
|---|---|---|
| Fedora | 39, 40 | x86_64, aarch64 |
| EPEL | 8, 9 | x86_64, aarch64 |
| CentOS Stream | 8, 9 | x86_64, aarch64 |

## References

- [Copr Documentation](https://docs.pagure.org/copr.copr/)
- [RPM Packaging Guide](https://rpm-packaging-guide.github.io/)
- [Fedora Packaging Guide](https://docs.fedoraproject.org/en-US/packaging-guidelines/)

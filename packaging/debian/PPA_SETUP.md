# Launchpad PPA Setup Guide for SysTask

## Prerequisites

1. **Launchpad Account**
   - Go to https://launchpad.net/
   - Sign up or log in with your account

2. **GPG Key Setup**
   ```bash
   # Generate GPG key (if you don't have one)
   gpg --gen-key
   
   # Export public key
   gpg --export -a your-email@example.com > pubkey.asc
   
   # Upload to Ubuntu keyserver
   gpg --send-keys --keyserver keyserver.ubuntu.com YOUR_KEY_ID
   ```

3. **Install dput and devscripts**
   ```bash
   sudo apt-get install dput devscripts ubuntu-dev-tools
   ```

## Create PPA on Launchpad

1. Go to https://launchpad.net/~yourusername/+create-ppa
2. Fill in:
   - **PPA name**: `systask` (or `systask-releases`)
   - **Description**: "Official SysTask releases - Terminal system manager for Linux servers"
   - **Display name**: Keep default
3. Click "Create PPA"
4. Note your PPA URL: `ppa:yourusername/systask`

## Build Source Package

```bash
cd ~/CascadeProjects/stk

# Create debian directory structure
mkdir -p debian

# Copy control file
cp packaging/debian/control debian/

# Create changelog
cat > debian/changelog << 'EOF'
systask (1.0.2-1) focal; urgency=medium

  * Fixed SFTP copy/paste functionality
  * Fixed progress bar display
  * Enhanced transfer queue UI
  * Added system package support

 -- Your Name <your-email@example.com>  Wed, 22 Jan 2026 05:00:00 +0600

systask (1.0.0-1) focal; urgency=medium

  * Initial release

 -- Your Name <your-email@example.com>  Mon, 20 Jan 2026 00:00:00 +0600
EOF

# Create rules file
cat > debian/rules << 'EOF'
#!/usr/bin/make -f

%:
	dh $@

override_dh_auto_build:
	go build -o systask ./main.go

override_dh_auto_install:
	dh_auto_install
	install -Dm755 systask debian/systask/usr/bin/systask
	install -Dm644 config.yaml.example debian/systask/etc/systask/config.yaml.example
EOF
chmod +x debian/rules

# Create compat file
echo "13" > debian/compat

# Create source/format file
mkdir -p debian/source
echo "3.0 (quilt)" > debian/source/format

# Build source package
debuild -S -sa
```

## Upload to PPA

```bash
# Configure dput
cat > ~/.dput.cf << 'EOF'
[ppa]
fqdn = ppa.launchpad.net
method = sftp
incoming = ~yourusername/ppa/ubuntu
login = yourusername
allow_unsigned_uploads = 0
EOF

# Upload to PPA
dput ppa:yourusername/systask ../systask_1.0.2-1_source.changes
```

## Verify Upload

1. Go to https://launchpad.net/~yourusername/+archive/ubuntu/systask
2. Wait for build to complete (usually 5-30 minutes)
3. Check build status for each Ubuntu release

## Users Installation

Once built and published:

```bash
# Add PPA
sudo add-apt-repository ppa:yourusername/systask

# Update package list
sudo apt update

# Install
sudo apt install systask
```

## Supported Ubuntu Versions

By default, PPA builds for:
- Ubuntu 20.04 LTS (focal)
- Ubuntu 22.04 LTS (jammy)
- Ubuntu 24.04 LTS (noble)
- Debian 11 (bullseye)
- Debian 12 (bookworm)

## Troubleshooting

### Build Fails
- Check build logs at https://launchpad.net/~yourusername/+archive/ubuntu/systask
- Common issues:
  - Missing dependencies in `debian/control`
  - Go version too old
  - Missing build files

### GPG Key Issues
```bash
# List your keys
gpg --list-keys

# If key not on keyserver
gpg --send-keys --keyserver keyserver.ubuntu.com YOUR_KEY_ID

# Wait 5-10 minutes for keyserver sync
```

### Upload Issues
```bash
# Check dput configuration
cat ~/.dput.cf

# Test upload (dry-run)
dput -s ppa:yourusername/systask ../systask_1.0.2-1_source.changes
```

## Maintenance

### Update to New Version

```bash
# Update version in debian/changelog
dch -i

# Build and upload
debuild -S -sa
dput ppa:yourusername/systask ../systask_1.0.3-1_source.changes
```

### Delete Old Versions

1. Go to PPA page
2. Click on the version
3. Click "Delete packages"

## References

- [Launchpad PPA Guide](https://help.launchpad.net/Packaging/PPA)
- [Debian Packaging Guide](https://www.debian.org/doc/manuals/debian-faq/pkg-basics.en.html)
- [dput Documentation](https://manpages.ubuntu.com/manpages/focal/man1/dput.1.html)

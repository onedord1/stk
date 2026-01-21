# AUR Submission Guide for SysTask

## Prerequisites

1. **Create AUR Account**
   - Go to https://aur.archlinux.org/register/
   - Create your account
   - Set up SSH key for AUR access

2. **Generate SSH Key for AUR**
   ```bash
   ssh-keygen -t ed25519 -f ~/.ssh/aur -C "aur@systask"
   ```

3. **Add SSH Key to AUR Account**
   - Log in to https://aur.archlinux.org
   - Go to Account Settings
   - Paste your public key (~/.ssh/aur.pub)

4. **Configure SSH for AUR**
   ```bash
   cat >> ~/.ssh/config << 'EOF'
   Host aur.archlinux.org
       IdentityFile ~/.ssh/aur
       User aur
   EOF
   chmod 600 ~/.ssh/config
   ```

## Submit to AUR

### First Time Submission

```bash
# Clone the AUR repository (creates new package)
git clone ssh://aur@aur.archlinux.org/systask.git
cd systask

# Copy PKGBUILD and .SRCINFO
cp ../PKGBUILD .
cp ../.SRCINFO .

# Verify the package
makepkg --printsrcinfo > .SRCINFO

# Commit and push
git add PKGBUILD .SRCINFO
git commit -m "Initial commit for systask v1.0.2"
git push
```

### Update Existing Package

```bash
cd systask

# Update PKGBUILD with new version
nano PKGBUILD

# Regenerate .SRCINFO
makepkg --printsrcinfo > .SRCINFO

# Commit and push
git add PKGBUILD .SRCINFO
git commit -m "Update to v1.0.3"
git push
```

## Installation from AUR

Once submitted and approved, users can install with:

```bash
# Using yay
yay -S systask

# Using paru
paru -S systask

# Manual installation
git clone https://aur.archlinux.org/systask.git
cd systask
makepkg -si
```

## AUR Maintenance

### Regular Updates
After each release:
1. Update version in PKGBUILD
2. Update sha256sums (or set to SKIP)
3. Regenerate .SRCINFO
4. Commit with message: "Update to vX.Y.Z"
5. Push to AUR

### Package Deletion
If needed to remove from AUR:
```bash
cd systask
git rm -r .
git commit -m "Remove package"
git push
```

## Troubleshooting

### SSH Connection Issues
```bash
# Test SSH connection
ssh -T aur@aur.archlinux.org

# If fails, verify SSH key is added to AUR account
ssh-add ~/.ssh/aur
```

### SRCINFO Generation Fails
```bash
# Ensure PKGBUILD is valid
makepkg --printsrcinfo

# If it fails, check PKGBUILD syntax
bash -n PKGBUILD
```

### Package Conflicts
If the package name already exists:
- Check if it's maintained
- Contact the maintainer
- Request takeover if unmaintained

## References

- [AUR Submission Guidelines](https://wiki.archlinux.org/title/AUR_submission_guidelines)
- [PKGBUILD Reference](https://wiki.archlinux.org/title/PKGBUILD)
- [AUR FAQ](https://wiki.archlinux.org/title/AUR#FAQ)

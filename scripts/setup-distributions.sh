#!/bin/bash
# Setup script for PPA and Copr distribution

set -e

VERSION="1.0.3"
GITHUB_USER="onedord1"
GITHUB_REPO="stk"
LAUNCHPAD_USER="${LAUNCHPAD_USER:-}"
COPR_USER="${COPR_USER:-}"

echo "🚀 SysTask Distribution Setup"
echo "=============================="
echo ""

# Function to setup PPA
setup_ppa() {
    echo "📦 Setting up Launchpad PPA"
    echo "============================"
    echo ""
    
    if [ -z "$LAUNCHPAD_USER" ]; then
        echo "⚠️  LAUNCHPAD_USER environment variable not set"
        echo "Set it with: export LAUNCHPAD_USER=yourusername"
        echo ""
        echo "Steps to set up PPA manually:"
        echo "1. Go to https://launchpad.net/~yourusername/+create-ppa"
        echo "2. Create PPA named 'systask'"
        echo "3. Follow instructions in packaging/debian/PPA_SETUP.md"
        return 1
    fi
    
    echo "✓ Launchpad user: $LAUNCHPAD_USER"
    echo ""
    echo "Next steps:"
    echo "1. Ensure you have GPG key set up:"
    echo "   gpg --gen-key"
    echo "   gpg --send-keys --keyserver keyserver.ubuntu.com YOUR_KEY_ID"
    echo ""
    echo "2. Install build tools:"
    echo "   sudo apt-get install dput devscripts ubuntu-dev-tools"
    echo ""
    echo "3. Build and upload source package:"
    echo "   cd ~/CascadeProjects/stk"
    echo "   debuild -S -sa"
    echo "   dput ppa:$LAUNCHPAD_USER/systask ../systask_${VERSION}-1_source.changes"
    echo ""
    echo "4. Verify at: https://launchpad.net/~$LAUNCHPAD_USER/+archive/ubuntu/systask"
    echo ""
}

# Function to setup Copr
setup_copr() {
    echo "📦 Setting up Fedora Copr"
    echo "=========================="
    echo ""
    
    if [ -z "$COPR_USER" ]; then
        echo "⚠️  COPR_USER environment variable not set"
        echo "Set it with: export COPR_USER=yourusername"
        echo ""
        echo "Steps to set up Copr manually:"
        echo "1. Go to https://copr.fedorainfracloud.org/coprs/create/"
        echo "2. Create project named 'systask'"
        echo "3. Enable GitHub webhook"
        echo "4. Follow instructions in packaging/rpm/COPR_SETUP.md"
        return 1
    fi
    
    echo "✓ Copr user: $COPR_USER"
    echo ""
    echo "Next steps:"
    echo "1. Install copr-cli:"
    echo "   sudo dnf install copr-cli"
    echo ""
    echo "2. Configure copr-cli:"
    echo "   copr-cli config"
    echo ""
    echo "3. Create Copr project:"
    echo "   copr-cli create systask \\"
    echo "     --chroot fedora-39-x86_64 \\"
    echo "     --chroot fedora-39-aarch64 \\"
    echo "     --chroot fedora-40-x86_64 \\"
    echo "     --chroot fedora-40-aarch64 \\"
    echo "     --chroot epel-8-x86_64 \\"
    echo "     --chroot epel-8-aarch64 \\"
    echo "     --chroot epel-9-x86_64 \\"
    echo "     --chroot epel-9-aarch64"
    echo ""
    echo "4. Trigger first build:"
    echo "   copr-cli build-package systask \\"
    echo "     --git-url https://github.com/$GITHUB_USER/$GITHUB_REPO \\"
    echo "     --git-branch stable \\"
    echo "     --spec packaging/rpm/systask.spec"
    echo ""
    echo "5. Verify at: https://copr.fedorainfracloud.org/coprs/$COPR_USER/systask/"
    echo ""
}

# Function to show installation commands
show_install_commands() {
    echo "📥 Installation Commands for Users"
    echo "==================================="
    echo ""
    
    if [ -n "$LAUNCHPAD_USER" ]; then
        echo "Ubuntu/Debian (via PPA):"
        echo "  sudo add-apt-repository ppa:$LAUNCHPAD_USER/systask"
        echo "  sudo apt update"
        echo "  sudo apt install systask"
        echo ""
    fi
    
    if [ -n "$COPR_USER" ]; then
        echo "Fedora/RHEL (via Copr):"
        echo "  sudo dnf copr enable $COPR_USER/systask"
        echo "  sudo dnf install systask"
        echo ""
    fi
    
    echo "Arch Linux (via AUR):"
    echo "  yay -S systask"
    echo "  # or"
    echo "  git clone https://aur.archlinux.org/systask.git && cd systask && makepkg -si"
    echo ""
}

# Main execution
echo "Choose setup option:"
echo "1) Setup PPA (Ubuntu/Debian)"
echo "2) Setup Copr (Fedora/RHEL)"
echo "3) Setup both"
echo "4) Show installation commands"
echo "5) Exit"
echo ""

read -p "Enter choice (1-5): " choice

case $choice in
    1)
        setup_ppa
        ;;
    2)
        setup_copr
        ;;
    3)
        setup_ppa
        echo ""
        setup_copr
        ;;
    4)
        show_install_commands
        ;;
    5)
        echo "Exiting..."
        exit 0
        ;;
    *)
        echo "Invalid choice"
        exit 1
        ;;
esac

echo ""
echo "📚 For detailed instructions, see:"
echo "  - PPA: packaging/debian/PPA_SETUP.md"
echo "  - Copr: packaging/rpm/COPR_SETUP.md"
echo ""

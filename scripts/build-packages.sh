#!/bin/bash
# Build SysTask packages for multiple distributions

set -e

VERSION="1.0.2"
PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BUILD_DIR="${PROJECT_DIR}/build"
DIST_DIR="${PROJECT_DIR}/dist"

echo "🔨 Building SysTask v${VERSION}"
echo "================================"

# Create directories
mkdir -p "${BUILD_DIR}" "${DIST_DIR}"

# Build binary
echo "📦 Building binary..."
cd "${PROJECT_DIR}"
go build -o "${BUILD_DIR}/systask" ./main.go
echo "✓ Binary built: ${BUILD_DIR}/systask"

# Build Debian package
build_deb() {
    echo ""
    echo "📦 Building Debian package..."
    
    DEB_DIR="${BUILD_DIR}/deb-build"
    rm -rf "${DEB_DIR}"
    mkdir -p "${DEB_DIR}/systask/usr/bin"
    mkdir -p "${DEB_DIR}/systask/etc/systask"
    mkdir -p "${DEB_DIR}/systask/usr/share/doc/systask"
    mkdir -p "${DEB_DIR}/systask/DEBIAN"
    
    # Copy files
    cp "${BUILD_DIR}/systask" "${DEB_DIR}/systask/usr/bin/"
    cp "${PROJECT_DIR}/config.yaml.example" "${DEB_DIR}/systask/etc/systask/"
    cp "${PROJECT_DIR}/README.md" "${DEB_DIR}/systask/usr/share/doc/systask/"
    cp "${PROJECT_DIR}/LICENSE" "${DEB_DIR}/systask/usr/share/doc/systask/"
    
    # Create control file
    cat > "${DEB_DIR}/systask/DEBIAN/control" << EOF
Package: systask
Version: ${VERSION}
Architecture: amd64
Maintainer: SysTask Team <team@systask.dev>
Homepage: https://github.com/onedord1/stk
Description: Terminal-based system management tool for Linux servers
 SysTask is a comprehensive terminal UI application for managing Linux servers.
Depends: openssh-client
Priority: optional
Section: admin
EOF

    # Create postinst script
    cat > "${DEB_DIR}/systask/DEBIAN/postinst" << 'EOF'
#!/bin/bash
chmod +x /usr/bin/systask
mkdir -p /etc/systask
echo "✓ SysTask installed successfully!"
echo "Run 'systask' to start."
EOF
    chmod 755 "${DEB_DIR}/systask/DEBIAN/postinst"
    
    # Build package
    dpkg-deb --build "${DEB_DIR}/systask" "${DIST_DIR}/systask_${VERSION}_amd64.deb"
    echo "✓ Debian package: ${DIST_DIR}/systask_${VERSION}_amd64.deb"
}

# Build RPM package
build_rpm() {
    echo ""
    echo "📦 Building RPM package..."
    
    RPM_DIR="${BUILD_DIR}/rpm-build"
    rm -rf "${RPM_DIR}"
    mkdir -p "${RPM_DIR}/"{BUILD,RPMS,SOURCES,SPECS,SRPMS}
    
    # Create spec file
    cat > "${RPM_DIR}/SPECS/systask.spec" << 'EOF'
Name:           systask
Version:        1.0.2
Release:        1%{?dist}
Summary:        Terminal-based system management tool for Linux servers
License:        MIT
URL:            https://github.com/onedord1/stk
BuildRequires:  golang >= 1.16
Requires:       openssh-clients

%description
SysTask is a comprehensive terminal UI application for managing Linux servers.

%prep
%setup -q -n stk-%{version}

%build
go build -o systask ./main.go

%install
install -Dm755 systask %{buildroot}%{_bindir}/systask
install -Dm644 config.yaml.example %{buildroot}%{_sysconfdir}/systask/config.yaml.example
install -Dm644 README.md %{buildroot}%{_docdir}/systask/README.md
install -Dm644 LICENSE %{buildroot}%{_docdir}/systask/LICENSE

%files
%{_bindir}/systask
%{_sysconfdir}/systask/config.yaml.example
%{_docdir}/systask/README.md
%{_docdir}/systask/LICENSE

%changelog
* Wed Jan 22 2026 SysTask Team <team@systask.dev> - 1.0.2-1
- Fixed SFTP copy/paste functionality
- Fixed progress bar display
EOF

    # Create source tarball
    cd "${PROJECT_DIR}"
    tar czf "${RPM_DIR}/SOURCES/stk-${VERSION}.tar.gz" \
        --exclude=.git \
        --exclude=.github \
        --exclude=build \
        --exclude=dist \
        --transform="s,^,stk-${VERSION}/," .
    
    # Build RPM
    rpmbuild -bb \
        --define="_topdir ${RPM_DIR}" \
        "${RPM_DIR}/SPECS/systask.spec" 2>/dev/null || {
        echo "⚠ RPM build requires rpmbuild. Install with:"
        echo "  sudo apt-get install rpm  # Debian/Ubuntu"
        echo "  sudo dnf install rpm-build  # Fedora"
        return 1
    }
    
    # Copy RPM to dist
    cp "${RPM_DIR}/RPMS"/*/*.rpm "${DIST_DIR}/" 2>/dev/null || true
    echo "✓ RPM package built"
}

# Build Arch package
build_arch() {
    echo ""
    echo "📦 Building Arch package..."
    
    ARCH_DIR="${BUILD_DIR}/arch-build"
    rm -rf "${ARCH_DIR}"
    mkdir -p "${ARCH_DIR}"
    
    # Create PKGBUILD
    cat > "${ARCH_DIR}/PKGBUILD" << 'EOF'
pkgname=systask
pkgver=1.0.2
pkgrel=1
pkgdesc="Terminal-based system management tool for Linux servers"
arch=('x86_64' 'aarch64')
url="https://github.com/onedord1/stk"
license=('MIT')
depends=('openssh')
makedepends=('go')
source=("${pkgname}-${pkgver}.tar.gz::https://github.com/onedord1/stk/archive/v${pkgver}.tar.gz")
sha256sums=('SKIP')

build() {
    cd "stk-${pkgver}"
    go build -o systask ./main.go
}

package() {
    cd "stk-${pkgver}"
    install -Dm755 systask "${pkgdir}/usr/bin/systask"
    install -Dm644 config.yaml.example "${pkgdir}/etc/systask/config.yaml.example"
    install -Dm644 README.md "${pkgdir}/usr/share/doc/systask/README.md"
    install -Dm644 LICENSE "${pkgdir}/usr/share/licenses/systask/LICENSE"
}
EOF

    cp "${ARCH_DIR}/PKGBUILD" "${DIST_DIR}/"
    echo "✓ Arch PKGBUILD: ${DIST_DIR}/PKGBUILD"
}

# Build all packages
build_deb || echo "⚠ Debian build skipped"
build_rpm || echo "⚠ RPM build skipped"
build_arch

echo ""
echo "================================"
echo "✅ Build complete!"
echo ""
echo "📁 Output directory: ${DIST_DIR}/"
ls -lh "${DIST_DIR}/"
echo ""
echo "📦 Package files ready for distribution:"
echo "  - Debian: systask_${VERSION}_amd64.deb"
echo "  - RPM: systask-${VERSION}-1.x86_64.rpm"
echo "  - Arch: PKGBUILD (for AUR submission)"

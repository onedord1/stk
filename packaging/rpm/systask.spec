Name:           systask
Version:        1.0.2
Release:        1%{?dist}
Summary:        Terminal-based system management tool for Linux servers

License:        MIT
URL:            https://github.com/onedord1/stk
Source0:        https://github.com/onedord1/stk/archive/v%{version}.tar.gz

BuildRequires:  golang >= 1.16
Requires:       openssh-clients

%description
SysTask is a comprehensive terminal UI application for managing Linux servers.

Features:
- Real-time system health monitoring (CPU, Memory, Disk, Network)
- Service management (start/stop/restart systemd services)
- Process viewer and management
- Log viewer with real-time following
- Disk management with partition operations
- Batch command execution across multiple servers
- User management
- Docker container management
- Package installer
- SFTP file manager with copy/paste operations
- SSH terminal access
- 12 built-in color themes
- Host management with SSH config integration
- Encrypted credential storage

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
- Enhanced transfer queue UI

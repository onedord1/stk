# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 1.0.x   | :white_check_mark: |

## Reporting a Vulnerability

If you discover a security vulnerability in SysTask, please report it responsibly:

1. **Do NOT** create a public GitHub issue
2. Email security concerns to: [your-email@example.com]
3. Include:
   - Description of the vulnerability
   - Steps to reproduce
   - Potential impact
   - Suggested fix (if any)

We will respond within 48 hours and work with you to address the issue.

## Security Features

### Credential Storage
- Passwords are encrypted with **AES-256-GCM**
- Key derivation uses **PBKDF2** with 100,000 iterations
- Unique salt generated per installation
- No plaintext passwords stored

### SSH Security
- Supports SSH key authentication (recommended)
- Supports SSH agent forwarding
- Respects `~/.ssh/config` settings
- No credentials logged or transmitted insecurely

### Best Practices
1. Use SSH keys instead of passwords
2. Store SSH keys with proper permissions (600)
3. Use a strong master password if using password storage
4. Keep SysTask updated to latest version

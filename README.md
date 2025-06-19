# SWUpdate CLI Client

[![CI](https://github.com/DatanoiseTV/swupdate-cli/actions/workflows/ci.yml/badge.svg)](https://github.com/DatanoiseTV/swupdate-cli/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/DatanoiseTV/swupdate-cli/branch/main/graph/badge.svg)](https://codecov.io/gh/DatanoiseTV/swupdate-cli)
[![Go Report Card](https://goreportcard.com/badge/github.com/DatanoiseTV/swupdate-cli)](https://goreportcard.com/report/github.com/DatanoiseTV/swupdate-cli)
[![License](https://img.shields.io/badge/license-BSD--3--Clause-blue.svg)](LICENSE)

A command-line client for uploading firmware to devices running [SWUpdate](https://sbabic.github.io/swupdate/). This tool provides a simple interface for firmware updates with real-time progress monitoring via WebSocket connections.

## Features

- **Firmware Upload**: Upload `.swu` firmware files to SWUpdate-enabled devices
- **Real-time Progress**: Monitor update progress through WebSocket connections
- **JSON Output**: Machine-parseable output for automation and logging
- **Device Restart**: Optional device restart after successful updates
- **TLS/SSL Support**: Secure connections with certificate verification
- **Certificate Management**: Custom CA certificates and client certificate authentication
- **Error Handling**: Comprehensive error reporting and timeout management
- **Verbose Logging**: Detailed output for debugging and monitoring

## Installation

### From Source

```bash
git clone https://github.com/DatanoiseTV/swupdate-cli.git
cd swupdate-cli
go build -o swupdate-client swupdate-client.go
```

### Binary Release

Download the pre-built static binary for your platform from the [releases page](https://github.com/DatanoiseTV/swupdate-cli/releases).

Each release includes:
- Static binaries for all platforms (no dependencies required)
- Checksum files (SHA256, SHA512, MD5) for verification
- Test coverage report (HTML and text format)

#### Linux (x86_64/amd64)
```bash
# Download binary and checksums
wget https://github.com/DatanoiseTV/swupdate-cli/releases/latest/download/swupdate-client-linux-x86_64
wget https://github.com/DatanoiseTV/swupdate-cli/releases/latest/download/checksums.sha256

# Verify checksum
sha256sum -c checksums.sha256 --ignore-missing

# Install system-wide
sudo install -m 755 swupdate-client-linux-x86_64 /usr/local/bin/swupdate-client

# Or install for current user
mkdir -p ~/.local/bin
cp swupdate-client-linux-x86_64 ~/.local/bin/swupdate-client
chmod +x ~/.local/bin/swupdate-client
```

#### Linux (ARM64)
```bash
wget https://github.com/DatanoiseTV/swupdate-cli/releases/latest/download/swupdate-client-linux-arm64
sudo install -m 755 swupdate-client-linux-arm64 /usr/local/bin/swupdate-client
```

#### Linux (ARM 32-bit)
```bash
# ARMv7
wget https://github.com/DatanoiseTV/swupdate-cli/releases/latest/download/swupdate-client-linux-armv7

# ARMv6 (Raspberry Pi Zero/1)
wget https://github.com/DatanoiseTV/swupdate-cli/releases/latest/download/swupdate-client-linux-armv6

sudo install -m 755 swupdate-client-linux-armv* /usr/local/bin/swupdate-client
```

#### macOS (Intel)
```bash
# Download
curl -LO https://github.com/DatanoiseTV/swupdate-cli/releases/latest/download/swupdate-client-darwin-x86_64

# Make executable
chmod +x swupdate-client-darwin-x86_64

# Install
sudo mv swupdate-client-darwin-x86_64 /usr/local/bin/swupdate-client

# Or using Homebrew custom tap (if available)
# brew install datanoisetv/tap/swupdate-client
```

#### macOS (Apple Silicon M1/M2/M3)
```bash
# Download
curl -LO https://github.com/DatanoiseTV/swupdate-cli/releases/latest/download/swupdate-client-darwin-arm64

# Make executable
chmod +x swupdate-client-darwin-arm64

# Install
sudo mv swupdate-client-darwin-arm64 /usr/local/bin/swupdate-client
```

#### Windows (64-bit)
```powershell
# PowerShell - Download
Invoke-WebRequest -Uri "https://github.com/DatanoiseTV/swupdate-cli/releases/latest/download/swupdate-client-windows-x86_64.exe" -OutFile "swupdate-client.exe"

# Verify checksum (PowerShell)
Invoke-WebRequest -Uri "https://github.com/DatanoiseTV/swupdate-cli/releases/latest/download/checksums.sha256" -OutFile "checksums.sha256"
(Get-FileHash swupdate-client.exe -Algorithm SHA256).Hash

# Add to PATH or move to a directory in PATH
# Example: C:\Program Files\swupdate-client\
New-Item -ItemType Directory -Force -Path "C:\Program Files\swupdate-client"
Move-Item swupdate-client.exe "C:\Program Files\swupdate-client\"
# Add "C:\Program Files\swupdate-client" to system PATH
```

#### Windows (32-bit)
```powershell
# PowerShell
Invoke-WebRequest -Uri "https://github.com/DatanoiseTV/swupdate-cli/releases/latest/download/swupdate-client-windows-386.exe" -OutFile "swupdate-client.exe"
```

#### FreeBSD / OpenBSD / NetBSD
```bash
# Build from source is recommended for BSD systems
git clone https://github.com/DatanoiseTV/swupdate-cli.git
cd swupdate-cli
go build -o swupdate-client swupdate-client.go
sudo install -m 755 swupdate-client /usr/local/bin/
```

#### Docker
```bash
# Run directly from Docker (example)
docker run --rm -v $(pwd):/firmware \
  ghcr.io/datanoisetv/swupdate-client:latest \
  -ip 192.168.1.100 -file /firmware/update.swu
```

#### Embedded Linux (Yocto/Buildroot)
For embedded systems, download the appropriate architecture:
- ARMv6: `swupdate-client-linux-armv6`
- ARMv7: `swupdate-client-linux-armv7`
- ARM64: `swupdate-client-linux-arm64`
- x86_64: `swupdate-client-linux-x86_64`

```bash
# Example for ARMv7 embedded device
wget -O /usr/bin/swupdate-client https://github.com/DatanoiseTV/swupdate-cli/releases/latest/download/swupdate-client-linux-armv7
chmod +x /usr/bin/swupdate-client
```

## Usage

### Basic Usage

```bash
# Default: HTTP without TLS (typical for local/embedded devices)
./swupdate-client -ip 192.168.1.100 -file firmware.swu
```

### Complete Example

```bash
./swupdate-client -ip 192.168.1.100 -port 8080 -file firmware.swu -restart -verbose
```

### JSON Output for Automation

```bash
./swupdate-client -ip 192.168.1.100 -file firmware.swu -json > update.log
```

### Command Line Options

| Flag | Default | Description |
|------|---------|-------------|
| `-ip` | `192.168.1.100` | IP address of the SWUpdate device |
| `-port` | `8080` | Port of the SWUpdate web server |
| `-file` | (required) | Firmware file (.swu) to upload |
| `-timeout` | `5m0s` | Timeout for operations |
| `-verbose` | `false` | Enable verbose output |
| `-json` | `false` | Output progress and messages in JSON format |
| `-tls` | `false` | Use HTTPS/WSS instead of HTTP/WS (default is HTTP) |
| `-insecure` | `false` | Skip TLS certificate verification (only with -tls) |
| `-ca-cert` | | Path to custom CA certificate file |
| `-client-cert` | | Path to client certificate file |
| `-client-key` | | Path to client private key file |
| `-restart` | `false` | Restart device after successful update |

## JSON Output Format

When using the `-json` flag, the client outputs structured JSON messages:

### Log Messages
```json
{
  "type": "upload",
  "level": "INFO", 
  "message": "Uploading firmware: firmware.swu (2.34 MB)",
  "time": "2023-12-01T10:30:00Z"
}
```

### WebSocket Events
```json
{
  "type": "status",
  "status": "START",
  "level": "INFO"
}
```

### Progress Updates
```json
{
  "type": "step",
  "name": "kernel",
  "percent": "75"
}
```

## Development

### Requirements

- Go 1.19 or later
- Dependencies managed via Go modules

### Building

```bash
go build -o swupdate-client swupdate-client.go
```

### Testing

```bash
go test -v
```

### Dependencies

- [gorilla/websocket](https://github.com/gorilla/websocket) - WebSocket client implementation

## SWUpdate Server Setup

This client is designed to work with SWUpdate's built-in web server. By default, it uses HTTP for embedded devices. Ensure your SWUpdate configuration includes:

```
# Basic HTTP setup (default)
webserver = {
    listen = ":8080";
    document_root = "/www";
};

# Optional: HTTPS setup (use with -tls flag)
webserver = {
    listen = ":8443";
    document_root = "/www";
    ssl = {
        cert = "/etc/ssl/certs/server.crt";
        key = "/etc/ssl/private/server.key";
    };
};
```

## Error Handling

The client handles various error conditions:

- **Network timeouts**: Configurable timeout for all operations
- **File not found**: Validates firmware file exists before upload
- **Upload failures**: Reports HTTP status codes and error messages
- **WebSocket disconnection**: Continues operation if WebSocket fails
- **Device restart failures**: Reports but doesn't fail the update

## Exit Codes

- `0`: Success
- `1`: Error (file not found, upload failed, etc.)

## Examples

### Basic Firmware Update
```bash
./swupdate-client -ip 10.0.0.100 -file my-firmware.swu
```

### Update with Device Restart
```bash
./swupdate-client -ip 10.0.0.100 -file my-firmware.swu -restart
```

### Automated Update with Logging
```bash
./swupdate-client -ip 10.0.0.100 -file my-firmware.swu -json | tee update-$(date +%Y%m%d).log
```

### Secure Update with TLS
```bash
./swupdate-client -ip 10.0.0.100 -file my-firmware.swu -tls -ca-cert ca.crt
```

### Update with Client Certificate Authentication
```bash
./swupdate-client -ip 10.0.0.100 -file my-firmware.swu -tls -client-cert client.crt -client-key client.key
```

### Update with Custom Timeout
```bash
./swupdate-client -ip 10.0.0.100 -file large-firmware.swu -timeout 10m
```

## License

This project is licensed under the BSD 3-Clause License with attribution requirement - see the [LICENSE](LICENSE) file for details.

**Important**: Any product that uses this software must include the following acknowledgment visibly in its documentation and any releases:
"This product includes software developed by DatanoiseTV / Datanoise UG (haftungsbeschränkt)."

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass
6. Submit a pull request

## Support

For issues and questions:

1. Check the [SWUpdate documentation](https://sbabic.github.io/swupdate/)
2. Review existing issues in this repository
3. Create a new issue with detailed information about your problem
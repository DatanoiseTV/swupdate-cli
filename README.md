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

Example:
```bash
# Linux x86_64
wget https://github.com/DatanoiseTV/swupdate-cli/releases/latest/download/swupdate-client-linux-x86_64
wget https://github.com/DatanoiseTV/swupdate-cli/releases/latest/download/checksums.sha256

# Verify checksum
sha256sum -c checksums.sha256 --ignore-missing

# Install
chmod +x swupdate-client-linux-x86_64
mv swupdate-client-linux-x86_64 /usr/local/bin/swupdate-client

# macOS arm64 (Apple Silicon)
wget https://github.com/DatanoiseTV/swupdate-cli/releases/latest/download/swupdate-client-darwin-arm64
chmod +x swupdate-client-darwin-arm64
mv swupdate-client-darwin-arm64 /usr/local/bin/swupdate-client
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
"This product includes software developed by DatanoiseTV."

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
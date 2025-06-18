# SWUpdate CLI Client

A command-line client for uploading firmware to devices running [SWUpdate](https://sbabic.github.io/swupdate/). This tool provides a simple interface for firmware updates with real-time progress monitoring via WebSocket connections.

## Features

- **Firmware Upload**: Upload `.swu` firmware files to SWUpdate-enabled devices
- **Real-time Progress**: Monitor update progress through WebSocket connections
- **JSON Output**: Machine-parseable output for automation and logging
- **Device Restart**: Optional device restart after successful updates
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

Download the latest binary from the [releases page](https://github.com/DatanoiseTV/swupdate-cli/releases).

## Usage

### Basic Usage

```bash
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

This client is designed to work with SWUpdate's built-in web server. Ensure your SWUpdate configuration includes:

```
webserver = {
    listen = ":8080";
    document_root = "/www";
};

suricatta = {
    enable = true;
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

### Update with Custom Timeout
```bash
./swupdate-client -ip 10.0.0.100 -file large-firmware.swu -timeout 10m
```

## License

This project is licensed under the MIT License - see the LICENSE file for details.

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
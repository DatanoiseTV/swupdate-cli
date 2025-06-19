// SWUpdate CLI Client provides firmware upload capabilities for SWUpdate-enabled devices
package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/gorilla/websocket"
)

// Build-time variables (set via ldflags)
var (
	version   = "dev"
	commit    = "unknown"
	branch    = "unknown"
	buildDate = "unknown"
)

// Config holds all configuration parameters for the SWUpdate client
type Config struct {
	IPAddress      string        // Target device IP address
	Port           int           // SWUpdate web server port
	Filename       string        // Path to firmware file (.swu)
	Timeout        time.Duration // Network operation timeout
	Verbose        bool          // Enable detailed logging
	JSONOutput     bool          // Output structured JSON instead of human-readable text
	TLS            bool          // Use HTTPS/WSS instead of HTTP/WS
	InsecureTLS    bool          // Skip TLS certificate verification
	CertFile       string        // Path to custom CA certificate file
	ClientCertFile string        // Path to client certificate file
	ClientKeyFile  string        // Path to client private key file
}

// SWUpdateEvent represents a WebSocket event from the SWUpdate server
type SWUpdateEvent struct {
	Type    string `json:"type"`              // Event type (status, step, message, etc.)
	Level   string `json:"level,omitempty"`   // Log level (INFO, WARN, ERROR)
	Text    string `json:"text,omitempty"`    // Human-readable message
	Number  string `json:"number,omitempty"`  // Step number
	Step    string `json:"step,omitempty"`    // Current step
	Name    string `json:"name,omitempty"`    // Package or component name
	Percent string `json:"percent,omitempty"` // Progress percentage
	Status  string `json:"status,omitempty"`  // Update status (START, RUN, SUCCESS, etc.)
	Source  string `json:"source,omitempty"`  // Update source information
}

// LogMessage represents a structured log entry for JSON output mode
type LogMessage struct {
	Type    string    `json:"type"`            // Message category
	Level   string    `json:"level,omitempty"` // Log level
	Message string    `json:"message"`         // Log message content
	Time    time.Time `json:"time"`            // Timestamp
}

// SWUpdateClient manages communication with an SWUpdate-enabled device
type SWUpdateClient struct {
	config Config          // Client configuration
	wsConn *websocket.Conn // WebSocket connection for progress monitoring
}

// NewSWUpdateClient creates a new client instance with the given configuration
func NewSWUpdateClient(config Config) *SWUpdateClient {
	return &SWUpdateClient{
		config: config,
	}
}

// createTLSConfig creates a TLS configuration based on the client settings
func (c *SWUpdateClient) createTLSConfig() (*tls.Config, error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: c.config.InsecureTLS,
	}

	// Load custom CA certificate if provided
	if c.config.CertFile != "" {
		caCert, err := os.ReadFile(c.config.CertFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA certificate file: %w", err)
		}

		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("failed to parse CA certificate")
		}
		tlsConfig.RootCAs = caCertPool
	}

	// Load client certificate and key if provided
	if c.config.ClientCertFile != "" && c.config.ClientKeyFile != "" {
		clientCert, err := tls.LoadX509KeyPair(c.config.ClientCertFile, c.config.ClientKeyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load client certificate: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{clientCert}
	}

	return tlsConfig, nil
}

// connectWebSocket establishes a WebSocket connection for real-time progress monitoring
func (c *SWUpdateClient) connectWebSocket(ctx context.Context) error {
	scheme := "ws"
	if c.config.TLS {
		scheme = "wss"
	}

	wsURL := url.URL{
		Scheme: scheme,
		Host:   fmt.Sprintf("%s:%d", c.config.IPAddress, c.config.Port),
		Path:   "/ws",
	}

	if c.config.Verbose {
		log.Printf("Connecting to WebSocket: %s", wsURL.String())
	}

	dialer := websocket.DefaultDialer
	dialer.HandshakeTimeout = c.config.Timeout

	// Configure TLS if enabled
	if c.config.TLS {
		tlsConfig, err := c.createTLSConfig()
		if err != nil {
			return fmt.Errorf("failed to create TLS configuration: %w", err)
		}
		dialer.TLSClientConfig = tlsConfig
	}

	conn, _, err := dialer.DialContext(ctx, wsURL.String(), nil)
	if err != nil {
		return fmt.Errorf("failed to connect to WebSocket: %w", err)
	}

	c.wsConn = conn
	return nil
}

func (c *SWUpdateClient) listenWebSocket(ctx context.Context) {
	wsConn := c.wsConn
	if wsConn == nil {
		return
	}

	defer func() {
		if wsConn != nil {
			wsConn.Close()
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			var event SWUpdateEvent
			err := wsConn.ReadJSON(&event)
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("WebSocket error: %v", err)
				}
				return
			}

			c.handleWebSocketEvent(event)
		}
	}
}

func (c *SWUpdateClient) logMessage(msgType, level, message string) {
	if c.config.JSONOutput {
		logMsg := LogMessage{
			Type:    msgType,
			Level:   level,
			Message: message,
			Time:    time.Now(),
		}
		jsonData, _ := json.Marshal(logMsg)
		fmt.Println(string(jsonData))
	} else {
		switch level {
		case "ERROR":
			fmt.Printf("Error: %s\n", message)
		case "WARN":
			fmt.Printf("Warning: %s\n", message)
		case "INFO":
			if c.config.Verbose || msgType == "status" || msgType == "progress" {
				fmt.Println(message)
			}
		default:
			fmt.Println(message)
		}
	}
}

func (c *SWUpdateClient) handleWebSocketEvent(event SWUpdateEvent) {
	if c.config.JSONOutput {
		jsonData, _ := json.Marshal(event)
		fmt.Println(string(jsonData))
		return
	}

	switch event.Type {
	case "status":
		c.handleStatusEvent(event)
	case "step":
		c.handleStepEvent(event)
	case "message":
		c.handleMessageEvent(event)
	case "info":
		c.handleInfoEvent(event)
	case "source":
		c.handleSourceEvent(event)
	default:
		c.handleUnknownEvent(event)
	}
}

func (c *SWUpdateClient) handleStatusEvent(event SWUpdateEvent) {
	statusMessages := map[string]struct {
		level   string
		message string
	}{
		"START":   {"INFO", "Update started"},
		"RUN":     {"INFO", "Update running"},
		"SUCCESS": {"INFO", "Update completed successfully"},
		"FAILURE": {"ERROR", "Update failed"},
		"DONE":    {"INFO", "Update process finished"},
		"IDLE":    {"INFO", "System idle"},
	}

	if msg, ok := statusMessages[event.Status]; ok {
		if event.Status == "IDLE" && !c.config.Verbose {
			return
		}
		c.logMessage("status", msg.level, msg.message)
	} else {
		c.logMessage("status", "INFO", fmt.Sprintf("Status: %s", event.Status))
	}
}

func (c *SWUpdateClient) handleStepEvent(event SWUpdateEvent) {
	if event.Percent != "" && event.Name != "" {
		c.logMessage("progress", "INFO", fmt.Sprintf("Installing %s: %s%%", event.Name, event.Percent))
	} else if event.Step != "" && event.Number != "" {
		c.logMessage("progress", "INFO", fmt.Sprintf("Step %s of %s", event.Step, event.Number))
	}
}

func (c *SWUpdateClient) handleMessageEvent(event SWUpdateEvent) {
	switch event.Level {
	case "ERROR":
		c.logMessage("message", "ERROR", event.Text)
	case "WARN":
		c.logMessage("message", "WARN", event.Text)
	default:
		if c.config.Verbose && event.Text != "" {
			c.logMessage("message", "INFO", event.Text)
		}
	}
}

func (c *SWUpdateClient) handleInfoEvent(event SWUpdateEvent) {
	if c.config.Verbose && event.Text != "" {
		c.logMessage("info", "INFO", event.Text)
	}
}

func (c *SWUpdateClient) handleSourceEvent(event SWUpdateEvent) {
	if c.config.Verbose {
		c.logMessage("source", "INFO", fmt.Sprintf("Update source: %s", event.Source))
	}
}

func (c *SWUpdateClient) handleUnknownEvent(event SWUpdateEvent) {
	if c.config.Verbose {
		c.logMessage("unknown", "INFO", fmt.Sprintf("Unknown event type: %s", event.Type))
	}
}

// uploadFirmware uploads the firmware file to the SWUpdate device via HTTP multipart form
func (c *SWUpdateClient) uploadFirmware(ctx context.Context) error {
	file, err := os.Open(c.config.Filename)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", c.config.Filename, err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file stats: %w", err)
	}

	c.logMessage("upload", "INFO", fmt.Sprintf("Uploading firmware: %s (%.2f MB)",
		filepath.Base(c.config.Filename),
		float64(stat.Size())/(1024*1024)))

	var requestBody bytes.Buffer
	multipartWriter := multipart.NewWriter(&requestBody)

	part, err := multipartWriter.CreateFormFile("file", filepath.Base(c.config.Filename))
	if err != nil {
		return fmt.Errorf("failed to create form file: %w", err)
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return fmt.Errorf("failed to copy file data: %w", err)
	}

	multipartWriter.Close()

	scheme := "http"
	if c.config.TLS {
		scheme = "https"
	}
	uploadURL := fmt.Sprintf("%s://%s:%d/upload", scheme, c.config.IPAddress, c.config.Port)

	req, err := http.NewRequestWithContext(ctx, "POST", uploadURL, &requestBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", multipartWriter.FormDataContentType())

	// Create HTTP client with TLS configuration
	client := &http.Client{
		Timeout: c.config.Timeout,
	}

	if c.config.TLS {
		tlsConfig, err := c.createTLSConfig()
		if err != nil {
			return fmt.Errorf("failed to create TLS configuration: %w", err)
		}
		client.Transport = &http.Transport{
			TLSClientConfig: tlsConfig,
		}
	}

	if c.config.Verbose {
		log.Printf("Uploading to: %s", uploadURL)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to upload firmware: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload failed with status %d: %s", resp.StatusCode, string(body))
	}

	c.logMessage("upload", "INFO", "Firmware uploaded successfully")
	return nil
}

func (c *SWUpdateClient) restartDevice(ctx context.Context) error {
	scheme := "http"
	if c.config.TLS {
		scheme = "https"
	}
	restartURL := fmt.Sprintf("%s://%s:%d/restart", scheme, c.config.IPAddress, c.config.Port)

	req, err := http.NewRequestWithContext(ctx, "POST", restartURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create restart request: %w", err)
	}

	// Create HTTP client with TLS configuration
	client := &http.Client{
		Timeout: c.config.Timeout,
	}

	if c.config.TLS {
		tlsConfig, err := c.createTLSConfig()
		if err != nil {
			return fmt.Errorf("failed to create TLS configuration: %w", err)
		}
		client.Transport = &http.Transport{
			TLSClientConfig: tlsConfig,
		}
	}

	if c.config.Verbose {
		log.Printf("Sending restart request to: %s", restartURL)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to restart device: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("restart failed with status %d: %s", resp.StatusCode, string(body))
	}

	c.logMessage("restart", "INFO", "Device restart initiated")
	return nil
}

// Update performs the complete firmware update process including WebSocket monitoring and optional restart
func (c *SWUpdateClient) Update(ctx context.Context, restart bool) error {
	wsCtx, wsCancel := context.WithCancel(ctx)
	defer wsCancel()

	if err := c.connectWebSocket(wsCtx); err != nil {
		log.Printf("Warning: Failed to connect to WebSocket: %v", err)
		log.Println("Proceeding without progress monitoring...")
	} else {
		go c.listenWebSocket(wsCtx)
	}

	if err := c.uploadFirmware(ctx); err != nil {
		return err
	}

	time.Sleep(2 * time.Second)

	if restart {
		if err := c.restartDevice(ctx); err != nil {
			log.Printf("Warning: Failed to restart device: %v", err)
		}
	}

	return nil
}

func main() {
	var config Config
	var restart bool
	var showVersion bool

	flag.StringVar(&config.IPAddress, "ip", "192.168.1.100", "IP address of the swupdate device")
	flag.IntVar(&config.Port, "port", 8080, "Port of the swupdate web server")
	flag.StringVar(&config.Filename, "file", "", "Firmware file (.swu) to upload")
	flag.DurationVar(&config.Timeout, "timeout", 5*time.Minute, "Timeout for operations")
	flag.BoolVar(&config.Verbose, "verbose", false, "Enable verbose output")
	flag.BoolVar(&config.JSONOutput, "json", false, "Output progress and messages in JSON format")
	flag.BoolVar(&config.TLS, "tls", false, "Use HTTPS/WSS instead of HTTP/WS")
	flag.BoolVar(&config.InsecureTLS, "insecure", false, "Skip TLS certificate verification")
	flag.StringVar(&config.CertFile, "ca-cert", "", "Path to custom CA certificate file")
	flag.StringVar(&config.ClientCertFile, "client-cert", "", "Path to client certificate file")
	flag.StringVar(&config.ClientKeyFile, "client-key", "", "Path to client private key file")
	flag.BoolVar(&restart, "restart", false, "Restart device after successful update")
	flag.BoolVar(&showVersion, "version", false, "Show version information")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "SWUpdate Client - Upload firmware to swupdate-capable devices\n")
		fmt.Fprintf(os.Stderr, "Version: %s (branch: %s, commit: %s, built: %s)\n\n", version, branch, commit, buildDate)
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s -ip 192.168.1.100 -file firmware.swu -restart\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -ip 192.168.1.100 -file firmware.swu -json > update.log\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -ip 192.168.1.100 -file firmware.swu -tls -ca-cert ca.crt\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -ip 192.168.1.100 -file firmware.swu -tls -insecure\n", os.Args[0])
	}

	flag.Parse()

	if showVersion {
		fmt.Printf("swupdate-client version %s\n", version)
		fmt.Printf("  Branch: %s\n", branch)
		fmt.Printf("  Commit: %s\n", commit)
		fmt.Printf("  Built:  %s\n", buildDate)
		os.Exit(0)
	}

	if config.Filename == "" {
		fmt.Fprintf(os.Stderr, "Error: firmware file (-file) is required\n\n")
		flag.Usage()
		os.Exit(1)
	}

	if _, err := os.Stat(config.Filename); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: firmware file '%s' does not exist\n", config.Filename)
		os.Exit(1)
	}

	client := NewSWUpdateClient(config)

	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()

	client.logMessage("connection", "INFO", fmt.Sprintf("Connecting to swupdate device at %s:%d", config.IPAddress, config.Port))

	if err := client.Update(ctx, restart); err != nil {
		fmt.Fprintf(os.Stderr, "Update failed: %v\n", err)
		os.Exit(1)
	}

	client.logMessage("completion", "INFO", "Update process completed")
}

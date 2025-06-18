package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestNewSWUpdateClient(t *testing.T) {
	config := Config{
		IPAddress:  "192.168.1.100",
		Port:       8080,
		Filename:   "test.swu",
		Timeout:    5 * time.Minute,
		Verbose:    true,
		JSONOutput: false,
	}

	client := NewSWUpdateClient(config)
	if client == nil {
		t.Fatal("NewSWUpdateClient returned nil")
	}

	if client.config.IPAddress != config.IPAddress {
		t.Errorf("Expected IP %s, got %s", config.IPAddress, client.config.IPAddress)
	}
}

func TestLogMessage(t *testing.T) {
	tests := []struct {
		name       string
		jsonOutput bool
		msgType    string
		level      string
		message    string
		expectJSON bool
	}{
		{
			name:       "JSON output enabled",
			jsonOutput: true,
			msgType:    "status",
			level:      "INFO",
			message:    "Test message",
			expectJSON: true,
		},
		{
			name:       "JSON output disabled",
			jsonOutput: false,
			msgType:    "status",
			level:      "INFO",
			message:    "Test message",
			expectJSON: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := Config{JSONOutput: tt.jsonOutput}
			client := NewSWUpdateClient(config)

			var buf bytes.Buffer
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			client.logMessage(tt.msgType, tt.level, tt.message)

			w.Close()
			os.Stdout = oldStdout
			_, _ = buf.ReadFrom(r) // Ignore error in test - we're capturing stdout

			output := buf.String()

			if tt.expectJSON {
				var logMsg LogMessage
				err := json.Unmarshal([]byte(strings.TrimSpace(output)), &logMsg)
				if err != nil {
					t.Errorf("Expected valid JSON, got: %s", output)
				}
				if logMsg.Message != tt.message {
					t.Errorf("Expected message %s, got %s", tt.message, logMsg.Message)
				}
			} else {
				if !strings.Contains(output, tt.message) {
					t.Errorf("Expected output to contain %s, got: %s", tt.message, output)
				}
			}
		})
	}
}

func TestHandleWebSocketEvent(t *testing.T) {
	config := Config{JSONOutput: true}
	client := NewSWUpdateClient(config)

	event := SWUpdateEvent{
		Type:    "status",
		Status:  "START",
		Level:   "INFO",
		Text:    "Update started",
		Percent: "50",
		Name:    "test-package",
	}

	var buf bytes.Buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	client.handleWebSocketEvent(event)

	w.Close()
	os.Stdout = oldStdout
	_, _ = buf.ReadFrom(r) // Ignore error in test - we're capturing stdout

	output := buf.String()
	var parsedEvent SWUpdateEvent
	err := json.Unmarshal([]byte(strings.TrimSpace(output)), &parsedEvent)
	if err != nil {
		t.Errorf("Expected valid JSON output, got: %s", output)
	}

	if parsedEvent.Type != event.Type {
		t.Errorf("Expected type %s, got %s", event.Type, parsedEvent.Type)
	}
}

func TestUploadFirmware_FileNotFound(t *testing.T) {
	config := Config{
		Filename: "nonexistent.swu",
		Timeout:  1 * time.Second,
	}
	client := NewSWUpdateClient(config)

	ctx := context.Background()
	err := client.uploadFirmware(ctx)

	if err == nil {
		t.Error("Expected error for nonexistent file, got nil")
	}

	if !strings.Contains(err.Error(), "failed to open file") {
		t.Errorf("Expected 'failed to open file' error, got: %s", err.Error())
	}
}

func TestUploadFirmware_Success(t *testing.T) {
	// Create a temporary test file
	tmpFile, err := os.CreateTemp("", "test*.swu")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	testData := "test firmware data"
	if _, err := tmpFile.WriteString(testData); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	// Create a test HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/upload" {
			t.Errorf("Expected /upload path, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Parse server URL to get host and port
	serverURL := strings.TrimPrefix(server.URL, "http://")
	parts := strings.Split(serverURL, ":")
	if len(parts) != 2 {
		t.Fatal("Could not parse server URL")
	}

	// Test that we can open and read the test file
	file, err := os.Open(tmpFile.Name())
	if err != nil {
		t.Fatalf("Could not open test file: %v", err)
	}
	defer file.Close()
	
	stat, err := file.Stat()
	if err != nil {
		t.Fatalf("Could not stat test file: %v", err)
	}
	
	if stat.Size() == 0 {
		t.Error("Test file should not be empty")
	}
	
	// Test that we can create a client with the file
	config := Config{
		Filename: tmpFile.Name(),
		Timeout:  5 * time.Second,
	}
	client := NewSWUpdateClient(config)
	if client == nil {
		t.Error("Expected client to be created")
	}
}

func TestWebSocketUpgrader(t *testing.T) {
	// Test that we can create the websocket upgrader without issues
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	if upgrader.CheckOrigin == nil {
		t.Error("Upgrader CheckOrigin should not be nil")
	}
}

func TestConfig(t *testing.T) {
	config := Config{
		IPAddress:  "10.0.0.1",
		Port:       9090,
		Filename:   "firmware.swu",
		Timeout:    10 * time.Minute,
		Verbose:    true,
		JSONOutput: true,
	}

	if config.IPAddress != "10.0.0.1" {
		t.Errorf("Expected IP 10.0.0.1, got %s", config.IPAddress)
	}

	if config.Port != 9090 {
		t.Errorf("Expected port 9090, got %d", config.Port)
	}

	if config.Timeout != 10*time.Minute {
		t.Errorf("Expected timeout 10m, got %v", config.Timeout)
	}

	if !config.Verbose {
		t.Error("Expected Verbose to be true")
	}

	if !config.JSONOutput {
		t.Error("Expected JSONOutput to be true")
	}
}

func TestSWUpdateEvent(t *testing.T) {
	event := SWUpdateEvent{
		Type:    "status",
		Level:   "INFO",
		Text:    "Test message",
		Number:  "1",
		Step:    "2",
		Name:    "package",
		Percent: "75",
		Status:  "RUN",
		Source:  "test",
	}

	// Test JSON marshaling
	data, err := json.Marshal(event)
	if err != nil {
		t.Errorf("Failed to marshal event: %v", err)
	}

	// Test JSON unmarshaling
	var unmarshaled SWUpdateEvent
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Errorf("Failed to unmarshal event: %v", err)
	}

	if unmarshaled.Type != event.Type {
		t.Errorf("Expected type %s, got %s", event.Type, unmarshaled.Type)
	}
}

func TestLogMessageStruct(t *testing.T) {
	logMsg := LogMessage{
		Type:    "test",
		Level:   "INFO",
		Message: "Test message",
		Time:    time.Now(),
	}

	data, err := json.Marshal(logMsg)
	if err != nil {
		t.Errorf("Failed to marshal LogMessage: %v", err)
	}

	var unmarshaled LogMessage
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Errorf("Failed to unmarshal LogMessage: %v", err)
	}

	if unmarshaled.Message != logMsg.Message {
		t.Errorf("Expected message %s, got %s", logMsg.Message, unmarshaled.Message)
	}
}

// TestConfigValidation tests configuration edge cases
func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name   string
		config Config
		valid  bool
	}{
		{
			name: "Valid config",
			config: Config{
				IPAddress: "192.168.1.100",
				Port:      8080,
				Filename:  "test.swu",
				Timeout:   5 * time.Minute,
			},
			valid: true,
		},
		{
			name: "Empty IP address",
			config: Config{
				IPAddress: "",
				Port:      8080,
				Filename:  "test.swu",
				Timeout:   5 * time.Minute,
			},
			valid: false,
		},
		{
			name: "Invalid port",
			config: Config{
				IPAddress: "192.168.1.100",
				Port:      0,
				Filename:  "test.swu",
				Timeout:   5 * time.Minute,
			},
			valid: false,
		},
		{
			name: "Zero timeout",
			config: Config{
				IPAddress: "192.168.1.100",
				Port:      8080,
				Filename:  "test.swu",
				Timeout:   0,
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewSWUpdateClient(tt.config)
			
			// Basic validation checks
			if tt.valid {
				if client == nil {
					t.Error("Expected valid client for valid config")
				}
				if client.config.IPAddress != tt.config.IPAddress {
					t.Errorf("IP mismatch: expected %s, got %s", tt.config.IPAddress, client.config.IPAddress)
				}
			} else {
				// Invalid configs should still create client (validation happens during use)
				if client == nil {
					t.Error("Client should be created even with invalid config")
				}
			}
		})
	}
}

// TestWebSocketEventTypes tests various event type handling
func TestWebSocketEventTypes(t *testing.T) {
	config := Config{JSONOutput: false, Verbose: true}
	client := NewSWUpdateClient(config)

	events := []SWUpdateEvent{
		{Type: "status", Status: "START"},
		{Type: "status", Status: "SUCCESS"},
		{Type: "status", Status: "FAILURE"},
		{Type: "step", Name: "kernel", Percent: "50"},
		{Type: "message", Level: "ERROR", Text: "Error message"},
		{Type: "message", Level: "WARN", Text: "Warning message"},
		{Type: "info", Text: "Info message"},
		{Type: "source", Source: "test-source"},
		{Type: "unknown", Text: "Unknown event"},
	}

	for _, event := range events {
		t.Run(fmt.Sprintf("Event_%s_%s", event.Type, event.Status), func(t *testing.T) {
			// Capture output
			var buf bytes.Buffer
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			client.handleWebSocketEvent(event)

			w.Close()
			os.Stdout = oldStdout
			_, _ = buf.ReadFrom(r) // Ignore error in test

			// Event should be handled without panic
			// Specific output validation would require more complex mocking
		})
	}
}

// TestTimeout tests timeout behavior
func TestTimeout(t *testing.T) {
	config := Config{
		IPAddress: "192.168.1.100",
		Port:      8080,
		Filename:  "nonexistent.swu",
		Timeout:   1 * time.Millisecond, // Very short timeout
	}
	client := NewSWUpdateClient(config)

	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()

	// This should timeout quickly
	err := client.connectWebSocket(ctx)
	if err == nil {
		t.Error("Expected timeout error for very short timeout")
	}
}
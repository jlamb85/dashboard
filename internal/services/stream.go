package services

import (
	"fmt"
	"net"
	"time"
)

// StreamInfo represents information about a stream on a port
type StreamInfo struct {
	Port        int
	Active      bool
	Protocol    string
	Details     string
	LastChecked time.Time
}

// CheckStreamOnPort checks if a stream is active on the specified port
func CheckStreamOnPort(port int) (*StreamInfo, error) {
	stream := &StreamInfo{
		Port:        port,
		Active:      false,
		Protocol:    "HTTP",
		Details:     "",
		LastChecked: time.Now(),
	}

	// Try to connect to the port
	address := fmt.Sprintf("localhost:%d", port)
	conn, err := net.DialTimeout("tcp", address, 2*time.Second)
	
	if err != nil {
		// Port is not accessible or no service is running
		stream.Active = false
		stream.Details = "No service detected on this port"
		return stream, nil
	}
	
	defer conn.Close()
	
	// If we can connect, assume the stream is active
	stream.Active = true
	stream.Details = fmt.Sprintf("Service is listening on port %d", port)
	
	return stream, nil
}

// GetStreamURL returns the URL for accessing the stream on a given port
func GetStreamURL(port int) string {
	return fmt.Sprintf("http://localhost:%d", port)
}

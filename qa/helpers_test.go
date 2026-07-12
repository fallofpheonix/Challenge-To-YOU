package qa

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

// netDialer wraps net.Dial for testing port availability.
type netDialer struct{}

func (d *netDialer) dial(port int) error {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("localhost:%d", port), 500*time.Millisecond)
	if err != nil {
		return err
	}
	conn.Close()
	return nil
}

// netDialTimeout checks if a TCP connection can be made.
func netDialTimeout(addr string, timeout time.Duration) (net.Conn, error) {
	return net.DialTimeout("tcp", addr, timeout)
}

// httpGetClient performs HTTP GET requests for testing.
type httpGetClient struct{}

func (c *httpGetClient) Get(url string) (string, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// runShellCommand executes a shell command with a timeout and returns combined output.
func runShellCommand(cmd string, timeout time.Duration) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return "", fmt.Errorf("empty command")
	}
	c := exec.CommandContext(ctx, parts[0], parts[1:]...)
	out, err := c.CombinedOutput()
	return string(out), err
}

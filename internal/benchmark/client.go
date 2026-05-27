package benchmark

import (
	"context"
	"fmt"
	"time"

	"github.com/talkincode/toughradius/v9/internal/radiusd"
	"layeh.com/radius"
)

// Client wraps the RADIUS client with additional functionality for benchmarking.
// It provides consistent timeout handling, retry logic, and error categorization.
type Client struct {
	radiusClient *radius.Client
	authAddr     string
	acctAddr     string
	timeout      time.Duration
}

// NewClient creates a new benchmark RADIUS client.
//
// Parameters:
//   - server: RADIUS server address (IP or hostname)
//   - authPort: Authentication port (typically 1812)
//   - acctPort: Accounting port (typically 1813)
//   - timeout: Request timeout duration
//
// Returns:
//   - *Client: Configured RADIUS client
func NewClient(server string, authPort, acctPort int, timeout time.Duration) *Client {
	return &Client{
		radiusClient: &radius.Client{
			Retry:              0, // No retries for accurate benchmarking
			MaxPacketErrors:    0,
			InsecureSkipVerify: true,
		},
		authAddr: fmt.Sprintf("%s:%d", server, authPort),
		acctAddr: fmt.Sprintf("%s:%d", server, acctPort),
		timeout:  timeout,
	}
}

// SendAuthRequest sends an Access-Request packet and waits for a response.
//
// The method tracks request and response bytes, latency, and error conditions.
// It updates the provided statistics object with the results.
//
// Parameters:
//   - packet: Access-Request packet to send
//   - stats: Statistics object to update with results
//
// Returns:
//   - *radius.Packet: Received response packet (nil on timeout/error)
//   - error: Error if the request fails or times out
func (c *Client) SendAuthRequest(packet *radius.Packet, stats *Statistics) (*radius.Packet, error) {
	stats.IncrAuthRequest()
	stats.AddRequestBytes(int64(radiusd.Length(packet)))

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	startTime := time.Now()
	resp, err := c.radiusClient.Exchange(ctx, packet, c.authAddr)
	latency := time.Since(startTime).Milliseconds()

	if err != nil {
		if err == context.DeadlineExceeded {
			stats.IncrAuthTimeout()
		}
		stats.IncrAuthDrop()
		return nil, err
	}

	if resp != nil {
		stats.AddResponseBytes(int64(radiusd.Length(resp)))
		stats.RecordAuthLatency(latency)

		switch resp.Code {
		case radius.CodeAccessAccept:
			stats.IncrAuthAccept()
		case radius.CodeAccessReject:
			stats.IncrAuthReject()
		default:
			stats.IncrAuthDrop()
			return resp, fmt.Errorf("unexpected response code: %v", resp.Code)
		}
	}

	return resp, nil
}

// SendAcctRequest sends an Accounting-Request packet and waits for a response.
//
// The method tracks request and response bytes, latency, and error conditions.
// It updates the provided statistics object with the results. For Start requests,
// it also increments the online counter; for Stop requests, the offline counter.
//
// Parameters:
//   - packet: Accounting-Request packet to send
//   - stats: Statistics object to update with results
//   - isStart: True if this is an Accounting-Start request
//   - isStop: True if this is an Accounting-Stop request
//
// Returns:
//   - *radius.Packet: Received response packet (nil on timeout/error)
//   - error: Error if the request fails or times out
func (c *Client) SendAcctRequest(packet *radius.Packet, stats *Statistics, isStart, isStop bool) (*radius.Packet, error) {
	stats.AddRequestBytes(int64(radiusd.Length(packet)))

	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	startTime := time.Now()
	resp, err := c.radiusClient.Exchange(ctx, packet, c.acctAddr)
	latency := time.Since(startTime).Milliseconds()

	if err != nil {
		if err == context.DeadlineExceeded {
			stats.IncrAcctTimeout()
		}
		stats.IncrAcctDrop()
		return nil, err
	}

	if resp != nil {
		stats.AddResponseBytes(int64(radiusd.Length(resp)))
		stats.RecordAcctLatency(latency)

		if resp.Code == radius.CodeAccountingResponse {
			stats.IncrAcctResponse()

			// Track session state changes
			if isStart {
				stats.IncrOnline()
			} else if isStop {
				stats.IncrOffline()
			}
		} else {
			stats.IncrAcctDrop()
			return resp, fmt.Errorf("unexpected response code: %v", resp.Code)
		}
	}

	return resp, nil
}

// Close releases any resources held by the client.
// Currently a no-op but provided for future extensibility.
func (c *Client) Close() error {
	return nil
}

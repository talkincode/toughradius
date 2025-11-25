package benchmark

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync/atomic"
	"time"

	"github.com/panjf2000/ants/v2"
	"github.com/talkincode/toughradius/v9/pkg/common"
	"layeh.com/radius/rfc2866"
)

// User represents a test user for benchmarking.
type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
	IP       string `json:"ipaddr"`
	MAC      string `json:"macaddr"`
}

// Task orchestrates the benchmark testing process.
// It manages worker pools, statistics collection, and progress reporting.
type Task struct {
	config        *Config
	client        *Client
	packetBuilder *PacketBuilder
	reporter      *Reporter
	stats         *Statistics
	workerPool    *ants.Pool
	users         []User
	counter       int64 // Total transactions completed
	ctx           context.Context
	cancel        context.CancelFunc
}

// NewTask creates a new benchmark task with the given configuration.
//
// Parameters:
//   - cfg: Benchmark configuration
//
// Returns:
//   - *Task: Configured task instance
//   - error: Error if initialization fails
func NewTask(cfg *Config) (*Task, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// Create worker pool
	pool, err := ants.NewPool(int(cfg.Load.Concurrency))
	if err != nil {
		return nil, fmt.Errorf("failed to create worker pool: %w", err)
	}

	// Create client
	client := NewClient(
		cfg.Server.Address,
		cfg.Server.AuthPort,
		cfg.Server.AcctPort,
		cfg.GetTimeout(),
	)

	// Create packet builder
	packetBuilder := NewPacketBuilder(
		cfg.Server.Secret,
		cfg.NAS.Identifier,
		cfg.NAS.IP,
	)

	// Create reporter
	reporter, err := NewReporter(cfg.Output.CSVFile)
	if err != nil {
		pool.Release()
		return nil, fmt.Errorf("failed to create reporter: %w", err)
	}

	// Load test users
	users, err := loadUsers(cfg)
	if err != nil {
		pool.Release()
		_ = reporter.Close() //nolint:errcheck
		return nil, fmt.Errorf("failed to load users: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Task{
		config:        cfg,
		client:        client,
		packetBuilder: packetBuilder,
		reporter:      reporter,
		stats:         NewStatistics(),
		workerPool:    pool,
		users:         users,
		counter:       0,
		ctx:           ctx,
		cancel:        cancel,
	}, nil
}

// Run executes the benchmark test.
//
// This method:
//  1. Starts worker goroutines to submit tasks
//  2. Starts a monitoring goroutine to report progress
//  3. Waits for all tasks to complete
//  4. Reports final statistics
//
// Returns:
//   - error: Error if the test fails
func (t *Task) Run() error {
	defer t.cleanup()

	fmt.Printf("\n========== Starting RADIUS Benchmark ==========\n")
	fmt.Printf("Server: %s:%d (Auth) / %s:%d (Acct)\n",
		t.config.Server.Address, t.config.Server.AuthPort,
		t.config.Server.Address, t.config.Server.AcctPort)
	fmt.Printf("Concurrency: %d | Total Transactions: %d\n",
		t.config.Load.Concurrency, t.config.Load.Total)
	fmt.Printf("Test Users: %d | Timeout: %ds | Interval: %ds\n",
		len(t.users), t.config.Server.Timeout, t.config.Load.Interval)
	fmt.Printf("================================================\n\n")

	// Start progress monitoring
	go t.monitorProgress()

	// Submit tasks to worker pool
	go t.submitTasks()

	// Wait for completion
	t.waitForCompletion()

	// Print final statistics
	fmt.Printf("\n========== Benchmark Completed ==========\n")
	t.reporter.PrintProgress(t.stats, atomic.LoadInt64(&t.counter))

	return nil
}

// submitTasks continuously submits benchmark tasks to the worker pool.
func (t *Task) submitTasks() {
	userIndex := 0

	for {
		// Check if we've reached the target
		if atomic.LoadInt64(&t.counter) >= t.config.Load.Total {
			return
		}

		// Check if context is canceled
		select {
		case <-t.ctx.Done():
			return
		default:
		}

		// Get next user (round-robin)
		user := t.users[userIndex]
		userIndex = (userIndex + 1) % len(t.users)

		// Submit task to worker pool
		err := t.workerPool.Submit(func() {
			t.executeTransaction(&user)
		})

		if err != nil {
			// Pool might be closed
			return
		}
	}
}

// executeTransaction performs a complete RADIUS authentication and accounting cycle for a user.
//
// The transaction includes:
//  1. Access-Request (authentication)
//  2. Accounting-Start
//  3. Accounting-Interim-Update
//  4. Accounting-Stop
//
// Parameters:
//   - user: Test user to authenticate
func (t *Task) executeTransaction(user *User) {
	// Check if we've reached the limit
	current := atomic.AddInt64(&t.counter, 1)
	if current > t.config.Load.Total {
		return
	}

	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "Transaction panic: %v\n", r)
		}
	}()

	// Step 1: Authentication
	authPacket, err := t.packetBuilder.BuildAuthRequest(user.Username, user.Password, user.MAC)
	if err != nil {
		t.stats.IncrAuthDrop()
		return
	}

	authResp, err := t.client.SendAuthRequest(authPacket, t.stats)
	if err != nil || authResp == nil {
		return // Stats already updated by client
	}

	// Only continue accounting if authentication succeeded
	if authResp.Code != 1 { // radius.CodeAccessAccept
		return
	}

	// Generate session ID for accounting
	sessionID := common.UUID()

	// Step 2: Accounting-Start
	t.stats.IncrAcctStart()
	startPacket, err := t.packetBuilder.BuildAcctRequest(
		user.Username,
		user.IP,
		user.MAC,
		sessionID,
		rfc2866.AcctStatusType_Value_Start,
	)
	if err != nil {
		t.stats.IncrAcctDrop()
		return
	}

	_, err = t.client.SendAcctRequest(startPacket, t.stats, true, false)
	if err != nil {
		return // Stats already updated
	}

	// Step 3: Accounting-Interim-Update
	t.stats.IncrAcctUpdate()
	updatePacket, err := t.packetBuilder.BuildAcctRequest(
		user.Username,
		user.IP,
		user.MAC,
		sessionID,
		rfc2866.AcctStatusType_Value_InterimUpdate,
	)
	if err != nil {
		t.stats.IncrAcctDrop()
		return
	}

	_, err = t.client.SendAcctRequest(updatePacket, t.stats, false, false)
	if err != nil {
		return
	}

	// Step 4: Accounting-Stop
	t.stats.IncrAcctStop()
	stopPacket, err := t.packetBuilder.BuildAcctRequest(
		user.Username,
		user.IP,
		user.MAC,
		sessionID,
		rfc2866.AcctStatusType_Value_Stop,
	)
	if err != nil {
		t.stats.IncrAcctDrop()
		return
	}

	_, err = t.client.SendAcctRequest(stopPacket, t.stats, false, true)
	if err != nil {
		return
	}
}

// monitorProgress periodically reports benchmark progress.
func (t *Task) monitorProgress() {
	ticker := time.NewTicker(time.Duration(t.config.Load.Interval) * time.Second)
	defer ticker.Stop()

	prevSnapshot := t.stats.Snapshot()

	for {
		select {
		case <-t.ctx.Done():
			return
		case <-ticker.C:
			currentSnapshot := t.stats.Snapshot()
			duration := currentSnapshot.LastUpdate.Sub(prevSnapshot.LastUpdate)

			// Update performance metrics
			currentSnapshot.UpdatePerformanceMetrics(prevSnapshot, duration)

			// Print to console
			t.reporter.PrintProgress(currentSnapshot, atomic.LoadInt64(&t.counter))

			// Write to CSV
			if err := t.reporter.WriteCSVRow(currentSnapshot); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to write CSV: %v\n", err)
			}

			prevSnapshot = currentSnapshot
		}
	}
}

// waitForCompletion waits for all tasks to finish.
func (t *Task) waitForCompletion() {
	// Poll until target is reached
	for {
		if atomic.LoadInt64(&t.counter) >= t.config.Load.Total {
			// Wait for worker pool to drain
			for t.workerPool.Running() > 0 {
				time.Sleep(100 * time.Millisecond)
			}
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
}

// cleanup releases resources.
func (t *Task) cleanup() {
	t.cancel()
	if t.workerPool != nil {
		t.workerPool.Release()
	}
	if t.client != nil {
		_ = t.client.Close() //nolint:errcheck
	}
	if t.reporter != nil {
		_ = t.reporter.Close() //nolint:errcheck
	}
}

// loadUsers loads test users from data file or creates a single user from config.
func loadUsers(cfg *Config) ([]User, error) {
	users := make([]User, 0)

	// Try to load from file first
	if cfg.User.DataFile != "" && fileExists(cfg.User.DataFile) {
		data, err := os.ReadFile(cfg.User.DataFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read user data file: %w", err)
		}

		if err := json.Unmarshal(data, &users); err != nil {
			return nil, fmt.Errorf("failed to parse user data file: %w", err)
		}

		if len(users) > 0 {
			return users, nil
		}
	}

	// Fallback to single user from config
	users = append(users, User{
		Username: cfg.User.Username,
		Password: cfg.User.Password,
		IP:       cfg.User.IP,
		MAC:      cfg.User.MAC,
	})

	return users, nil
}

// fileExists checks if a file exists.
func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

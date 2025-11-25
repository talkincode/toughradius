package benchmark

import (
	"bytes"
	"encoding/csv"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

func TestReporterCSVOutput(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "report.csv")

	reporter, err := NewReporter(path)
	if err != nil {
		t.Fatalf("NewReporter returned error: %v", err)
	}
	defer reporter.Close()

	stats := &Statistics{}
	stats.Auth.Requests = 10
	stats.Auth.Accepts = 7
	stats.Auth.Rejects = 2
	stats.Auth.Drops = 1
	stats.Auth.Timeouts = 0
	stats.Auth.LatencyUnder10ms = 5
	stats.Auth.Latency10to100ms = 3
	stats.Auth.Latency100to1000ms = 1
	stats.Auth.LatencyOver1000ms = 1

	stats.Acct.StartRequests = 4
	stats.Acct.StopRequests = 3
	stats.Acct.UpdateRequests = 2
	stats.Acct.Responses = 5
	stats.Acct.Drops = 1
	stats.Acct.Timeouts = 1
	stats.Acct.OnlineCount = 6
	stats.Acct.OfflineCount = 2
	stats.Acct.LatencyUnder10ms = 2
	stats.Acct.Latency10to100ms = 1
	stats.Acct.Latency100to1000ms = 1
	stats.Acct.LatencyOver1000ms = 1

	stats.Network.RequestBytes = 1024
	stats.Network.ResponseBytes = 2048

	stats.Performance.AuthQPS = 20
	stats.Performance.AcctQPS = 10
	stats.Performance.TotalQPS = 30
	stats.Performance.OnlineRate = 3
	stats.Performance.OfflineRate = 1
	stats.Performance.UploadRate = 4096
	stats.Performance.DownloadRate = 2048

	if err := reporter.WriteCSVRow(stats); err != nil {
		t.Fatalf("WriteCSVRow returned error: %v", err)
	}

	// Close to flush underlying file
	if err := reporter.Close(); err != nil {
		t.Fatalf("Close returned error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read csv: %v", err)
	}

	records, err := csv.NewReader(bytes.NewReader(data)).ReadAll()
	if err != nil {
		t.Fatalf("failed to parse csv: %v", err)
	}

	if len(records) != 2 {
		t.Fatalf("expected header and one data row, got %d", len(records))
	}

	header := records[0]
	row := records[1]

	if len(header) != len(row) {
		t.Fatalf("header and row length mismatch: %d vs %d", len(header), len(row))
	}

	if header[0] != "Timestamp" {
		t.Fatalf("unexpected header[0]: %s", header[0])
	}

	if row[1] != "10" || row[2] != "7" {
		t.Fatalf("unexpected auth counts row[1]=%s row[2]=%s", row[1], row[2])
	}

	// Column 13 (index 12) is OnlineCount, ensure it matches
	onlineIdx := 12
	online, err := strconv.Atoi(row[onlineIdx])
	if err != nil || online != 6 {
		t.Fatalf("expected OnlineCount 6, got %s (err=%v)", row[onlineIdx], err)
	}
}

func TestReporterWriteCSVRowNoFile(t *testing.T) {
	reporter, err := NewReporter("")
	if err != nil {
		t.Fatalf("NewReporter returned error: %v", err)
	}

	stats := &Statistics{}
	stats.StartTime = time.Now()
	if err := reporter.WriteCSVRow(stats); err != nil {
		t.Fatalf("WriteCSVRow should be no-op, got %v", err)
	}

	if err := reporter.Close(); err != nil {
		t.Fatalf("Close returned error: %v", err)
	}
}

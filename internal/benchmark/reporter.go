package benchmark

import (
	"encoding/csv"
	"fmt"
	"os"
	"time"

	"github.com/talkincode/toughradius/v9/pkg/timeutil"
)

// Reporter handles formatting and outputting benchmark results.
// It supports both console and CSV output formats.
type Reporter struct {
	csvWriter *csv.Writer
	csvFile   *os.File
}

// NewReporter creates a new Reporter instance.
//
// Parameters:
//   - csvPath: Path to CSV output file (empty string to disable CSV output)
//
// Returns:
//   - *Reporter: Configured reporter instance
//   - error: Error if CSV file cannot be created
func NewReporter(csvPath string) (*Reporter, error) {
	r := &Reporter{}

	if csvPath != "" {
		file, err := os.Create(csvPath) //nolint:gosec // G304: path is user-specified output file
		if err != nil {
			return nil, fmt.Errorf("failed to create CSV file: %w", err)
		}
		r.csvFile = file
		r.csvWriter = csv.NewWriter(file)

		// Write CSV header
		header := []string{
			"Timestamp",
			"AuthReq", "AuthAccept", "AuthReject", "AuthDrop", "AuthTimeout",
			"AcctStart", "AcctStop", "AcctUpdate", "AcctResp", "AcctDrop", "AcctTimeout",
			"OnlineCount", "OfflineCount",
			"ReqBytes", "RespBytes",
			"Auth0-10ms", "Auth10-100ms", "Auth100-1000ms", "Auth>1000ms",
			"Acct0-10ms", "Acct10-100ms", "Acct100-1000ms", "Acct>1000ms",
			"AuthQPS", "AcctQPS", "TotalQPS",
			"OnlineRate", "OfflineRate",
			"UploadRate", "DownloadRate",
		}
		if err := r.csvWriter.Write(header); err != nil {
			_ = file.Close() //nolint:errcheck
			return nil, fmt.Errorf("failed to write CSV header: %w", err)
		}
		r.csvWriter.Flush()
	}

	return r, nil
}

// PrintProgress prints current statistics to the console in a human-readable format.
//
// The output includes:
//   - Transaction count and time elapsed
//   - Peak and realtime QPS metrics
//   - Total counters for auth/acct requests
//   - Latency distribution
//   - Network bandwidth usage
//
// Parameters:
//   - stats: Statistics snapshot to display
//   - totalTrans: Total number of transactions completed
func (r *Reporter) PrintProgress(stats *Statistics, totalTrans int64) {
	elapsed := time.Since(stats.StartTime).Seconds()

	output := fmt.Sprintf(
		"----------- RADIUS Benchmark Statistics: Total Transactions: %d ----------------------\n"+
			"-- Start Time: %s | Last Update: %s | Elapsed: %.2f seconds\n"+
			"-- Max QPS: Auth: %d | Acct: %d | Total: %d | Online: %d | Offline: %d\n"+
			"-- Realtime QPS: Auth: %d | Acct: %d | Total: %d | Online: %d | Offline: %d\n"+
			"-- Auth Total: Requests: %d | Accepts: %d | Rejects: %d | Drops: %d | Timeouts: %d\n"+
			"-- Acct Total: Start: %d | Stop: %d | Update: %d | Responses: %d | Drops: %d | Timeouts: %d\n"+
			"-- Network: Request KB: %d | Response KB: %d | Upload: %d KB/s | Download: %d KB/s\n"+
			"-- Auth Latency: 0-10ms: %d | 10-100ms: %d | 100-1000ms: %d | >1000ms: %d\n"+
			"-- Acct Latency: 0-10ms: %d | 10-100ms: %d | 100-1000ms: %d | >1000ms: %d\n"+
			"-------------------------------------------------------------------------------------------\n",
		totalTrans,
		stats.StartTime.Format(timeutil.YYYYMMDDHHMMSS_LAYOUT),
		stats.LastUpdate.Format(timeutil.YYYYMMDDHHMMSS_LAYOUT),
		elapsed,
		stats.Performance.MaxAuthQPS,
		stats.Performance.MaxAcctQPS,
		stats.Performance.MaxTotalQPS,
		stats.Performance.MaxOnlineRate,
		stats.Performance.MaxOfflineRate,
		stats.Performance.AuthQPS,
		stats.Performance.AcctQPS,
		stats.Performance.TotalQPS,
		stats.Performance.OnlineRate,
		stats.Performance.OfflineRate,
		stats.Auth.Requests,
		stats.Auth.Accepts,
		stats.Auth.Rejects,
		stats.Auth.Drops,
		stats.Auth.Timeouts,
		stats.Acct.StartRequests,
		stats.Acct.StopRequests,
		stats.Acct.UpdateRequests,
		stats.Acct.Responses,
		stats.Acct.Drops,
		stats.Acct.Timeouts,
		stats.Network.RequestBytes/1024,
		stats.Network.ResponseBytes/1024,
		stats.Performance.UploadRate/1024,
		stats.Performance.DownloadRate/1024,
		stats.Auth.LatencyUnder10ms,
		stats.Auth.Latency10to100ms,
		stats.Auth.Latency100to1000ms,
		stats.Auth.LatencyOver1000ms,
		stats.Acct.LatencyUnder10ms,
		stats.Acct.Latency10to100ms,
		stats.Acct.Latency100to1000ms,
		stats.Acct.LatencyOver1000ms,
	)

	fmt.Print(output)
}

// WriteCSVRow writes a statistics snapshot to the CSV file.
//
// Parameters:
//   - stats: Statistics snapshot to write
//
// Returns:
//   - error: Error if write or flush fails
func (r *Reporter) WriteCSVRow(stats *Statistics) error {
	if r.csvWriter == nil {
		return nil // CSV output disabled
	}

	row := []string{
		fmt.Sprintf("%d", time.Now().Unix()),
		fmt.Sprintf("%d", stats.Auth.Requests),
		fmt.Sprintf("%d", stats.Auth.Accepts),
		fmt.Sprintf("%d", stats.Auth.Rejects),
		fmt.Sprintf("%d", stats.Auth.Drops),
		fmt.Sprintf("%d", stats.Auth.Timeouts),
		fmt.Sprintf("%d", stats.Acct.StartRequests),
		fmt.Sprintf("%d", stats.Acct.StopRequests),
		fmt.Sprintf("%d", stats.Acct.UpdateRequests),
		fmt.Sprintf("%d", stats.Acct.Responses),
		fmt.Sprintf("%d", stats.Acct.Drops),
		fmt.Sprintf("%d", stats.Acct.Timeouts),
		fmt.Sprintf("%d", stats.Acct.OnlineCount),
		fmt.Sprintf("%d", stats.Acct.OfflineCount),
		fmt.Sprintf("%d", stats.Network.RequestBytes),
		fmt.Sprintf("%d", stats.Network.ResponseBytes),
		fmt.Sprintf("%d", stats.Auth.LatencyUnder10ms),
		fmt.Sprintf("%d", stats.Auth.Latency10to100ms),
		fmt.Sprintf("%d", stats.Auth.Latency100to1000ms),
		fmt.Sprintf("%d", stats.Auth.LatencyOver1000ms),
		fmt.Sprintf("%d", stats.Acct.LatencyUnder10ms),
		fmt.Sprintf("%d", stats.Acct.Latency10to100ms),
		fmt.Sprintf("%d", stats.Acct.Latency100to1000ms),
		fmt.Sprintf("%d", stats.Acct.LatencyOver1000ms),
		fmt.Sprintf("%d", stats.Performance.AuthQPS),
		fmt.Sprintf("%d", stats.Performance.AcctQPS),
		fmt.Sprintf("%d", stats.Performance.TotalQPS),
		fmt.Sprintf("%d", stats.Performance.OnlineRate),
		fmt.Sprintf("%d", stats.Performance.OfflineRate),
		fmt.Sprintf("%d", stats.Performance.UploadRate),
		fmt.Sprintf("%d", stats.Performance.DownloadRate),
	}

	if err := r.csvWriter.Write(row); err != nil {
		return fmt.Errorf("failed to write CSV row: %w", err)
	}

	r.csvWriter.Flush()
	return r.csvWriter.Error()
}

// Close flushes and closes the CSV file if opened.
func (r *Reporter) Close() error {
	if r.csvWriter != nil {
		r.csvWriter.Flush()
		if err := r.csvWriter.Error(); err != nil {
			return err
		}
	}
	if r.csvFile != nil {
		return r.csvFile.Close()
	}
	return nil
}

package benchmark

import (
	"testing"
	"time"
)

func TestStatisticsLatencyBuckets(t *testing.T) {
	stats := NewStatistics()

	stats.RecordAuthLatency(5)
	stats.RecordAuthLatency(50)
	stats.RecordAuthLatency(500)
	stats.RecordAuthLatency(1500)

	if stats.Auth.LatencyUnder10ms != 1 || stats.Auth.Latency10to100ms != 1 || stats.Auth.Latency100to1000ms != 1 || stats.Auth.LatencyOver1000ms != 1 {
		t.Fatalf("unexpected auth latency buckets: %+v", stats.Auth)
	}

	stats.RecordAcctLatency(5)
	stats.RecordAcctLatency(50)
	stats.RecordAcctLatency(500)
	stats.RecordAcctLatency(1500)

	if stats.Acct.LatencyUnder10ms != 1 || stats.Acct.Latency10to100ms != 1 || stats.Acct.Latency100to1000ms != 1 || stats.Acct.LatencyOver1000ms != 1 {
		t.Fatalf("unexpected acct latency buckets: %+v", stats.Acct)
	}
}

func TestStatisticsUpdatePerformanceMetrics(t *testing.T) {
	prev := NewStatistics()
	curr := NewStatistics()

	prev.Auth.Accepts = 10
	prev.Auth.Rejects = 5
	curr.Auth.Accepts = 25
	curr.Auth.Rejects = 15

	prev.Acct.Responses = 4
	curr.Acct.Responses = 14

	prev.Acct.OnlineCount = 2
	curr.Acct.OnlineCount = 12
	prev.Acct.OfflineCount = 1
	curr.Acct.OfflineCount = 6

	prev.Network.RequestBytes = 1000
	curr.Network.RequestBytes = 6000
	prev.Network.ResponseBytes = 500
	curr.Network.ResponseBytes = 2500

	duration := 5 * time.Second
	curr.UpdatePerformanceMetrics(prev, duration)

	if curr.Performance.AuthQPS != 5 { // ((25+15)-(10+5)) / 5 = 25/5
		t.Fatalf("unexpected auth qps: %+v", curr.Performance)
	}
	if curr.Performance.AcctQPS != 2 { // (14-4)/5
		t.Fatalf("unexpected acct qps: %+v", curr.Performance)
	}
	if curr.Performance.TotalQPS != 7 {
		t.Fatalf("unexpected total qps: %+v", curr.Performance)
	}
	if curr.Performance.OnlineRate != 2 { // (12-2)/5
		t.Fatalf("unexpected online rate: %+v", curr.Performance)
	}
	if curr.Performance.OfflineRate != 1 { // (6-1)/5
		t.Fatalf("unexpected offline rate: %+v", curr.Performance)
	}
	if curr.Performance.UploadRate != 1000 { // (6000-1000)/5
		t.Fatalf("unexpected upload rate: %+v", curr.Performance)
	}
	if curr.Performance.DownloadRate != 400 { // (2500-500)/5
		t.Fatalf("unexpected download rate: %+v", curr.Performance)
	}

	if curr.Performance.MaxTotalQPS != curr.Performance.TotalQPS {
		t.Fatalf("expected peak totals to match current values")
	}
	if curr.LastUpdate.Before(prev.LastUpdate) {
		t.Fatalf("expected LastUpdate to advance")
	}
}

func TestStatisticsSnapshotIsolation(t *testing.T) {
	stats := NewStatistics()
	stats.IncrAuthRequest()
	stats.IncrAuthAccept()
	stats.IncrAcctStart()
	stats.AddRequestBytes(128)

	snap := stats.Snapshot()

	stats.IncrAuthRequest()
	stats.IncrAuthReject()
	stats.IncrAcctStart()
	stats.AddRequestBytes(256)

	if snap.Auth.Requests != 1 || snap.Auth.Accepts != 1 {
		t.Fatalf("snapshot auth counters incorrect: %+v", snap.Auth)
	}
	if snap.Acct.StartRequests != 1 {
		t.Fatalf("snapshot acct counters incorrect: %+v", snap.Acct)
	}
	if snap.Network.RequestBytes != 128 {
		t.Fatalf("snapshot network counters incorrect: %+v", snap.Network)
	}
}

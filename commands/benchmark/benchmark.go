package main

import (
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/talkincode/toughradius/v8/common/timeutil"
)

type BenchmarkStat struct {
	AuthReq         int64
	AuthAccept      int64
	AuthReject      int64
	AuthRejectdelay int64
	AuthDrop        int64
	AcctStart       int64
	AcctStop        int64
	AcctUpdate      int64
	OnlineStart     int64
	OnlineStop      int64
	AcctResp        int64
	AcctDrop        int64
	ReqBytes        int64
	RespBytes       int64
	AuthTimeout     int64
	AcctTimeout     int64
	AuthMs10        int64
	AuthMs100       int64
	AuthMs1000      int64
	AuthMsGe1000    int64
	AcctMs10        int64
	AcctMs100       int64
	AcctMs1000      int64
	AcctMsGe1000    int64
	// stat
	AuthQps        int64
	AcctQps        int64
	AllQps         int64
	OnlineRate     int64
	OfflineRate    int64
	UpRate         int64
	DownRate       int64
	MaxAuthQps     int64
	MaxAcctQps     int64
	MaxAllQps      int64
	MaxOnlineRate  int64
	MaxOfflineRate int64
	MaxUpRate      int64
	MaxDownRate    int64
	StatStart      int64
	StatEnd        int64
}

type BenchmarkTask struct {
	currStat   *BenchmarkStat
	TotalStat  *BenchmarkStat
	StartTime  time.Time
	LastUpdate time.Time
	Counter    int64
	Lock       sync.Mutex
}

func NewBenchmarkTask() *BenchmarkTask {
	return &BenchmarkTask{
		currStat:   &BenchmarkStat{},
		TotalStat:  &BenchmarkStat{},
		StartTime:  time.Now(),
		LastUpdate: time.Now(),
		Counter:    0,
		Lock:       sync.Mutex{},
	}
}

func (task *BenchmarkTask) IncrAuthCast(cast int64) {
	if cast <= 10 {
		atomic.AddInt64(&task.TotalStat.AuthMs10, 1)
	} else if cast > 10 && cast <= 100 {
		atomic.AddInt64(&task.TotalStat.AuthMs100, 1)
	} else if cast > 100 && cast <= 1000 {
		atomic.AddInt64(&task.TotalStat.AuthMs1000, 1)
	} else {
		atomic.AddInt64(&task.TotalStat.AuthMsGe1000, 1)
	}
}

func (task *BenchmarkTask) IncrAcctCast(cast int64) {
	if cast <= 10 {
		atomic.AddInt64(&task.TotalStat.AcctMs10, 1)
	} else if cast > 10 && cast <= 100 {
		atomic.AddInt64(&task.TotalStat.AcctMs100, 1)
	} else if cast > 100 && cast <= 1000 {
		atomic.AddInt64(&task.TotalStat.AcctMs1000, 1)
	} else {
		atomic.AddInt64(&task.TotalStat.AcctMsGe1000, 1)
	}
}

func (task *BenchmarkTask) GetCurrStat() *BenchmarkStat {
	task.Lock.Lock()
	var t = task.currStat
	task.Lock.Unlock()
	return t
}

func (task *BenchmarkTask) IncrReqBytes(b int64) {
	task.Lock.Lock()
	atomic.AddInt64(&task.currStat.ReqBytes, b)
	task.Lock.Unlock()
}

func (task *BenchmarkTask) IncrRespBytes(b int64) {
	task.Lock.Lock()
	atomic.AddInt64(&task.currStat.RespBytes, b)
	task.Lock.Unlock()
}

func (task *BenchmarkTask) IncrCounter(name string) {
	task.Lock.Lock()
	switch name {
	case "Counter":
		atomic.AddInt64(&task.Counter, 1)
	case "AuthReq":
		atomic.AddInt64(&task.currStat.AuthReq, 1)
	case "AuthAccept":
		atomic.AddInt64(&task.currStat.AuthAccept, 1)
	case "AuthReject":
		atomic.AddInt64(&task.currStat.AuthReject, 1)
	case "AuthDrop":
		atomic.AddInt64(&task.currStat.AuthDrop, 1)
	case "AcctStart":
		atomic.AddInt64(&task.currStat.AcctStart, 1)
	case "AcctStop":
		atomic.AddInt64(&task.currStat.AcctStop, 1)
	case "AcctUpdate":
		atomic.AddInt64(&task.currStat.AcctUpdate, 1)
	case "AcctResp":
		atomic.AddInt64(&task.currStat.AcctResp, 1)
	case "AcctDrop":
		atomic.AddInt64(&task.currStat.AcctDrop, 1)
	case "AuthTimeout":
		atomic.AddInt64(&task.currStat.AuthTimeout, 1)
	case "AuthMs10":
		atomic.AddInt64(&task.currStat.AuthMs10, 1)
	case "AuthMs100":
		atomic.AddInt64(&task.currStat.AuthMs100, 1)
	case "AuthMs1000":
		atomic.AddInt64(&task.currStat.AuthMs1000, 1)
	case "AuthMsGe1000":
		atomic.AddInt64(&task.currStat.AuthMsGe1000, 1)
	case "AcctMs10":
		atomic.AddInt64(&task.currStat.AcctMs10, 1)
	case "AcctMs100":
		atomic.AddInt64(&task.currStat.AcctMs100, 1)
	case "AcctMs1000":
		atomic.AddInt64(&task.currStat.AcctMs1000, 1)
	case "AcctMsGe1000":
		atomic.AddInt64(&task.currStat.AcctMsGe1000, 1)
	case "OnlineStart":
		atomic.AddInt64(&task.currStat.OnlineStart, 1)
	case "OnlineStop":
		atomic.AddInt64(&task.currStat.OnlineStop, 1)
	}
	task.Lock.Unlock()
}

func comporeMaxValue(v1, v2 int64) int64 {
	if v2 > v1 {
		return v2
	}
	return v1
}

func (task *BenchmarkTask) UpdateStat() {
	task.Lock.Lock()
	tmpstat := *task.currStat
	task.currStat = &BenchmarkStat{}
	var cast = time.Since(task.LastUpdate)
	task.LastUpdate = time.Now()
	atomic.AddInt64(&task.TotalStat.AuthReq, tmpstat.AuthReq)
	atomic.AddInt64(&task.TotalStat.AuthAccept, tmpstat.AuthAccept)
	atomic.AddInt64(&task.TotalStat.AuthReject, tmpstat.AuthReject)
	atomic.AddInt64(&task.TotalStat.AuthDrop, tmpstat.AuthDrop)
	atomic.AddInt64(&task.TotalStat.AcctStart, tmpstat.AcctStart)
	atomic.AddInt64(&task.TotalStat.AcctStop, tmpstat.AcctStop)
	atomic.AddInt64(&task.TotalStat.AcctUpdate, tmpstat.AcctUpdate)
	atomic.AddInt64(&task.TotalStat.AcctResp, tmpstat.AcctResp)
	atomic.AddInt64(&task.TotalStat.AcctDrop, tmpstat.AcctDrop)
	atomic.AddInt64(&task.TotalStat.ReqBytes, tmpstat.ReqBytes)
	atomic.AddInt64(&task.TotalStat.RespBytes, tmpstat.RespBytes)
	atomic.AddInt64(&task.TotalStat.AuthTimeout, tmpstat.AuthTimeout)
	atomic.AddInt64(&task.TotalStat.AcctTimeout, tmpstat.AcctTimeout)
	atomic.AddInt64(&task.TotalStat.AuthMs10, tmpstat.AuthMs10)
	atomic.AddInt64(&task.TotalStat.AuthMs100, tmpstat.AuthMs100)
	atomic.AddInt64(&task.TotalStat.AuthMs1000, tmpstat.AuthMs1000)
	atomic.AddInt64(&task.TotalStat.AuthMsGe1000, tmpstat.AuthMsGe1000)
	atomic.AddInt64(&task.TotalStat.AcctMs10, tmpstat.AcctMs10)
	atomic.AddInt64(&task.TotalStat.AcctMs100, tmpstat.AcctMs100)
	atomic.AddInt64(&task.TotalStat.AcctMs1000, tmpstat.AcctMs1000)
	atomic.AddInt64(&task.TotalStat.AcctMsGe1000, tmpstat.AcctMsGe1000)
	atomic.AddInt64(&task.TotalStat.OnlineStart, tmpstat.OnlineStart)
	atomic.AddInt64(&task.TotalStat.OnlineStop, tmpstat.OnlineStop)
	task.TotalStat.AuthQps = int64(float64(tmpstat.AuthAccept+tmpstat.AuthReject) / cast.Seconds())
	task.TotalStat.AcctQps = int64(float64(tmpstat.AcctResp) / cast.Seconds())
	task.TotalStat.AllQps = int64(float64(tmpstat.AuthAccept+tmpstat.AuthReject+tmpstat.AcctResp) / cast.Seconds())
	task.TotalStat.OnlineRate = int64(float64(tmpstat.OnlineStart) / cast.Seconds())
	task.TotalStat.OfflineRate = int64(float64(tmpstat.OnlineStop) / cast.Seconds())
	task.TotalStat.UpRate = int64(float64(tmpstat.ReqBytes) / cast.Seconds())
	task.TotalStat.DownRate = int64(float64(tmpstat.RespBytes) / cast.Seconds())
	task.TotalStat.MaxAuthQps = comporeMaxValue(task.TotalStat.MaxAuthQps, task.TotalStat.AuthQps)
	task.TotalStat.MaxAcctQps = comporeMaxValue(task.TotalStat.MaxAcctQps, task.TotalStat.AcctQps)
	task.TotalStat.MaxAllQps = comporeMaxValue(task.TotalStat.MaxAllQps, task.TotalStat.AllQps)
	task.TotalStat.MaxOnlineRate = comporeMaxValue(task.TotalStat.MaxOnlineRate, task.TotalStat.OnlineRate)
	task.TotalStat.MaxOfflineRate = comporeMaxValue(task.TotalStat.MaxOfflineRate, task.TotalStat.OfflineRate)
	task.TotalStat.MaxUpRate = comporeMaxValue(task.TotalStat.MaxUpRate, task.TotalStat.UpRate)
	task.TotalStat.MaxDownRate = comporeMaxValue(task.TotalStat.MaxDownRate, task.TotalStat.DownRate)
	task.Lock.Unlock()
}

func (task *BenchmarkTask) GetLineStatHeader() string {
	return "AuthReq,AuthAccept,AuthReject,AuthDrop,AcctStart,AcctStop,AcctUpdate,AcctResp,AcctDrop,AuthTimeout,AcctTimeout," +
		"ReqBytes,RespBytes,AuthMs10,AuthMs100,AuthMs1000,AuthMsGe1000,AcctMs10,AcctMs100,AcctMs1000,AcctMsGe1000,AuthQps,AcctQps,AllQps,OnlineRate,OfflineRate,Time"
}

func (task *BenchmarkTask) GetCurrentStatLine() string {
	return fmt.Sprintf("%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d,%d",
		task.TotalStat.AuthReq,
		task.TotalStat.AuthAccept,
		task.TotalStat.AuthReject,
		task.TotalStat.AuthDrop,
		task.TotalStat.AcctStart,
		task.TotalStat.AcctStop,
		task.TotalStat.AcctUpdate,
		task.TotalStat.AcctResp,
		task.TotalStat.AcctDrop,
		task.TotalStat.AuthTimeout,
		task.TotalStat.AcctTimeout,
		task.TotalStat.ReqBytes,
		task.TotalStat.RespBytes,
		task.TotalStat.AuthMs10,
		task.TotalStat.AuthMs100,
		task.TotalStat.AuthMs1000,
		task.TotalStat.AuthMsGe1000,
		task.TotalStat.AcctMs10,
		task.TotalStat.AcctMs100,
		task.TotalStat.AcctMs1000,
		task.TotalStat.AcctMsGe1000,
		task.TotalStat.AuthQps,
		task.TotalStat.AcctQps,
		task.TotalStat.AllQps,
		task.TotalStat.OnlineRate,
		task.TotalStat.OfflineRate,
		time.Now().Unix(),
	)
}

func (task *BenchmarkTask) GetTotalStatValues() []interface{} {
	return []interface{}{
		task.TotalStat.AuthReq,
		task.TotalStat.AuthAccept,
		task.TotalStat.AuthReject,
		task.TotalStat.AuthRejectdelay,
		task.TotalStat.AuthDrop,
		task.TotalStat.AcctStart,
		task.TotalStat.AcctStop,
		task.TotalStat.AcctUpdate,
		task.TotalStat.AcctResp,
		task.TotalStat.AcctDrop,
		task.TotalStat.AuthTimeout,
		task.TotalStat.AcctTimeout,
		task.TotalStat.ReqBytes,
		task.TotalStat.RespBytes,
		task.TotalStat.AuthMs10,
		task.TotalStat.AuthMs100,
		task.TotalStat.AuthMs1000,
		task.TotalStat.AuthMsGe1000,
		task.TotalStat.AcctMs10,
		task.TotalStat.AcctMs100,
		task.TotalStat.AcctMs1000,
		task.TotalStat.AcctMsGe1000,
		task.TotalStat.AuthQps,
		task.TotalStat.AcctQps,
		task.TotalStat.AllQps,
		task.TotalStat.OnlineRate,
		task.TotalStat.OfflineRate,
		task.TotalStat.MaxAuthQps,
		task.TotalStat.MaxAcctQps,
		task.TotalStat.MaxAllQps,
		task.TotalStat.MaxOnlineRate,
		task.TotalStat.MaxOfflineRate,
	}
}

func (task *BenchmarkTask) PrintCurrentStat() {
	strs := fmt.Sprintf(
		"----------- Radius Benchmark Statistic:  Total Transaction Num: %d -----------------------\n"+
			"-- StartTime: %s LastUpdate: %s, Time used %.2f Second\n"+
			"-- Max QPS: Auth: %d, Acct: %d, Both: %d, Online: %d, Offline: %d \n"+
			"-- Realtime QPS: Auth: %d, Acct: %d, Both: %d, Online: %d, Offline: %d \n"+
			"-- Total: AuthReq: %d, AuthAccept: %d, AuthReject: %d, AuthDrop: %d\n"+
			"-- Total: AcctStart: %d, AcctStop: %d, AcctUpdate: %d, AcctResp: %d, AcctDrop: %d\n"+
			"-- Total: AuthTimeout: %d, AcctTimeout: %d, Req KBytes: %d, Resp KBytes: %d\n"+
			"-- Auth Cast: 0-10 MS: %d, 10-100 MS: %d, 100-1000 MS: %d, > 1000 MS: %d\n"+
			"-- Acct Cast: 0-10 MS: %d, 10-100 MS: %d, 100-1000 MS: %d, > 1000 MS: %d\n"+
			"---------------------------------------------------------------------------------------------\n",
		task.Counter,
		task.StartTime.Format(timeutil.YYYYMMDDHHMMSS_LAYOUT),
		task.LastUpdate.Format(timeutil.YYYYMMDDHHMMSS_LAYOUT),
		task.LastUpdate.Sub(task.StartTime).Seconds(),
		task.TotalStat.MaxAuthQps,
		task.TotalStat.MaxAcctQps,
		task.TotalStat.MaxAllQps,
		task.TotalStat.MaxOnlineRate,
		task.TotalStat.MaxOfflineRate,
		task.TotalStat.AuthQps,
		task.TotalStat.AcctQps,
		task.TotalStat.AllQps,
		task.TotalStat.OnlineRate,
		task.TotalStat.OfflineRate,
		task.TotalStat.AuthReq,
		task.TotalStat.AuthAccept,
		task.TotalStat.AuthReject,
		task.TotalStat.AuthDrop,
		task.TotalStat.AcctStart,
		task.TotalStat.AcctStop,
		task.TotalStat.AcctUpdate,
		task.TotalStat.AcctResp,
		task.TotalStat.AcctDrop,
		task.TotalStat.AuthTimeout,
		task.TotalStat.AcctTimeout,
		task.TotalStat.ReqBytes/1024,
		task.TotalStat.RespBytes/1024,
		task.TotalStat.AuthMs10,
		task.TotalStat.AuthMs100,
		task.TotalStat.AuthMs1000,
		task.TotalStat.AuthMsGe1000,
		task.TotalStat.AcctMs10,
		task.TotalStat.AcctMs100,
		task.TotalStat.AcctMs1000,
		task.TotalStat.AcctMsGe1000,
	)
	fmt.Fprintln(os.Stdout, strs)
}

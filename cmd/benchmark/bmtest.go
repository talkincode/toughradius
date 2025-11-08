package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	_ "net/http/pprof"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/talkincode/toughradius/v9/internal/radiusd"
	"github.com/talkincode/toughradius/v9/pkg/common"
	"github.com/talkincode/toughradius/v9/pkg/timeutil"
	"layeh.com/radius"
	"layeh.com/radius/rfc2865"
	"layeh.com/radius/rfc2866"
	"layeh.com/radius/rfc2869"
)

// 命令行定义
var (
	// DEBUG 可打印详细日志，包括SQL
	h            = flag.Bool("h", false, "help usage")
	datafile     = flag.String("d", "bmdata.json", "test data file")
	user         = flag.String("u", "test01", "test username")
	passwd       = flag.String("p", "111111", "test password")
	secret       = flag.String("s", "secret", "radius secret")
	nasip        = flag.String("nasip", "127.0.0.1", "nas ip")
	userip       = flag.String("ip", "127.0.0.1", "user ip")
	usermac      = flag.String("mac", "11:11:11:11:11:11", "user mac")
	server       = flag.String("server", "127.0.0.1", "radius server")
	output       = flag.String("o", "tradtest.csv", "output csv result")
	encyrpt      = flag.String("e", "pap", "pap/chap")
	authport     = flag.Int("authport", 1812, "auth port")
	acctport     = flag.Int("acctport", 1813, "acct port")
	statInterval = flag.Int("i", 5, "stat interval seconds")
	total        = flag.Int64("n", 100, "total numbers")
	concurrency  = flag.Int64("c", 10, "concurrency")
	timeout      = flag.Int64("t", 10, "request timeout ")

	auth       = flag.Bool("auth", false, "send auth testing")
	acctstart  = flag.Bool("acct-start", false, "send acct start testing")
	acctupdate = flag.Bool("acct-update", false, "send acct update testing")
	acctstop   = flag.Bool("acct-stop", false, "send acct stop testing")
	benchmark  = flag.Bool("b", false, "benchmark test")
)

func printHelp() {
	if *h {
		ustr := fmt.Sprintf("toughradius benchmark tools, version: 1.0, Usage:tradtest -h\nOptions:")
		fmt.Fprintf(os.Stderr, ustr)
		flag.PrintDefaults()
		os.Exit(0)
	}
}

type BenchmarkTask struct {
	currStat   *BenchmarkStat
	TotalStat  *BenchmarkStat
	StartTime  time.Time
	LastUpdate time.Time
	Counter    int64
	Lock       sync.Mutex
}

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

type User struct {
	Username string
	Password string
	Ipaddr   string
	Macaddr  string
}

type GoLimit struct {
	ch chan int
}

func NewGoLimit(max int) *GoLimit {
	return &GoLimit{ch: make(chan int, max)}
}

func (g *GoLimit) Add() {
	g.ch <- 1
}

func (g *GoLimit) Done() {
	<-g.ch
}

func loadUserData() []User {
	users := make([]User, 0)
	if common.FileExists(*datafile) {
		data := common.Must2(ioutil.ReadFile(*datafile))
		common.Must(json.Unmarshal(data.([]byte), &users))
	}
	return users
}

func parseIp(ipstr string) net.IP {
	ip := net.ParseIP(ipstr)
	if ip == nil {
		return net.ParseIP("0.0.0.0")
	}
	return ip
}

func getAuthRequest(username, password, mac string) (*radius.Packet, error) {
	radreq := radius.New(radius.CodeAccessRequest, []byte(*secret))
	_ = rfc2865.UserName_SetString(radreq, username)
	if *encyrpt == "pap" {
		_ = rfc2865.UserPassword_SetString(radreq, password)
	} else {
		return nil, fmt.Errorf("not support")
	}
	rfc2865.NASIdentifier_Set(radreq, []byte("tradtest"))
	rfc2865.NASIPAddress_Set(radreq, parseIp(*nasip))
	rfc2865.NASPort_Set(radreq, 0)
	rfc2865.NASPortType_Set(radreq, 0)
	rfc2869.NASPortID_Set(radreq, []byte("slot=2;subslot=2;port=22;vlanid=100;"))
	rfc2865.CalledStationID_SetString(radreq, "11:11:11:11:11:11")
	rfc2865.CallingStationID_SetString(radreq, mac)
	return radreq, nil
}

func getAcctRequest(username, ipaddr, mac, sessionid string, accttype rfc2866.AcctStatusType) *radius.Packet {
	req := radius.New(radius.CodeAccountingRequest, []byte(*secret))
	rfc2865.UserName_SetString(req, username)
	rfc2865.NASIdentifier_Set(req, []byte("tradtest"))
	rfc2865.NASIPAddress_Set(req, parseIp(*nasip))
	rfc2865.NASPort_Set(req, 0)
	rfc2865.NASPortType_Set(req, 0)
	rfc2869.NASPortID_Set(req, []byte("slot=2;subslot=2;port=22;vlanid=100;"))
	rfc2865.CalledStationID_Set(req, []byte("11:11:11:11:11:11"))
	rfc2865.CallingStationID_Set(req, []byte(mac))
	rfc2866.AcctSessionID_SetString(req, sessionid)
	rfc2866.AcctInputOctets_Set(req, 0)
	rfc2866.AcctOutputOctets_Set(req, 0)
	rfc2866.AcctInputPackets_Set(req, 0)
	rfc2866.AcctOutputPackets_Set(req, 0)
	rfc2865.FramedIPAddress_Set(req, parseIp(ipaddr))
	rfc2866.AcctStatusType_Set(req, accttype)
	return req
}

func sendAuth() error {
	fmt.Fprintln(os.Stdout, fmt.Sprintf("Send AccessRequest to %s:%d secret=%s user=%s passwd=%s", *server, *authport, *secret, *user, *passwd))
	radreq, err := getAuthRequest(*user, *passwd, *usermac)
	if err != nil {
		return err
	}
	fmt.Fprintln(os.Stdout, radiusd.FmtPacket(radreq))
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(*timeout))
	defer cancel()
	radresp, err := radius.Exchange(ctx, radreq, fmt.Sprintf("%s:%d", *server, *authport))
	if err != nil {
		return err
	}
	fmt.Fprintln(os.Stdout, radiusd.FmtPacket(radresp))
	return nil
}

func sendAcct() error {
	sessionid, _ := common.UUIDBase32()
	if *acctstart {
		startreq := getAcctRequest(*user, *userip, *usermac, sessionid, rfc2866.AcctStatusType_Value_Start)
		fmt.Fprintln(os.Stdout, radiusd.FmtPacket(startreq))
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(*timeout))
		defer cancel()
		radresp, err := radius.Exchange(ctx, startreq, fmt.Sprintf("%s:%d", *server, *acctport))
		if err != nil {
			return err
		}
		fmt.Fprintln(os.Stdout, radiusd.FmtPacket(radresp))
	}
	if *acctupdate {
		startreq := getAcctRequest(*user, *userip, *usermac, sessionid, rfc2866.AcctStatusType_Value_InterimUpdate)
		fmt.Fprintln(os.Stdout, radiusd.FmtPacket(startreq))
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(*timeout))
		defer cancel()
		radresp, err := radius.Exchange(ctx, startreq, fmt.Sprintf("%s:%d", *server, *acctport))
		if err != nil {
			return err
		}
		fmt.Fprintln(os.Stdout, radiusd.FmtPacket(radresp))
	}
	if *acctstop {
		startreq := getAcctRequest(*user, *userip, *usermac, sessionid, rfc2866.AcctStatusType_Value_Stop)
		fmt.Fprintln(os.Stdout, radiusd.FmtPacket(startreq))
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(*timeout))
		defer cancel()
		radresp, err := radius.Exchange(ctx, startreq, fmt.Sprintf("%s:%d", *server, *acctport))
		if err != nil {
			return err
		}
		fmt.Fprintln(os.Stdout, radiusd.FmtPacket(radresp))
	}
	return nil
}

func TransBenchmarkAuth(user *User, bmtask *BenchmarkTask) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}()
	if bmtask.Counter >= *total {
		return
	}
	bmtask.IncrCounter("Counter")
	var client = &radius.Client{
		Retry:              0,
		MaxPacketErrors:    0,
		InsecureSkipVerify: true,
	}
	radreq, err := getAuthRequest(user.Username, user.Password, user.Macaddr)
	bmtask.IncrCounter("AuthReq")
	if err != nil || radreq == nil {
		bmtask.IncrCounter("AuthDrop")
		return
	}

	bmtask.IncrReqBytes(int64(radiusd.Length(radreq)))
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(*timeout))
	defer cancel()
	stattime := time.Now()
	radresp, err := client.Exchange(ctx, radreq, fmt.Sprintf("%s:%d", *server, *authport))
	bmtask.IncrRespBytes(int64(radiusd.Length(radresp)))
	bmtask.IncrAuthCast(time.Since(stattime).Milliseconds())
	if err != nil || radresp == nil {
		bmtask.IncrCounter("AuthDrop")
		if err == context.DeadlineExceeded {
			bmtask.IncrCounter("AuthTimeout")
		}

		return
	}

	bmtask.IncrReqBytes(int64(radiusd.Length(radresp)))

	switch radresp.Code {
	case radius.CodeAccessReject:
		bmtask.IncrCounter("AuthReject")
		return
	case radius.CodeAccessAccept:
		bmtask.IncrCounter("AuthAccept")
	}

	sessionid := common.UUID()
	// accounting start
	startreq := getAcctRequest(user.Username, user.Password, user.Macaddr, sessionid, rfc2866.AcctStatusType_Value_Start)
	bmtask.IncrCounter("AcctStart")
	bmtask.IncrReqBytes(int64(radiusd.Length(startreq)))
	ctx2, cancel2 := context.WithTimeout(context.Background(), time.Second*time.Duration(*timeout))
	defer cancel2()
	stattime = time.Now()
	startresp, err := client.Exchange(ctx2, startreq, fmt.Sprintf("%s:%d", *server, *acctport))
	bmtask.IncrRespBytes(int64(radiusd.Length(startresp)))
	bmtask.IncrAcctCast(time.Since(stattime).Milliseconds())
	if err != nil || startresp == nil {
		if err == context.DeadlineExceeded {
			bmtask.IncrCounter("AcctTimeout")
		}
		bmtask.IncrCounter("AcctDrop")
		return
	}
	bmtask.IncrCounter("AcctResp")
	bmtask.IncrCounter("OnlineStart")
	// time.Sleep(time.Second * 10)
	// accounting update
	upreq := getAcctRequest(user.Username, user.Password, user.Macaddr, sessionid, rfc2866.AcctStatusType_Value_InterimUpdate)
	bmtask.IncrReqBytes(int64(radiusd.Length(upreq)))
	bmtask.IncrCounter("AcctUpdate")
	ctx3, cancel3 := context.WithTimeout(context.Background(), time.Second*time.Duration(*timeout))
	defer cancel3()
	stattime = time.Now()
	upresp, err := client.Exchange(ctx3, upreq, fmt.Sprintf("%s:%d", *server, *acctport))
	bmtask.IncrRespBytes(int64(radiusd.Length(upresp)))
	bmtask.IncrAcctCast(time.Since(stattime).Milliseconds())
	if err != nil || upresp == nil {
		if err == context.DeadlineExceeded {
			bmtask.IncrCounter("AcctTimeout")
		}
		bmtask.IncrCounter("AcctDrop")
		return
	}
	bmtask.IncrCounter("AcctResp")
	// time.Sleep(time.Second * 10)
	// accounting update
	stopreq := getAcctRequest(user.Username, user.Password, user.Macaddr, sessionid, rfc2866.AcctStatusType_Value_Stop)
	bmtask.IncrReqBytes(int64(radiusd.Length(stopreq)))
	bmtask.IncrCounter("AcctStop")
	ctx4, cancel4 := context.WithTimeout(context.Background(), time.Second*time.Duration(*timeout))
	defer cancel4()
	stattime = time.Now()
	stopresp, err := client.Exchange(ctx4, stopreq, fmt.Sprintf("%s:%d", *server, *acctport))
	bmtask.IncrRespBytes(int64(radiusd.Length(stopresp)))
	bmtask.IncrAcctCast(time.Since(stattime).Milliseconds())
	if err != nil || stopresp == nil {
		if err == context.DeadlineExceeded {
			bmtask.IncrCounter("AcctTimeout")
		}
		bmtask.IncrCounter("AcctDrop")
		return
	}
	bmtask.IncrCounter("AcctResp")
	bmtask.IncrCounter("OnlineStop")
}

func TestBenchmark() {
	users := loadUserData()
	if len(users) == 0 {
		users = append(users, User{Username: *user, Password: *passwd, Ipaddr: *userip, Macaddr: *usermac})
	}
	bmtask := NewBenchmarkTask()

	glimit := NewGoLimit(int(*concurrency))
	go func() {
		for {
			if bmtask.Counter >= *total {
				break
			}
			for _, user := range users {
				if bmtask.Counter >= *total {
					return
				}
				glimit.Add()
				go func(g *GoLimit, u *User) {
					TransBenchmarkAuth(u, bmtask)
					g.Done()
				}(glimit, &user)
			}
		}
	}()

	for {
		if bmtask.Counter >= *total {
			break
		}
		bmtask.UpdateStat()
		bmtask.PrintCurrentStat()
		time.Sleep(time.Second * time.Duration(*statInterval))
	}
	// batch.QueueComplete()
	// batch.WaitAll()
	// for r := range batch.Results() {
	// 	fmt.Fprintln(os.Stdout, bmtask.Counter, r)
	// }
	// workpool.Close()
	bmtask.UpdateStat()
	bmtask.PrintCurrentStat()
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.Parse()
	printHelp()

	hinfo, err := host.Info()
	if err == nil {
		fmt.Fprintln(os.Stdout, fmt.Sprintf("\n-- System info: %s %s %s %s", hinfo.OS, hinfo.KernelArch, hinfo.KernelVersion, hinfo.Platform))

	}

	meminfo, err := mem.VirtualMemory()
	if err == nil {
		fmt.Fprintln(os.Stdout, fmt.Sprintf("-- Memary Total %d MB, Available %d MB", meminfo.Total/1048576, meminfo.Available/1048576))
	}

	cinfo, _ := cpu.Info()
	for _, c := range cinfo {
		fmt.Fprintln(os.Stdout, fmt.Sprintf("-- Cpu %d %s Cores %d %s  Cache %d\n", c.CPU, c.ModelName, c.Cores, c.VendorID, c.CacheSize))
	}

	if *auth {
		fmt.Fprintln(os.Stdout, "Radius Auth test... ", sendAuth())
		return
	}

	if *acctstart || *acctupdate || *acctstop {
		fmt.Fprintln(os.Stdout, "Radius Auth test... ", sendAcct())
		return
	}

	if *benchmark {
		fmt.Fprintln(os.Stdout, fmt.Sprintf("Start Benchmark Testing: Thread Concurrency %d, Total %d, Stat interval %d sec, RadsecRequest Timeout %d sec\n", *concurrency, *total, *statInterval, *timeout))
		TestBenchmark()
		fmt.Fprintln(os.Stdout, fmt.Sprintf("Start Benchmark Done: Thread Concurrency %d, Total %d, Stat interval %d sec, RadsecRequest Timeout %d sec\n", *concurrency, *total, *statInterval, *timeout))
	}

}

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
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
	"github.com/talkincode/toughradius/common"
	"github.com/talkincode/toughradius/toughradius"
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
	fmt.Fprintln(os.Stdout, toughradius.FmtPacket(radreq))
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(*timeout))
	defer cancel()
	radresp, err := radius.Exchange(ctx, radreq, fmt.Sprintf("%s:%d", *server, *authport))
	if err != nil {
		return err
	}
	fmt.Fprintln(os.Stdout, toughradius.FmtPacket(radresp))
	return nil
}

func sendAcct() error {
	sessionid, _ := common.UUIDBase32()
	if *acctstart {
		startreq := getAcctRequest(*user, *userip, *usermac, sessionid, rfc2866.AcctStatusType_Value_Start)
		fmt.Fprintln(os.Stdout, toughradius.FmtPacket(startreq))
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(*timeout))
		defer cancel()
		radresp, err := radius.Exchange(ctx, startreq, fmt.Sprintf("%s:%d", *server, *acctport))
		if err != nil {
			return err
		}
		fmt.Fprintln(os.Stdout, toughradius.FmtPacket(radresp))
	}
	if *acctupdate {
		startreq := getAcctRequest(*user, *userip, *usermac, sessionid, rfc2866.AcctStatusType_Value_InterimUpdate)
		fmt.Fprintln(os.Stdout, toughradius.FmtPacket(startreq))
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(*timeout))
		defer cancel()
		radresp, err := radius.Exchange(ctx, startreq, fmt.Sprintf("%s:%d", *server, *acctport))
		if err != nil {
			return err
		}
		fmt.Fprintln(os.Stdout, toughradius.FmtPacket(radresp))
	}
	if *acctstop {
		startreq := getAcctRequest(*user, *userip, *usermac, sessionid, rfc2866.AcctStatusType_Value_Stop)
		fmt.Fprintln(os.Stdout, toughradius.FmtPacket(startreq))
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(*timeout))
		defer cancel()
		radresp, err := radius.Exchange(ctx, startreq, fmt.Sprintf("%s:%d", *server, *acctport))
		if err != nil {
			return err
		}
		fmt.Fprintln(os.Stdout, toughradius.FmtPacket(radresp))
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

	bmtask.IncrReqBytes(int64(toughradius.Length(radreq)))
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(*timeout))
	defer cancel()
	stattime := time.Now()
	radresp, err := client.Exchange(ctx, radreq, fmt.Sprintf("%s:%d", *server, *authport))
	bmtask.IncrRespBytes(int64(toughradius.Length(radresp)))
	bmtask.IncrAuthCast(time.Since(stattime).Milliseconds())
	if err != nil || radresp == nil {
		bmtask.IncrCounter("AuthDrop")
		if err == context.DeadlineExceeded {
			bmtask.IncrCounter("AuthTimeout")
		}

		return
	}

	bmtask.IncrReqBytes(int64(toughradius.Length(radresp)))

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
	bmtask.IncrReqBytes(int64(toughradius.Length(startreq)))
	ctx2, cancel2 := context.WithTimeout(context.Background(), time.Second*time.Duration(*timeout))
	defer cancel2()
	stattime = time.Now()
	startresp, err := client.Exchange(ctx2, startreq, fmt.Sprintf("%s:%d", *server, *acctport))
	bmtask.IncrRespBytes(int64(toughradius.Length(startresp)))
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
	bmtask.IncrReqBytes(int64(toughradius.Length(upreq)))
	bmtask.IncrCounter("AcctUpdate")
	ctx3, cancel3 := context.WithTimeout(context.Background(), time.Second*time.Duration(*timeout))
	defer cancel3()
	stattime = time.Now()
	upresp, err := client.Exchange(ctx3, upreq, fmt.Sprintf("%s:%d", *server, *acctport))
	bmtask.IncrRespBytes(int64(toughradius.Length(upresp)))
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
	bmtask.IncrReqBytes(int64(toughradius.Length(stopreq)))
	bmtask.IncrCounter("AcctStop")
	ctx4, cancel4 := context.WithTimeout(context.Background(), time.Second*time.Duration(*timeout))
	defer cancel4()
	stattime = time.Now()
	stopresp, err := client.Exchange(ctx4, stopreq, fmt.Sprintf("%s:%d", *server, *acctport))
	bmtask.IncrRespBytes(int64(toughradius.Length(stopresp)))
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

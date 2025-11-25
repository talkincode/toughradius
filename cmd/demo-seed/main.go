package main

import (
	"flag"
	"fmt"
	"math/rand"
	"time"

	"github.com/talkincode/toughradius/v9/config"
	"github.com/talkincode/toughradius/v9/internal/app"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const demoMarker = "demo-seed"

type demoSeeder struct {
	db              *gorm.DB
	now             time.Time
	historyDays     int
	ctx             seedContext
	onlineCount     int
	accountingCount int
}

type seedContext struct {
	nodes    map[string]*domain.NetNode
	nas      map[string]*domain.NetNas
	profiles map[string]*domain.RadiusProfile
	users    map[string]*domain.RadiusUser
}

func main() {
	cfgPath := flag.String("c", "toughradius.yml", "path to config file")
	days := flag.Int("days", 7, "number of days of accounting history to generate")
	flag.Parse()

	cfg := config.LoadConfig(*cfgPath)
	application := app.NewApplication(cfg)
	application.Init(cfg)
	defer application.Release()

	seeder := &demoSeeder{
		db:          application.DB(),
		now:         time.Now(),
		historyDays: *days,
	}

	if err := seeder.Run(); err != nil {
		zap.L().Fatal("seed data failed", zap.Error(err))
	}

	fmt.Printf("Demo data inserted successfully!\n")
	fmt.Printf("  Profiles: %d\n", len(seeder.ctx.profiles))
	fmt.Printf("  Users: %d\n", len(seeder.ctx.users))
	fmt.Printf("  Online sessions: %d\n", seeder.onlineCount)
	fmt.Printf("  Accounting records: %d\n", seeder.accountingCount)
}

func (s *demoSeeder) Run() error {
	s.ctx = seedContext{}
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.cleanup(tx); err != nil {
			return err
		}
		if err := s.seedNodes(tx); err != nil {
			return err
		}
		if err := s.seedNAS(tx); err != nil {
			return err
		}
		if err := s.seedProfiles(tx); err != nil {
			return err
		}
		if err := s.seedUsers(tx); err != nil {
			return err
		}
		if err := s.seedOnlineSessions(tx); err != nil {
			return err
		}
		if err := s.seedAccountingHistory(tx); err != nil {
			return err
		}
		return nil
	})
}

func (s *demoSeeder) cleanup(tx *gorm.DB) error {
	if err := tx.Where("nas_class = ?", demoMarker).Delete(&domain.RadiusAccounting{}).Error; err != nil {
		return err
	}
	if err := tx.Where("nas_class = ?", demoMarker).Delete(&domain.RadiusOnline{}).Error; err != nil {
		return err
	}
	if err := tx.Where("remark = ?", demoMarker).Delete(&domain.RadiusUser{}).Error; err != nil {
		return err
	}
	if err := tx.Where("remark = ?", demoMarker).Delete(&domain.RadiusProfile{}).Error; err != nil {
		return err
	}
	if err := tx.Where("remark = ?", demoMarker).Delete(&domain.NetNas{}).Error; err != nil {
		return err
	}
	if err := tx.Where("remark = ?", demoMarker).Delete(&domain.NetNode{}).Error; err != nil {
		return err
	}
	return nil
}

func (s *demoSeeder) seedNodes(tx *gorm.DB) error {
	s.ctx.nodes = make(map[string]*domain.NetNode)
	specs := []struct {
		Name string
		Tags string
	}{
		{Name: "demo-core", Tags: "core,metro"},
		{Name: "demo-edge", Tags: "edge,ftth"},
	}

	for _, spec := range specs {
		record := domain.NetNode{
			Name:   spec.Name,
			Remark: demoMarker,
			Tags:   spec.Tags,
		}
		if err := tx.Where("name = ?", spec.Name).
			Assign(record).
			FirstOrCreate(&record).Error; err != nil {
			return err
		}
		s.ctx.nodes[spec.Name] = &record
	}
	return nil
}

func (s *demoSeeder) seedNAS(tx *gorm.DB) error {
	s.ctx.nas = make(map[string]*domain.NetNas)
	specs := []struct {
		Name       string
		Identifier string
		Hostname   string
		IPAddr     string
		Secret     string
		Node       string
		VendorCode string
	}{
		{Name: "demo-bras-1", Identifier: "demo-bras-1", Hostname: "bras1.demo.local", IPAddr: "10.0.0.1", Secret: "demo-secret", Node: "demo-core", VendorCode: "2011"},
		{Name: "demo-bras-2", Identifier: "demo-bras-2", Hostname: "bras2.demo.local", IPAddr: "10.0.1.1", Secret: "demo-secret", Node: "demo-edge", VendorCode: "25506"},
	}

	for _, spec := range specs {
		node := s.ctx.nodes[spec.Node]
		if node == nil {
			return fmt.Errorf("node %s not found", spec.Node)
		}
		record := domain.NetNas{
			NodeId:     node.ID,
			Name:       spec.Name,
			Identifier: spec.Identifier,
			Hostname:   spec.Hostname,
			Ipaddr:     spec.IPAddr,
			Secret:     spec.Secret,
			Status:     "enabled",
			VendorCode: spec.VendorCode,
			Tags:       "demo",
			Remark:     demoMarker,
		}
		if err := tx.Where("identifier = ?", spec.Identifier).
			Assign(record).
			FirstOrCreate(&record).Error; err != nil {
			return err
		}
		s.ctx.nas[spec.Name] = &record
	}
	return nil
}

func (s *demoSeeder) seedProfiles(tx *gorm.DB) error {
	s.ctx.profiles = make(map[string]*domain.RadiusProfile)
	node := s.ctx.nodes["demo-edge"]
	if node == nil {
		return fmt.Errorf("default node not found")
	}
	specs := []struct {
		Name      string
		UpRate    int
		DownRate  int
		ActiveNum int
	}{
		{Name: "demo-basic", UpRate: 50_000, DownRate: 200_000, ActiveNum: 2},
		{Name: "demo-premium", UpRate: 100_000, DownRate: 500_000, ActiveNum: 4},
		{Name: "demo-vip", UpRate: 200_000, DownRate: 1_000_000, ActiveNum: 8},
	}

	for _, spec := range specs {
		record := domain.RadiusProfile{
			NodeId:    node.ID,
			Name:      spec.Name,
			Status:    "enabled",
			ActiveNum: spec.ActiveNum,
			UpRate:    spec.UpRate,
			DownRate:  spec.DownRate,
			Remark:    demoMarker,
		}
		if err := tx.Where("name = ?", spec.Name).
			Assign(record).
			FirstOrCreate(&record).Error; err != nil {
			return err
		}
		s.ctx.profiles[spec.Name] = &record
	}
	return nil
}

func (s *demoSeeder) seedUsers(tx *gorm.DB) error {
	s.ctx.users = make(map[string]*domain.RadiusUser)
	specs := []struct {
		Username   string
		Realname   string
		Profile    string
		Node       string
		Status     string
		ExpireDays int
		Mobile     string
		Addr       string
	}{
		{"demo-alice", "Alice Chen", "demo-basic", "demo-core", "enabled", 180, "13800000001", "Building A"},
		{"demo-bob", "Bob Li", "demo-basic", "demo-edge", "enabled", 365, "13800000002", "Building B"},
		{"demo-carol", "Carol Wu", "demo-premium", "demo-core", "enabled", 120, "13800000003", "Campus East"},
		{"demo-dave", "Dave Zhang", "demo-premium", "demo-edge", "disabled", 60, "13800000004", "Campus West"},
		{"demo-eve", "Eve Qian", "demo-vip", "demo-core", "enabled", 365, "13800000005", "HQ"},
		{"demo-frank", "Frank Gu", "demo-vip", "demo-edge", "enabled", 90, "13800000006", "Branch"},
	}

	for idx, spec := range specs {
		profile := s.ctx.profiles[spec.Profile]
		node := s.ctx.nodes[spec.Node]
		if profile == nil || node == nil {
			return fmt.Errorf("missing profile or node for user %s", spec.Username)
		}
		record := domain.RadiusUser{
			NodeId:     node.ID,
			ProfileId:  profile.ID,
			Username:   spec.Username,
			Password:   "123456",
			Realname:   spec.Realname,
			Mobile:     spec.Mobile,
			AddrPool:   "demo-pool",
			ActiveNum:  2,
			UpRate:     profile.UpRate,
			DownRate:   profile.DownRate,
			Vlanid1:    100 + idx,
			IpAddr:     fmt.Sprintf("10.8.0.%d", 10+idx),
			MacAddr:    fmt.Sprintf("00:11:22:33:44:%02X", idx),
			Status:     spec.Status,
			ExpireTime: s.now.AddDate(0, 0, spec.ExpireDays),
			Remark:     demoMarker,
		}
		if err := tx.Where("username = ?", spec.Username).
			Assign(record).
			FirstOrCreate(&record).Error; err != nil {
			return err
		}
		s.ctx.users[spec.Username] = &record
	}
	return nil
}

func (s *demoSeeder) seedOnlineSessions(tx *gorm.DB) error {
	if err := tx.Where("nas_class = ?", demoMarker).Delete(&domain.RadiusOnline{}).Error; err != nil {
		return err
	}

	specs := []struct {
		Username string
		NasName  string
		IP       string
		Mac      string
		StartAgo time.Duration
		Duration time.Duration
		UpMB     int64
		DownMB   int64
		PortType int
	}{
		{"demo-alice", "demo-bras-1", "10.66.0.10", "4C:1D:AA:12:34:10", 45 * time.Minute, 2 * time.Hour, 900, 1800, 15},
		{"demo-carol", "demo-bras-1", "10.66.0.11", "4C:1D:AA:12:34:11", 30 * time.Minute, 90 * time.Minute, 550, 900, 15},
		{"demo-eve", "demo-bras-2", "10.88.0.12", "4C:1D:AA:12:34:12", 15 * time.Minute, 3 * time.Hour, 1500, 2100, 19},
		{"temp-guest", "demo-bras-2", "10.88.0.50", "4C:1D:AA:12:34:50", 10 * time.Minute, 40 * time.Minute, 120, 160, 18},
	}

	for idx, spec := range specs {
		nas := s.ctx.nas[spec.NasName]
		if nas == nil {
			return fmt.Errorf("nas %s not found", spec.NasName)
		}
		start := s.now.Add(-spec.StartAgo)
		online := domain.RadiusOnline{
			Username:          spec.Username,
			NasId:             nas.Identifier,
			NasAddr:           nas.Ipaddr,
			NasPaddr:          nas.Ipaddr,
			SessionTimeout:    7200,
			FramedIpaddr:      spec.IP,
			FramedNetmask:     "255.255.255.0",
			MacAddr:           spec.Mac,
			NasPort:           int64(1000 + idx),
			NasClass:          demoMarker,
			NasPortId:         fmt.Sprintf("gigabitethernet0/%d", idx+1),
			NasPortType:       spec.PortType,
			ServiceType:       2,
			AcctSessionId:     fmt.Sprintf("demo-session-%d", idx+1),
			AcctSessionTime:   int(spec.Duration.Seconds()),
			AcctInputTotal:    spec.UpMB * 1024 * 1024,
			AcctOutputTotal:   spec.DownMB * 1024 * 1024,
			AcctInputPackets:  2000 + idx*300,
			AcctOutputPackets: 3000 + idx*400,
			AcctStartTime:     start,
			LastUpdate:        s.now,
		}
		if err := tx.Create(&online).Error; err != nil {
			return err
		}
		s.onlineCount++
	}
	return nil
}

func (s *demoSeeder) seedAccountingHistory(tx *gorm.DB) error {
	if err := tx.Where("nas_class = ?", demoMarker).Delete(&domain.RadiusAccounting{}).Error; err != nil {
		return err
	}

	if len(s.ctx.users) == 0 {
		return fmt.Errorf("no users to generate accounting history")
	}
	nasList := make([]*domain.NetNas, 0, len(s.ctx.nas))
	for _, nas := range s.ctx.nas {
		nasList = append(nasList, nas)
	}
	usernames := make([]string, 0, len(s.ctx.users))
	for name := range s.ctx.users {
		usernames = append(usernames, name)
	}
	rng := rand.New(rand.NewSource(42))
	startDay := startOfDay(s.now).AddDate(0, 0, -(s.historyDays - 1))

	for day := 0; day < s.historyDays; day++ {
		dayStart := startDay.AddDate(0, 0, day)
		for idx, username := range usernames {
			if idx >= 4 {
				break
			}
			sessionStart := dayStart.Add(time.Duration((idx*3+day)%24) * time.Hour).Add(time.Duration(rng.Intn(30)) * time.Minute)
			duration := time.Duration(30+rng.Intn(90)) * time.Minute
			upBytes := int64(400+day*30+idx*20) * 1024 * 1024
			downBytes := int64(900+day*40+idx*25) * 1024 * 1024
			nas := nasList[(day+idx)%len(nasList)]
			accounting := domain.RadiusAccounting{
				Username:          username,
				AcctSessionId:     fmt.Sprintf("demo-acct-%d-%d-%s", day, idx, username),
				NasId:             nas.Identifier,
				NasAddr:           nas.Ipaddr,
				NasPaddr:          nas.Ipaddr,
				SessionTimeout:    7200,
				FramedIpaddr:      fmt.Sprintf("10.99.%d.%d", idx+1, day+10),
				FramedNetmask:     "255.255.255.0",
				MacAddr:           fmt.Sprintf("6C:FA:A7:%02X:%02X:%02X", day, idx, (day+idx)%255),
				ServiceType:       2,
				NasPort:           int64(2000 + idx),
				NasPortId:         fmt.Sprintf("vlan/%d", 200+idx),
				NasPortType:       15,
				AcctSessionTime:   int(duration.Seconds()),
				AcctInputTotal:    upBytes,
				AcctOutputTotal:   downBytes,
				AcctInputPackets:  1500 + rng.Intn(900),
				AcctOutputPackets: 2000 + rng.Intn(1200),
				AcctStartTime:     sessionStart,
				AcctStopTime:      sessionStart.Add(duration),
				LastUpdate:        sessionStart.Add(duration),
				NasClass:          demoMarker,
			}
			if err := tx.Create(&accounting).Error; err != nil {
				return err
			}
			s.accountingCount++
		}
	}

	// Heavier traffic samples during last 24 hours for better charts
	hourlyStart := s.now.Add(-23 * time.Hour)
	topUsers := usernames
	if len(topUsers) > 2 {
		topUsers = topUsers[:2]
	}
	for hour := 0; hour < 24; hour++ {
		slot := hourlyStart.Add(time.Duration(hour) * time.Hour)
		for idx, username := range topUsers {
			nas := nasList[(hour+idx)%len(nasList)]
			upBytes := int64(120+hour*5+idx*30) * 1024 * 1024
			downBytes := int64(200+hour*7+idx*25) * 1024 * 1024
			accounting := domain.RadiusAccounting{
				Username:          username,
				AcctSessionId:     fmt.Sprintf("demo-24h-%d-%d-%s", hour, idx, username),
				NasId:             nas.Identifier,
				NasAddr:           nas.Ipaddr,
				NasPaddr:          nas.Ipaddr,
				SessionTimeout:    3600,
				FramedIpaddr:      fmt.Sprintf("10.200.%d.%d", idx+1, hour+1),
				FramedNetmask:     "255.255.255.0",
				MacAddr:           fmt.Sprintf("8A:BC:%02X:%02X:%02X:%02X", hour, idx, hour+idx, idx+10),
				ServiceType:       2,
				NasPort:           int64(3000 + idx),
				NasPortId:         fmt.Sprintf("pppoe-%d", hour+idx),
				NasPortType:       15,
				AcctSessionTime:   int((20 * time.Minute).Seconds()),
				AcctInputTotal:    upBytes,
				AcctOutputTotal:   downBytes,
				AcctInputPackets:  800 + hour*10 + idx*50,
				AcctOutputPackets: 900 + hour*12 + idx*60,
				AcctStartTime:     slot,
				AcctStopTime:      slot.Add(20 * time.Minute),
				LastUpdate:        slot.Add(20 * time.Minute),
				NasClass:          demoMarker,
			}
			if err := tx.Create(&accounting).Error; err != nil {
				return err
			}
			s.accountingCount++
		}
	}
	return nil
}

func startOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

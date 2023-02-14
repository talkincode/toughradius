package app

import (
	"fmt"
	"os"
	"time"

	"github.com/nakabonne/tstorage"
	"github.com/robfig/cron/v3"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/process"
	"github.com/spf13/cast"
	"github.com/talkincode/toughradius/common/zaplog"
	"github.com/talkincode/toughradius/common/zaplog/log"
	"github.com/talkincode/toughradius/models"
)

var cronParser = cron.NewParser(
	cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor,
)

func (a *Application) initJob() {
	loc, _ := time.LoadLocation(a.appConfig.System.Location)
	a.sched = cron.New(cron.WithLocation(loc), cron.WithParser(cronParser))

	var err error
	_, err = a.sched.AddFunc("@every 30s", func() {
		go a.SchedSystemMonitorTask()
		go a.SchedProcessMonitorTask()
	})

	_, err = a.sched.AddFunc("@every 60s", func() {
		a.SchedUpdateBatchCwmpStatus()
	})

	// database backup
	_, err = a.sched.AddFunc("@daily", func() {
		err := app.BackupDatabase()
		if err != nil {
			log.Errorf("database backup err %s", err.Error())
		}
	})

	_, err = a.sched.AddFunc("@daily", func() {
		a.gormDB.
			Where("opt_time < ? ", time.Now().
				Add(-time.Hour*24*365)).Delete(models.SysOprLog{})
	})

	if err != nil {
		log.Errorf("init job error %s", err.Error())
	}

	a.setupCwmpTask()

	a.sched.Start()
}

// SchedSystemMonitorTask system monitor
func (a *Application) SchedSystemMonitorTask() {
	defer func() {
		if err := recover(); err != nil {
			log.Error(err)
		}
	}()

	timestamp := time.Now().Unix()

	var cpuuse float64
	_cpuuse, err := cpu.Percent(0, false)
	if err == nil && len(_cpuuse) > 0 {
		cpuuse = _cpuuse[0]
	}
	err = zaplog.TSDB().InsertRows([]tstorage.Row{
		{
			Metric: "system_cpuuse",
			DataPoint: tstorage.DataPoint{
				Value:     cpuuse,
				Timestamp: timestamp,
			},
		},
	})
	if err != nil {
		log.Error("add timeseries data error:", err.Error())
	}

	_meminfo, err := mem.VirtualMemory()
	var memuse uint64
	if err == nil {
		memuse = _meminfo.Used
	}

	err = zaplog.TSDB().InsertRows([]tstorage.Row{
		{
			Metric: "system_memuse",
			DataPoint: tstorage.DataPoint{
				Value:     float64(memuse),
				Timestamp: timestamp,
			},
		},
	})
	if err != nil {
		log.Error("add timeseries data error:", err.Error())
	}
}

// SchedProcessMonitorTask app process monitor
func (a *Application) SchedProcessMonitorTask() {
	defer func() {
		if err := recover(); err != nil {
			log.Error(err)
		}
	}()

	timestamp := time.Now().Unix()

	p, err := process.NewProcess(int32(os.Getpid()))
	if err != nil {
		return
	}

	cpuuse, err := p.CPUPercent()
	if err != nil {
		cpuuse = 0
	}

	err = zaplog.TSDB().InsertRows([]tstorage.Row{
		{
			Metric: "toughradius_cpuuse",
			DataPoint: tstorage.DataPoint{
				Value:     cpuuse,
				Timestamp: timestamp,
			},
		},
	})
	if err != nil {
		log.Error("add timeseries data error:", err.Error())
	}

	meminfo, err := p.MemoryInfo()
	if err != nil {
		return
	}
	memuse := meminfo.RSS / 1024 / 1024

	err = zaplog.TSDB().InsertRows([]tstorage.Row{
		{
			Metric: "toughradius_memuse",
			DataPoint: tstorage.DataPoint{
				Value:     float64(memuse),
				Timestamp: timestamp,
			},
		},
	})
	if err != nil {
		log.Error("add timeseries data error:", err.Error())
	}
}

func (a *Application) SchedClearExpireData() {
	defer func() {
		if err := recover(); err != nil {
			log.Error(err)
		}
	}()
	// Clean expire online
	a.gormDB.Where("last_update <= ?",
		time.Now().Add(time.Second*300*-1)).
		Delete(&models.RadiusOnline{})

	// Clean up accounting logs
	hdays := a.GetSettingsStringValue("radius", ConfigRadiusAccountingHistoryDays)
	idays := cast.ToInt(hdays)
	if idays == 0 {
		idays = 90
	}
	a.gormDB.
		Where("acct_stop_time < ? ", time.Now().
			Add(-time.Hour*24*time.Duration(idays))).Delete(models.RadiusAccounting{})
}

func (a *Application) SchedUpdateBatchCwmpStatus() {
	a.gormDB.Model(&models.NetCpe{}).
		Where("cwmp_last_inform < ? and cwmp_status <> 'offline' ", time.Now().Add(-time.Second*300)).
		Update("cwmp_status", "offline")
}

func (a *Application) setupCwmpTask() {
	var err error
	_, err = a.sched.AddFunc("@every 5m", func() {
		_ = CreateCwmpScheduledTask("5m")
	})
	_, err = a.sched.AddFunc("@every 10m", func() {
		_ = CreateCwmpScheduledTask("10m")
	})
	_, err = a.sched.AddFunc("@every 30m", func() {
		_ = CreateCwmpScheduledTask("30m")
	})
	_, err = a.sched.AddFunc("@every 1h", func() {
		_ = CreateCwmpScheduledTask("1h")
	})
	_, err = a.sched.AddFunc("@every 4h", func() {
		_ = CreateCwmpScheduledTask("4h")
	})
	_, err = a.sched.AddFunc("@every 8h", func() {
		_ = CreateCwmpScheduledTask("8h")
	})
	_, err = a.sched.AddFunc("@every 12", func() {
		_ = CreateCwmpScheduledTask("12")
	})

	// Execute at 0:01 every morning
	_, err = a.sched.AddFunc("0 1 0 * * *", func() {
		_ = CreateCwmpScheduledTask("daily@h0")
	})

	// 1 to 23 cycles
	for i := 1; i < 24; i++ {
		log.Infof("add job daily@h%d cron( 0 0 %d * * * )", i, i)
		key := fmt.Sprintf("daily@h%d", i)
		cronkey := fmt.Sprintf("0 0 %d * * *", i)
		_, err = a.sched.AddFunc(cronkey, func() {
			_ = CreateCwmpScheduledTask(key)
		})
	}

	// 18:30 every day
	// _, err = a.sched.AddFunc("0 30 18 * * *", func() {
	// 	_ = CreateCwmpScheduledTask("daily@h18m30")
	// })

	if err != nil {
		log.Error(err)
	}
}

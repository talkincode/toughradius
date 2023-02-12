package metrics

import (
	"fmt"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/process"
	"github.com/talkincode/toughradius/common/web"
	"github.com/talkincode/toughradius/webserver"
)

func InitRouter() {

	webserver.GET("/admin/metrics/system/hostname", func(c echo.Context) error {
		hinfo, err := host.Info()
		_host := "unknow"
		if err == nil {
			_host = hinfo.Hostname
		}
		return c.Render(http.StatusOK, "metrics", web.NewMetrics("mdi mdi-server", _host, "主机名"))
	})

	webserver.GET("/admin/metrics/system/os", func(c echo.Context) error {
		hinfo, err := host.Info()
		_os := "unknow"
		if err == nil {
			_os = hinfo.OS
		}
		return c.Render(http.StatusOK, "metrics",
			web.NewMetrics("mdi mdi-lifebuoy", _os, "操作系统"))
	})

	webserver.GET("/admin/metrics/system/diskuse", func(c echo.Context) error {
		_diskinfo, err := disk.Usage("/")
		_usage := 0.0
		_total := uint64(0)
		if err == nil {
			_usage = _diskinfo.UsedPercent
			_total = _diskinfo.Total / (1024 * 1024 * 1024)
		}
		return c.Render(http.StatusOK, "metrics",
			web.NewMetrics("mdi mdi-harddisk", fmt.Sprintf("%.2f%%", _usage),
				fmt.Sprintf("磁盘 / 占用 (总大小: %d G)", _total)))
	})

	webserver.GET("/admin/metrics/system/cpuusage", func(c echo.Context) error {
		_cpuuse, err := cpu.Percent(0, false)
		_cpucount, _ := cpu.Counts(false)
		if err != nil {
			_cpuuse = []float64{0}
		}
		return c.Render(http.StatusOK, "metrics",
			web.NewMetrics("mdi mdi-circle-slice-2",
				fmt.Sprintf("%.2f%%", _cpuuse[0]),
				fmt.Sprintf("Cpu %d Core", _cpucount)))
	})

	webserver.GET("/admin/metrics/system/main/cpuusage", func(c echo.Context) error {
		var cpuuse float64
		p, err := process.NewProcess(int32(os.Getpid()))
		if err != nil {
			cpuuse, _ = p.CPUPercent()
		}
		return c.Render(http.StatusOK, "metrics",
			web.NewMetrics("mdi mdi-circle-slice-2", fmt.Sprintf("%.2f%%", cpuuse), "主程序Cpu负载"))
	})

	webserver.GET("/admin/metrics/system/memusage", func(c echo.Context) error {
		_meminfo, err := mem.VirtualMemory()
		_usage := 0.0
		_total := uint64(0)
		if err == nil {
			_usage = _meminfo.UsedPercent
			_total = _meminfo.Total / (1000 * 1000 * 1000)
		}
		return c.Render(http.StatusOK, "metrics",
			web.NewMetrics("mdi mdi-memory", fmt.Sprintf("%.2f%%", _usage),
				fmt.Sprintf("Memory Total: %d G", _total)))
	})

	webserver.GET("/admin/metrics/system/main/memusage", func(c echo.Context) error {
		var memuse uint64
		p, err := process.NewProcess(int32(os.Getpid()))
		if err == nil {
			meminfo, err := p.MemoryInfo()
			if err == nil {
				memuse = meminfo.RSS / 1024 / 1024
			}
		}

		return c.Render(http.StatusOK, "metrics",
			web.NewMetrics("mdi mdi-memory", fmt.Sprintf("%d MB", memuse), "主程序内存使用"))
	})

	webserver.GET("/admin/metrics/system/uptime", func(c echo.Context) error {
		hinfo, err := host.Info()
		_hour := uint64(0)
		if err == nil {
			_hour = hinfo.Uptime
		}
		return c.Render(http.StatusOK, "metrics",
			web.NewMetrics("mdi mdi-clock",
				fmt.Sprintf("%.1f Hour",
					float64(_hour)/float64(3600)), "运行时长"))
	})

	webserver.GET("/admin/metrics/unknow", func(c echo.Context) error {
		return c.Render(http.StatusOK, "metrics",
			web.NewMetrics("mdi mdi-lifebuoy", "N/A", "Unknow"))
	})

}

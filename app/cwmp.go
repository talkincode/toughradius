package app

import (
	"bytes"
	"errors"
	"os"
	"path"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/talkincode/toughradius/assets"
	"github.com/talkincode/toughradius/common"
	"github.com/talkincode/toughradius/common/cwmp"
	"github.com/talkincode/toughradius/common/timeutil"
	"github.com/talkincode/toughradius/common/web"
	"github.com/talkincode/toughradius/common/zaplog/log"
	"github.com/talkincode/toughradius/models"
)

type CwmpEventTable struct {
	cpeTable map[string]*CwmpCpe
	cpeLock  sync.Mutex
}

type CwmpCpe struct {
	Sn              string `json:"sn"`
	OUI             string `json:"oui"`
	taskTags        []string
	SoftwareVersion string `json:"software_version"`
	Manufacturer    string `json:"manufacturer"`
	ProductClass    string `json:"product_class"`
	cwmpQueueMap    chan models.CwmpEventData
	cwmpHPQueueMap  chan models.CwmpEventData
	LastInform      *cwmp.Inform `json:"latest_message"`
	LastUpdate      time.Time    `json:"last_update"`
	LastDataNotify  time.Time    `json:"last_data_notify"`
	IsRegister      bool         `json:"is_register"`
}

func NewCwmpEventTable() *CwmpEventTable {
	et := &CwmpEventTable{
		cpeTable: make(map[string]*CwmpCpe),
		cpeLock:  sync.Mutex{},
	}
	return et
}

func GetCwmpCpe(key string) *CwmpCpe {
	return app.CwmpTable().GetCwmpCpe(key)
}

func (c *CwmpEventTable) Size() int {
	c.cpeLock.Lock()
	defer c.cpeLock.Unlock()
	return len(c.cpeTable)
}

func (c *CwmpEventTable) ListSn() []string {
	c.cpeLock.Lock()
	defer c.cpeLock.Unlock()
	var snlist = make([]string, 0)
	for s, _ := range c.cpeTable {
		snlist = append(snlist, s)
	}
	return snlist
}

func (c *CwmpEventTable) GetCwmpCpe(key string) *CwmpCpe {
	if common.IsEmptyOrNA(key) {
		panic(errors.New("key is empty"))
	}
	c.cpeLock.Lock()
	defer c.cpeLock.Unlock()
	cpe, ok := c.cpeTable[key]
	if !ok {
		var count int64 = 0
		app.gormDB.Model(models.NetCpe{}).Where("sn=?", key).Count(&count)
		cpe = &CwmpCpe{
			Sn:             key,
			LastUpdate:     timeutil.EmptyTime,
			LastDataNotify: timeutil.EmptyTime,
			cwmpQueueMap:   make(chan models.CwmpEventData, 512),
			cwmpHPQueueMap: make(chan models.CwmpEventData, 1),
			LastInform:     nil,
			IsRegister:     count > 0,
		}
		c.cpeTable[key] = cpe
	}
	return cpe
}

func (c *CwmpEventTable) ClearCwmpCpe(key string) {
	c.cpeLock.Lock()
	defer c.cpeLock.Unlock()
	delete(c.cpeTable, key)
}

func (c *CwmpEventTable) ClearCwmpCpeCache(key string) {
	cpe := c.GetCwmpCpe(key)
	cpe.taskTags = nil
}

func (c *CwmpEventTable) UpdateCwmpCpe(key string, msg *cwmp.Inform) {
	cpe := c.GetCwmpCpe(key)
	cpe.UpdateStatus(msg)
}

func (c *CwmpCpe) UpdateStatus(msg *cwmp.Inform) {
	c.LastInform = msg
	c.LastUpdate = time.Now()
	if msg.ProductClass != "" {
		c.ProductClass = msg.ProductClass
	}
	if msg.OUI != "" {
		c.OUI = msg.OUI
	}
	if msg.Manufacturer != "" {
		c.Manufacturer = msg.Manufacturer
	}
	if msg.GetSoftwareVersion() != "" {
		c.SoftwareVersion = msg.GetSoftwareVersion()
	}
}

func (c *CwmpCpe) NotifyDataUpdate(force bool) {
	var ctime = time.Now()
	updateFlag := ctime.Sub(c.LastDataNotify).Seconds() > 300
	if force {
		updateFlag = true
	}
	if updateFlag {
		// events.Bus.Publish(events.EventCwmpInformUpdate, c.Sn, c.LastInform)
		c.OnInformUpdate()
		c.LastDataNotify = time.Now()
		// log.Infof("CPE %s OnInformUpdate", c.Sn)
	} else {
		// events.Bus.Publish(events.EventCwmpInformUpdateOnline, c.Sn)
		c.OnInformUpdateOnline()
		// log.Infof("CPE %s OnInformUpdateOnline", c.Sn)
	}
}

func (c *CwmpCpe) getQueue(hp bool) chan models.CwmpEventData {
	var que = c.cwmpQueueMap
	if hp {
		que = c.cwmpHPQueueMap
	}
	return que
}

func (c *CwmpCpe) TaskTags() (tags []string) {
	if c.taskTags != nil {
		return c.taskTags
	}
	var _tags string
	app.gormDB.Raw("select task_tags from net_cpe where sn = ? ", c.Sn).Scan(&_tags)
	_tags2 := strings.Split(strings.TrimSpace(_tags), ",")
	for _, tag := range _tags2 {
		if tag != "" {
			tags = append(tags, tag)
		}
	}
	if len(tags) > 0 {
		c.taskTags = tags
	}
	return
}

func setMapValue(vmap map[string]interface{}, name string, value interface{}) {
	if name != "" && value != "" {
		vmap[name] = value
	}
}

// RecvCwmpEventData 接收一个 Cwmp 事件
func (c *CwmpCpe) RecvCwmpEventData(timeoutMsec int, hp bool) (data *models.CwmpEventData, err error) {
	select {
	case _data := <-c.getQueue(hp):
		return &_data, nil
	case <-time.After(time.Millisecond * time.Duration(timeoutMsec)):
		return nil, errors.New("read cwmp event channel timeout")
	}
}

// GetCwmpPresetEventData 获取一个 Cwmp 预设任务执行
func (c *CwmpCpe) GetCwmpPresetEventData() (data *models.CwmpEventData, err error) {

	return nil, err
}

// SendCwmpEventData 发送一个 Cwmp 事件，
func (c *CwmpCpe) SendCwmpEventData(data models.CwmpEventData, timeoutMsec int, hp bool) error {
	select {
	case c.getQueue(hp) <- data:
		return nil
	case <-time.After(time.Millisecond * time.Duration(timeoutMsec)):
		return errors.New("cwmp event channel full, write timeout")
	}
}

// CheckRegister 检查设备注册情况
func (c *CwmpCpe) CheckRegister(ip string, msg *cwmp.Inform) {
	if app.GetTr069SettingsStringValue(ConfigCpeAutoRegister) != "enabled" {
		return
	}
	if !c.IsRegister {
		var ctime = time.Now()
		err := app.gormDB.Create(&models.NetCpe{
			ID:           common.UUIDint64(),
			NodeId:       AutoRegisterPopNodeId,
			CpeType:      "mikrotik",
			Sn:           msg.Sn,
			Name:         "Device-" + msg.Sn,
			Model:        msg.ProductClass,
			VendorCode:   "14988",
			Oui:          msg.OUI,
			Manufacturer: msg.Manufacturer,
			ProductClass: msg.ProductClass,
			Status:       "",
			Tags:         "",
			Remark:       "first register from " + ip,
			CwmpUrl:      msg.GetParam("Device.ManagementServer.ConnectionRequestURL"),
			CreatedAt:    ctime,
			UpdatedAt:    ctime,
		}).Error
		if err == nil {
			log.Info("Auto register new device: %s", msg.Sn)
			c.IsRegister = true
		} else {
			log.Errorf("CheckRegister create cpe error: %s", err)
		}
	}
}

func (c *CwmpCpe) UpdateManagementAuthInfo(session string, timeout int, hp bool) error {
	return c.SendCwmpEventData(models.CwmpEventData{
		Session: session,
		Sn:      c.Sn,
		Message: &cwmp.SetParameterValues{
			ID:     session,
			Name:   "",
			NoMore: 0,
			Params: map[string]cwmp.ValueStruct{
				"Device.ManagementServer.ConnectionRequestUsername": {
					Type:  "xsd:string",
					Value: c.Sn,
				},
				"Device.ManagementServer.ConnectionRequestPassword": {
					Type:  "xsd:string",
					Value: app.GetTr069SettingsStringValue("CpeConnectionRequestPassword"),
				},
			},
		},
	}, timeout, hp)
}

func (c *CwmpCpe) ProcessParameterNamesResponse(msg *cwmp.GetParameterNamesResponse) {
	for _, param := range msg.Params {
		if param.Writable == "" {
			continue
		}
		app.gormDB.Model(&models.NetCpeParam{}).
			Where("sn = ? and name = ?", c.Sn, param.Name).
			Update("writable", param.Writable)
	}
}

func (c *CwmpCpe) OnInformUpdate() {
	msg := c.LastInform
	valmap := map[string]interface{}{}
	setMapValue(valmap, "manufacturer", msg.Manufacturer)
	setMapValue(valmap, "product_class", msg.ProductClass)
	setMapValue(valmap, "oui", msg.OUI)
	setMapValue(valmap, "cwmp_status", "online")
	setMapValue(valmap, "cwmp_last_inform", time.Now())
	setMapValue(valmap, "cwmp_url", msg.GetParam("Device.ManagementServer.ConnectionRequestURL"))
	setMapValue(valmap, "software_version", msg.GetParam("Device.DeviceInfo.SoftwareVersion"))
	setMapValue(valmap, "hardware_version", msg.GetParam("Device.DeviceInfo.HardwareVersion"))
	setMapValue(valmap, "model", msg.GetParam("Device.DeviceInfo.ModelName"))
	setMapValue(valmap, "uptime", msg.GetParam("Device.DeviceInfo.UpTime"))
	setMapValue(valmap, "cpu_usage", msg.GetParam("Device.DeviceInfo.ProcessStatus.CPUUsage"))
	setMapValue(valmap, "memory_total", msg.GetParam("Device.DeviceInfo.MemoryStatus.Free"))
	setMapValue(valmap, "memory_free", msg.GetParam("Device.DeviceInfo.MemoryStatus.Total"))
	// mikrotik
	setMapValue(valmap, "arch_name", msg.GetParam("Device.DeviceInfo.X_MIKROTIK_ArchName"))
	setMapValue(valmap, "system_name", msg.GetParam("Device.DeviceInfo.X_MIKROTIK_SystemIdentity"))

	if len(valmap) > 0 {
		err := app.gormDB.Model(&models.NetCpe{}).Where("sn=?", c.Sn).Updates(valmap)
		if err.Error != nil {
			log.Error("EventCwmpInformUpdate error: ", err)
		}
	}
}

func (c *CwmpCpe) OnInformUpdateOnline() {
	err := app.gormDB.Model(&models.NetCpe{}).Where("sn=?", c.Sn).Updates(map[string]interface{}{
		"cwmp_status":      "online",
		"cwmp_last_inform": time.Now(),
	}).Error
	if err != nil {
		log.Error("EventCwmpInformUpdateOnline error: ", err)
	}
}

func (c *CwmpCpe) OnParamsUpdate(params map[string]string) {
	var getParam = func(name string) string {
		v, ok := params[name]
		if ok {
			return v
		}
		return ""
	}
	valmap := map[string]interface{}{}
	setMapValue(valmap, "cwmp_last_inform", time.Now())
	setMapValue(valmap, "cwmp_status", "online")
	setMapValue(valmap, "cwmp_url", getParam("Device.ManagementServer.ConnectionRequestURL"))
	setMapValue(valmap, "software_version", getParam("Device.DeviceInfo.SoftwareVersion"))
	setMapValue(valmap, "hardware_version", getParam("Device.DeviceInfo.HardwareVersion"))
	setMapValue(valmap, "model", getParam("Device.DeviceInfo.ModelName"))
	setMapValue(valmap, "uptime", getParam("Device.DeviceInfo.UpTime"))
	setMapValue(valmap, "cpu_usage", getParam("Device.DeviceInfo.ProcessStatus.CPUUsage"))
	setMapValue(valmap, "memory_total", getParam("Device.DeviceInfo.MemoryStatus.Free"))
	setMapValue(valmap, "memory_free", getParam("Device.DeviceInfo.MemoryStatus.Total"))
	// mikrotik
	setMapValue(valmap, "arch_name", getParam("Device.DeviceInfo.X_MIKROTIK_ArchName"))
	setMapValue(valmap, "system_name", getParam("Device.DeviceInfo.X_MIKROTIK_SystemIdentity"))

	if len(valmap) > 0 {
		err := app.gormDB.Model(&models.NetCpe{}).Where("sn=?", c.Sn).Updates(valmap).Error
		if err != nil {
			log.Error("OnParamsUpdate error: ", err.Error())
		} else {
			log.Info("OnParamsUpdate success")
		}
	}
	app.UpdateCwmpCpeRundata(c.Sn, params)
}

func (a *Application) UpdateCwmpCpeRundata(sn string, vmap map[string]string) {
	var pids []string
	var params []models.NetCpeParam
	for k, v := range vmap {
		pid := common.Md5Hash(sn + k)
		if common.InSlice(pid, pids) {
			continue
		}
		tag := ""
		switch {
		case strings.Contains(k, "Device.DeviceInfo."):
			tag = "Device.DeviceInfo."
		case strings.Contains(k, "Device.ManagementServer."):
			tag = "Device.ManagementServer."
		case strings.Contains(k, "Device.InterfaceStack."):
			tag = "Device.InterfaceStack."
		case strings.Contains(k, "Device.Cellular."):
			tag = "Device.Cellular."
		case strings.Contains(k, "Device.Ethernet."):
			tag = "Device.Ethernet."
		case strings.Contains(k, "Device.WiFi."):
			tag = "Device.WiFi."
		case strings.Contains(k, "Device.PPP."):
			tag = "Device.PPP."
		case strings.Contains(k, "Device.IP."):
			tag = "Device.IP."
		case strings.Contains(k, "Device.Routing."):
			tag = "Device.Routing."
		case strings.Contains(k, "Device.Hosts."):
			tag = "Device.Hosts."
		case strings.Contains(k, "Device.DNS."):
			tag = "Device.DNS."
		case strings.Contains(k, "Device.DHCPv4."):
			tag = "Device.DHCPv4."
		case strings.Contains(k, "Device.Firewall."):
			tag = "Device.Firewall."
		case strings.Contains(k, "Device.X_MIKROTIK_Interface."):
			tag = "Device.X_MIKROTIK_Interface."
		}

		pids = append(pids, pid)
		params = append(params, models.NetCpeParam{
			ID:        pid,
			Sn:        sn,
			Tag:       tag,
			Name:      k,
			Value:     v,
			UpdatedAt: time.Now(),
		})
	}
	err := a.gormDB.Model(&models.NetCpeParam{}).Save(&params).Error
	if err != nil {
		log.Errorf("UpdateCwmpCPERundata: %s", err.Error())
	} else {
		log.Infof("UpdateCwmpCPERundata for %s success, total %d", sn, len(pids))
	}
}

func (a *Application) InjectCwmpConfigVars(sn string, src string, extvars map[string]string) string {
	var cpe models.NetCpe
	err := a.gormDB.Model(&models.NetCpe{}).Where("sn=?", sn).First(&cpe).Error
	if err != nil {
		log.Errorf("InjectCwmpConfigVars: %s", err.Error())
	}
	tx := template.Must(template.New("cpe_cwmp_config_content").Parse(src))
	var bs []byte
	buff := bytes.NewBuffer(bs)

	token, _ := web.CreateToken(a.appConfig.Tr069.Secret, "cpe", "api", time.Hour*24*365)

	vars := map[string]interface{}{
		"cpe":                              cpe,
		"ToughradiusApiToken":              token,
		ConfigTR069AccessAddress:           a.GetTr069SettingsStringValue(ConfigTR069AccessAddress),
		ConfigTR069AccessPassword:          a.GetTr069SettingsStringValue(ConfigTR069AccessPassword),
		ConfigCpeConnectionRequestPassword: a.GetTr069SettingsStringValue(ConfigCpeConnectionRequestPassword),
	}

	for k, v := range extvars {
		vars[k] = v
	}

	err = tx.Execute(buff, vars)
	if err != nil {
		log.Errorf("InjectCwmpConfigVars: %s", err.Error())
		return src
	}
	return buff.String()
}

func (a *Application) GetCacrtContent() string {
	caCert := path.Join(a.appConfig.System.Workdir, "private/ca.crt")
	crtdata, err := os.ReadFile(caCert)
	if err != nil {
		crtdata = assets.CaCrt
	}
	return strings.TrimSpace(string(crtdata))
}

func MatchDevice(c models.NetCpe, oui, productClass, softwareVersion string) bool {
	anySlice := []string{"", "any", "N/A", "all"}
	var ov, pv, sv int
	if !common.InSlice(oui, anySlice) &&
		!common.InSlice(c.Oui, strings.Split(oui, ",")) {
		ov = 1
	}
	if !common.InSlice(productClass, anySlice) &&
		!common.InSlice(c.ProductClass, strings.Split(productClass, ",")) {
		pv = 1
	}
	if !common.InSlice(softwareVersion, anySlice) &&
		!common.InSlice(c.SoftwareVersion, strings.Split(softwareVersion, ",")) {
		sv = 1
	}
	return ov+pv+sv == 0
}

func MatchTaskTags(cpeTaskTags, configTaskTags string) bool {
	if strings.TrimSpace(configTaskTags) == "" {
		return true
	}
	if len(strings.TrimSpace(cpeTaskTags)) == 0 {
		return false
	}

	for _, tag := range strings.Split(configTaskTags, ",") {
		for _, _tag := range strings.Split(cpeTaskTags, ",") {
			_tag = strings.TrimSpace(_tag)
			if _tag != "" && _tag == tag {
				return true
			}
		}
	}
	return false
}

package app

import (
	"fmt"
	"strings"
	"time"

	"github.com/talkincode/toughradius/v8/common"
	"github.com/talkincode/toughradius/v8/common/cwmp"
	"github.com/talkincode/toughradius/v8/common/zaplog/log"
	"github.com/talkincode/toughradius/v8/models"
	"gopkg.in/yaml.v2"
)

const (
	BootStrapEvent string = "bootstrap"
	BootEvent      string = "boot"
	PeriodicEvent  string = "periodic"
	ScheduledEvent string = "scheduled"
)

func (c *CwmpCpe) GetLatestCwmpPresetTask() (*models.CwmpPresetTask, error) {
	var task models.CwmpPresetTask
	err := app.gormDB.
		Where("sn = ? and status = ? and created_at >= ?", c.Sn, "pending", time.Now().Add(time.Minute*-72)).
		Order("created_at asc").First(&task).Error
	if err != nil {
		return nil, err
	}
	app.gormDB.Model(&task).Update("status", "running")
	return &task, nil
}

func (c *CwmpCpe) MatchDevice(oui, productClass, softwareVersion string) bool {
	anySlice := []string{"", "any", "N/A", "all"}
	var ov, pv, sv int
	if !common.InSlice(oui, anySlice) &&
		!common.InSlice(c.OUI, strings.Split(oui, ",")) {
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

func (c *CwmpCpe) MatchTaskTags(tags string) bool {
	if tags == "" {
		return true
	}
	taskTags := c.TaskTags()
	if len(taskTags) == 0 {
		return false
	}
	for _, tag := range strings.Split(tags, ",") {
		if common.InSlice(tag, taskTags) {
			return true
		}
	}
	return false
}

// CreateCwmpPresetTaskById Manually trigger tasks
func (a *Application) CreateCwmpPresetTaskById(pid string, snlist []string) (err error) {

	if a.cwmpTable.Size() == 0 {
		return fmt.Errorf("no CPE connected")
	}

	var preset models.CwmpPreset
	err = app.gormDB.Where("id = ?", pid).First(&preset).Error
	if err != nil {
		return err
	}

	doPresetTask := func(sns ...string) {
		for _, _sn := range sns {
			c := a.cwmpTable.GetCwmpCpe(_sn)
			err = c.CreateCwmpPresetEventTask(preset.Event, pid)
			if err != nil {
				log.Errorf("CreateCwmpPresetById.CreateCwmpPresetEventTask error: %s", err)
			}
		}
	}

	if len(snlist) > 0 {
		doPresetTask(snlist...)
	} else {
		doPresetTask(a.cwmpTable.ListSn()...)
	}

	return
}

// CreateCwmpPresetEventTask Create scheduled event tasks
func (c *CwmpCpe) CreateCwmpPresetEventTask(event string, pid string) (err error) {
	if event == "" {
		return fmt.Errorf("event is empty")
	}
	var presets []models.CwmpPreset
	query := app.gormDB.Where("event = ?", event)
	if pid != "" {
		query = query.Where("id = ?", pid)
	}
	err = query.Order("priority asc").Find(&presets).Error
	if err != nil {
		return err
	}

	batch := common.UUID()

	for _, preset := range presets {
		if !c.MatchTaskTags(preset.TaskTags) {
			continue
		}
		var content models.CwmpPresetContent
		err = yaml.Unmarshal([]byte(preset.Content), &content)
		if err != nil {
			return err
		}

		// Create a factory settings delivery task
		if content.FactoryResetConfig != nil && content.FactoryResetConfig.Enabled {
			err = c.creatFactoryResetConfigDownloadTask(preset.ID, content.FactoryResetConfig, batch, event)
			if err != nil {
				log.Errorf("creatFactoryResetConfigDownloadTask: %s", err)
			}
		}

		// Create a firmware configuration delivery task
		if content.FirmwareConfig != nil && content.FirmwareConfig.Enabled {
			err = c.creatFirmwareConfigDownloadTask(preset.ID, content.FirmwareConfig, batch, event)
			if err != nil {
				log.Errorf("creatFirmwareConfigDownloadTask: %s", err)
			}
		}

		// Create a script to configure the download task
		if content.Downloads != nil {
			for _, download := range content.Downloads {
				err = c.creatDownloadTask(preset.ID, download, batch, event)
				if err != nil {
					log.Errorf("creatDownloadTask: %s", err)
				}
			}
		}

		// Create an upload task
		if content.Uploads != nil {
			for _, upload := range content.Uploads {
				err = c.creatUploadTask(preset.ID, upload, batch, event)
				if err != nil {
					log.Errorf("creatUploadTask: %s", err)
				}
			}
		}

		go func() {
			// Create parameter task
			if content.SetParameterValues != nil {
				err = c.creatSetParameterValuesTask(preset.ID, content.SetParameterValues, batch, event)
				if err != nil {
					log.Errorf("creatGetParameterValuesTask: %s", err)
				}
			}
			// Create a parameter acquisition task
			if content.GetParameterValues != nil {
				err = c.creatGetParameterValuesTask(content.GetParameterValues)
				if err != nil {
					log.Errorf("creatGetParameterValuesTask: %s", err)
				}
			}

		}()
	}

	return nil
}

// Factory Settings Download Task
func (c *CwmpCpe) creatFactoryResetConfigDownloadTask(presetId int64, freset *models.CwmpPresetFactoryResetConfig, batch, event string) error {
	if !freset.Enabled {
		return nil
	}
	var fconfig models.CwmpFactoryReset
	err := app.gormDB.Where("oid = ?", freset.Oid).First(&fconfig).Error
	if err != nil {
		return err
	}

	if !c.MatchDevice(fconfig.Oui, fconfig.ProductClass, fconfig.SoftwareVersion) {
		return fmt.Errorf("device not match CwmpFactoryResetConfig")
	}

	content := app.InjectCwmpConfigVars(c.Sn, fconfig.Content, map[string]string{
		"CacrtContent": app.GetCacrtContent(),
	})

	session := common.UUID()
	var token = common.Md5Hash(session + app.appConfig.Tr069.Secret + time.Now().Format("20060102"))
	msg := &cwmp.Download{
		ID:         session,
		NoMore:     0,
		CommandKey: session,
		FileType:   "X MIKROTIK Factory Configuration File",
		URL: fmt.Sprintf("%s/cwmpfiles/preset/%s/%s/latest.alter",
			app.GetTr069SettingsStringValue(ConfigTR069AccessAddress), session, token),
		Username:       "",
		Password:       "",
		FileSize:       len([]byte(content)),
		TargetFileName: session + ".alter",
		DelaySeconds:   freset.Delay,
		SuccessURL:     "",
		FailureURL:     "",
	}

	return app.gormDB.Create(&models.CwmpPresetTask{
		ID:        common.UUIDint64(),
		PresetId:  presetId,
		Event:     event,
		Oid:       freset.Oid,
		Name:      msg.GetName(),
		Onfail:    freset.OnFail,
		Batch:     batch,
		Session:   session,
		Sn:        c.Sn,
		Request:   string(msg.CreateXML()),
		Response:  "",
		Content:   content,
		Status:    "pending",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}).Error

}

// Firmware Download Task
func (c *CwmpCpe) creatFirmwareConfigDownloadTask(presetId int64, fconfig *models.CwmpPresetFirmwareConfig, batch, event string) error {
	if !fconfig.Enabled {
		return nil
	}
	var firmwareCfg models.CwmpFirmwareConfig
	err := app.gormDB.Where("oid = ?", fconfig.Oid).First(&firmwareCfg).Error
	if err != nil {
		return err
	}

	if !c.MatchDevice(firmwareCfg.Oui, firmwareCfg.ProductClass, firmwareCfg.SoftwareVersion) {
		return fmt.Errorf("device not match CwmpFirmwareConfig")
	}

	content := app.InjectCwmpConfigVars(c.Sn, firmwareCfg.Content, nil)

	session := common.UUID()
	var token = common.Md5Hash(session + app.appConfig.Tr069.Secret + time.Now().Format("20060102"))
	msg := &cwmp.Download{
		ID:         session,
		Name:       fmt.Sprintf("%s FirmwareConfig Task", c.Sn),
		NoMore:     0,
		CommandKey: session,
		FileType:   "1 Firmware Upgrade Image",
		URL: fmt.Sprintf("%s/cwmpfiles/preset/%s/%s/latest.xml",
			app.GetTr069SettingsStringValue(ConfigTR069AccessAddress), session, token),
		Username:       "",
		Password:       "",
		FileSize:       len([]byte(content)),
		TargetFileName: session + ".xml",
		DelaySeconds:   fconfig.Delay,
		SuccessURL:     "",
		FailureURL:     "",
	}

	return app.gormDB.Create(&models.CwmpPresetTask{
		ID:        common.UUIDint64(),
		PresetId:  presetId,
		Event:     event,
		Oid:       fconfig.Oid,
		Name:      msg.GetName(),
		Onfail:    fconfig.OnFail,
		Batch:     batch,
		Session:   session,
		Sn:        c.Sn,
		Request:   string(msg.CreateXML()),
		Response:  "",
		Content:   content,
		Status:    "pending",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}).Error
}

// 配置脚本下载任务
func (c *CwmpCpe) creatDownloadTask(presetId int64, download models.CwmpPresetDownload, batch, event string) error {
	if !download.Enabled {
		return nil
	}
	var script models.CwmpConfig
	err := app.gormDB.Where("oid = ?", download.Oid).First(&script).Error
	if err != nil {
		return err
	}

	if !c.MatchDevice(script.Oui, script.ProductClass, script.SoftwareVersion) {
		return fmt.Errorf("device not match CwmpConfig")
	}

	if !c.MatchTaskTags(script.TaskTags) {
		return fmt.Errorf("device task_tags  %s not match %s", c.TaskTags(), script.TaskTags)
	}

	// CPE Vars Replace
	var scontent = app.InjectCwmpConfigVars(c.Sn, script.Content, nil)

	session := common.UUID()
	var token = common.Md5Hash(session + app.appConfig.Tr069.Secret + time.Now().Format("20060102"))
	msg := &cwmp.Download{
		ID:         session,
		NoMore:     0,
		CommandKey: session,
		FileType:   "3 Vendor Configuration File",
		URL: fmt.Sprintf("%s/cwmpfiles/preset/%s/%s/latest.alter",
			app.GetTr069SettingsStringValue(ConfigTR069AccessAddress), session, token),
		Username:       "",
		Password:       "",
		FileSize:       len([]byte(scontent)),
		TargetFileName: common.IfEmptyStr(script.TargetFilename, session+".alter"),
		DelaySeconds:   0,
		SuccessURL:     "",
		FailureURL:     "",
	}

	err = app.gormDB.Create(&models.CwmpPresetTask{
		ID:        common.UUIDint64(),
		PresetId:  presetId,
		Event:     event,
		Oid:       script.Oid,
		Name:      msg.GetName(),
		Onfail:    download.OnFail,
		Batch:     batch,
		Session:   session,
		Sn:        c.Sn,
		Request:   string(msg.CreateXML()),
		Response:  "",
		Content:   scontent,
		Status:    "pending",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}).Error
	return err
}

func (c *CwmpCpe) creatUploadTask(presetId int64, upload models.CwmpPresetUpload, batch, event string) error {
	if !upload.Enabled {
		return nil
	}
	session := common.UUID()
	var token = common.Md5Hash(session + app.appConfig.Tr069.Secret + time.Now().Format("20060102"))
	msg := &cwmp.Upload{
		ID:         session,
		NoMore:     0,
		CommandKey: session,
		FileType:   upload.FileType,
		URL: fmt.Sprintf("%s/cwmpupload/%s/%s/%s.rsc",
			app.GetTr069SettingsStringValue(ConfigTR069AccessAddress), session, token, c.Sn+"_"+time.Now().Format("20060102")),
		Username:     "",
		Password:     "",
		DelaySeconds: 0,
	}

	app.gormDB.Create(&models.CwmpPresetTask{
		ID:        common.UUIDint64(),
		PresetId:  presetId,
		Event:     event,
		Oid:       upload.FileType,
		Name:      msg.GetName(),
		Onfail:    upload.OnFail,
		Batch:     batch,
		Session:   session,
		Sn:        c.Sn,
		Request:   string(msg.CreateXML()),
		Response:  "",
		Content:   "",
		Status:    "pending",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})
	return nil
}

// 参数设置任务
func (c *CwmpCpe) creatSetParameterValuesTask(presetId int64, values []models.CwmpPresetParameterValue, batch, event string) error {
	params := map[string]cwmp.ValueStruct{}
	for _, value := range values {
		params[value.Name] = cwmp.ValueStruct{
			Type:  "xsd:" + value.Type,
			Value: value.Value,
		}
	}
	session := "PresetTask-" + common.UUID()
	msg := &cwmp.SetParameterValues{
		ID:     session,
		NoMore: 0,
		Params: params,
	}

	return app.gormDB.Create(&models.CwmpPresetTask{
		ID:        common.UUIDint64(),
		PresetId:  presetId,
		Event:     event,
		Oid:       "N/A",
		Name:      msg.GetName(),
		Onfail:    "ignore",
		Batch:     batch,
		Session:   session,
		Sn:        c.Sn,
		Request:   string(msg.CreateXML()),
		Response:  "",
		Content:   "",
		Status:    "pending",
		ExecTime:  time.Now(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}).Error
}

// 获取参数值任务
func (c *CwmpCpe) creatGetParameterValuesTask(names []string) error {
	session := common.UUID()
	err := c.SendCwmpEventData(models.CwmpEventData{
		Session: session,
		Sn:      c.Sn,
		Message: &cwmp.GetParameterValues{
			ID:             session,
			NoMore:         0,
			ParameterNames: names,
		},
	}, 30000, false)
	return err
}

// UpdateParameterWritable 更新参数权限， 每次更新部分属性
func (c *CwmpCpe) UpdateParameterWritable(session string, timeoutsec int, hp bool) error {
	var params []models.NetCpeParam
	err := app.gormDB.Model(&models.NetCpeParam{}).
		Where("sn = ? and (writable = '' or writable is null) ", c.Sn).
		Find(&params).Error
	if err != nil {
		return err
	}
	var pmap = make(map[string]int, 0)
	for _, param := range params {
		pmap[getParentPath(param.Name)] = 1
	}
	for paramPath, _ := range pmap {
		err = c.SendCwmpEventData(models.CwmpEventData{
			Session: common.UUID(),
			Sn:      c.Sn,
			Message: &cwmp.GetParameterNames{
				ID:            session,
				NoMore:        0,
				NextLevel:     "true",
				ParameterPath: paramPath,
			},
		}, timeoutsec, hp)
		if err != nil {
			log.Error(err)
		}
	}
	return nil
}

// 参数格式为x.y.z 返回格式 x.y.
func getParentPath(path string) string {
	if strings.Contains(path, ".") {
		return path[:strings.LastIndex(path, ".")+1]
	}
	return path
}

// UpdateCwmpPresetTaskStatus Update cwmp preset task status
func UpdateCwmpPresetTaskStatus(msg cwmp.Message) (err error) {
	switch msg.GetName() {
	case "TransferComplete":
		tc := msg.(*cwmp.TransferComplete)
		var task models.CwmpPresetTask
		err = app.gormDB.Where("session = ?", tc.CommandKey).First(&task).Error
		if err != nil {
			return
		}

		err = app.gormDB.Model(&models.CwmpPresetTask{}).Where("session = ?", tc.CommandKey).Updates(map[string]interface{}{
			"status":    common.If(tc.FaultCode == 0, "success", "failure"),
			"response":  string(tc.CreateXML()),
			"exec_time": tc.StartTime,
			"resp_time": tc.CompleteTime,
		}).Error

		if tc.FaultCode > 0 && task.Batch != "" && task.Onfail == "cancel" {
			err = app.gormDB.Model(&models.CwmpPresetTask{}).
				Where("status = ? and batch = ?", "pending", task.Batch).Updates(map[string]interface{}{
				"status": "cancel",
			}).Error
			return
		}

	case "SetParameterValuesResponse":
		sm := msg.(*cwmp.SetParameterValuesResponse)
		if strings.HasPrefix(msg.GetID(), "PresetTask") {
			// 尝试更新预设任务状态
			_ = app.gormDB.Model(&models.CwmpPresetTask{}).Where("session = ?", sm.GetID()).Updates(map[string]interface{}{
				"status":    common.If(sm.Status == 0, "success", "failure"),
				"response":  string(sm.CreateXML()),
				"resp_time": time.Now(),
			})
		}
	}
	return
}

// UpdateCwmpConfigSessionStatus Update script task status
func UpdateCwmpConfigSessionStatus(tc *cwmp.TransferComplete) (err error) {
	err = app.gormDB.Model(&models.CwmpConfigSession{}).Where("session = ?", tc.CommandKey).Updates(map[string]interface{}{
		"exec_status": common.If(tc.FaultCode == 0, "success", "failure"),
		"last_error":  common.If(tc.FaultCode == 0, "Cwmp Vendor Configuration File transfer complete", tc.FaultString),
		"exec_time":   tc.StartTime,
		"resp_time":   tc.CompleteTime,
	}).Error
	return
}

// CancelBatchCwmpPresetTask 取消批次任务
// func CancelBatchCwmpPresetTask(session string) (err error) {
// 	var batch string
// 	err = app.gormDB.Model(&models.CwmpPresetTask{}).
// 		Where("session = ? and onfail = ?", session, "cancel").
// 		Pluck("batch", &batch).Error
// 	if err != nil {
// 		return
// 	}
// 	err = app.gormDB.Model(&models.CwmpPresetTask{}).
// 		Where("status = ? and batch = ?", "pending", batch).
// 		Updates(map[string]interface{}{
// 		"status": "cancel",
// 	}).Error
// 	return
// }

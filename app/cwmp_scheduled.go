package app

import (
	"fmt"
	"time"

	"github.com/talkincode/toughradius/common"
	"github.com/talkincode/toughradius/common/cwmp"
	"github.com/talkincode/toughradius/common/zaplog/log"
	"github.com/talkincode/toughradius/models"
	"gopkg.in/yaml.v2"
)

// CreateCwmpSchedEventTask Cwmp 定时任务， 由 CPE 触发
func (c *CwmpCpe) CreateCwmpSchedEventTask(schedKey string) (err error) {
	if schedKey == "" {
		return fmt.Errorf("schedKey is empty")
	}
	var presets []models.CwmpPreset
	err = app.gormDB.
		Where("event = ? and sched_key = ?", ScheduledEvent, schedKey).
		Order("priority asc").Find(&presets).Error
	if err != nil {
		return err
	}

	batch := common.UUID()

	for _, preset := range presets {
		if !c.MatchTaskTags(preset.TaskTags) {
			continue
		}
		var content models.CwmpPresetSched
		err = yaml.Unmarshal([]byte(preset.Content), &content)
		if err != nil {
			return err
		}

		simsg := &cwmp.ScheduleInform{
			ID:           common.UUID(),
			Name:         fmt.Sprintf("%s ScheduleInform", c.Sn),
			NoMore:       0,
			CommandKey:   schedKey,
			DelaySeconds: preset.Interval,
		}
		err = c.SendCwmpEventData(models.CwmpEventData{
			Session: schedKey,
			Sn:      c.Sn,
			Message: simsg,
		}, 30000, true)
		if err != nil {
			return err
		}

		// 创建脚本配置下载任务
		if content.Downloads != nil {
			for _, download := range content.Downloads {
				err = c.creatDownloadTask(preset.ID, download, batch, "schedule-"+schedKey)
				if err != nil {
					log.Errorf("creatDownloadTask: %s", err)
				}
			}
		}

		// 创建上传任务
		if content.Uploads != nil {
			for _, upload := range content.Uploads {
				err = c.creatUploadTask(preset.ID, upload, batch, "schedule-"+schedKey)
				if err != nil {
					log.Errorf("creatUploadTask: %s", err)
				}
			}
		}

		go func() {
			time.Sleep(1 * time.Second)

			// 创建参数任务
			if content.SetParameterValues != nil {
				err = c.creatSetParameterValuesTask(preset.ID, content.SetParameterValues, batch, "schedule-"+schedKey)
				if err != nil {
					log.Errorf("creatSetParameterValuesTask: %s", err)
				}
			}
			// 创建参数获取任务
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

func (c *CwmpCpe) ActiveCwmpSchedEventTask() (err error) {
	var presets []models.CwmpPreset
	query := app.gormDB.Where("event = ? and interval > 0", ScheduledEvent)
	err = query.Find(&presets).Error
	if err != nil {
		return err
	}

	for _, preset := range presets {
		if !c.MatchTaskTags(preset.TaskTags) {
			continue
		}
		var content models.CwmpPresetSched
		err = yaml.Unmarshal([]byte(preset.Content), &content)
		if err != nil {
			return err
		}

		simsg := &cwmp.ScheduleInform{
			ID:           common.UUID(),
			Name:         fmt.Sprintf("%s ScheduleInform", c.Sn),
			NoMore:       0,
			CommandKey:   preset.SchedKey,
			DelaySeconds: preset.Interval,
		}
		err = c.SendCwmpEventData(models.CwmpEventData{
			Session: preset.SchedKey,
			Sn:      c.Sn,
			Message: simsg,
		}, 30000, true)
		if err != nil {
			log.Errorf("SendCwmpEventData: %s", err)
		}
	}

	return nil
}

// CreateCwmpScheduledTask 系统 定时任务， 由系统触发
func CreateCwmpScheduledTask(schedPolicy string) (err error) {

	if app.cwmpTable.Size() == 0 {
		return fmt.Errorf("no CPE connected")
	}

	if schedPolicy == "" {
		return fmt.Errorf("schedPolicy is empty")
	}
	var schedPresets []models.CwmpPreset
	err = app.gormDB.
		Where("event = 'sys_scheduled' and sched_policy = ?", schedPolicy).
		Order("priority asc").Find(&schedPresets).Error
	if err != nil {
		return err
	}

	for _, sn := range app.cwmpTable.ListSn() {
		c := app.cwmpTable.GetCwmpCpe(sn)
		batch := common.UUID()

		for _, preset := range schedPresets {
			if !c.MatchTaskTags(preset.TaskTags) {
				continue
			}
			var content models.CwmpPresetSched
			err = yaml.Unmarshal([]byte(preset.Content), &content)
			if err != nil {
				return err
			}

			// 创建脚本配置下载任务
			if content.Downloads != nil {
				for _, download := range content.Downloads {
					err = c.creatDownloadTask(preset.ID, download, batch, "sys_scheduled")
					if err != nil {
						log.Errorf("creatDownloadTask: %s", err)
					}
				}
			}

			// 创建上传任务
			if content.Uploads != nil {
				for _, upload := range content.Uploads {
					err = c.creatUploadTask(preset.ID, upload, batch, "sys_scheduled")
					if err != nil {
						log.Errorf("creatUploadTask: %s", err)
					}
				}
			}

			_preset := preset
			go func() {
				time.Sleep(1 * time.Second)

				// 创建参数任务
				if content.SetParameterValues != nil {
					err = c.creatSetParameterValuesTask(_preset.ID, content.SetParameterValues, batch, "sys_scheduled")
					if err != nil {
						log.Errorf("creatSetParameterValuesTask: %s", err)
					}
				}
				// 创建参数获取任务
				if content.GetParameterValues != nil {
					err = c.creatGetParameterValuesTask(content.GetParameterValues)
					if err != nil {
						log.Errorf("creatGetParameterValuesTask: %s", err)
					}
				}

			}()
		}
	}
	return nil
}

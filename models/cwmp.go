package models

import (
	"time"

	"github.com/talkincode/toughradius/common/cwmp"
)

type CwmpConfig struct {
	ID              string    `gorm:"primaryKey" json:"id" form:"id"` // 主键 ID
	Oid             string    `gorm:"index" json:"oid" form:"oid"`
	Name            string    `json:"name" form:"name"`   // Name
	Level           string    `json:"level" form:"level"` // script level  normal｜major
	SoftwareVersion string    `json:"software_version" form:"software_version"`
	ProductClass    string    `json:"product_class" form:"product_class"`
	Oui             string    `json:"oui" form:"oui"`
	TaskTags        string    `gorm:"index" json:"task_tags" form:"task_tags"` // task label
	Content         string    `json:"content" form:"content"`                  // script content
	TargetFilename  string    `json:"target_filename" form:"target_filename"`
	Timeout         int64     `json:"timeout" form:"timeout"` // Execution Timeout Seconds
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type CwmpConfigSession struct {
	ID              int64     `gorm:"primaryKey" json:"id,string" form:"id"`          // primary key ID
	ConfigId        string    `gorm:"index" json:"config_id,string" form:"config_id"` // Script ID
	CpeId           int64     `gorm:"index" json:"cpe_id,string" form:"cpe_id"`       // CPE ID
	Session         string    `gorm:"index" json:"session" form:"session"`            // Session ID
	Name            string    `json:"name" form:"name"`                               // Name
	Level           string    `json:"level" form:"level"`                             // script level  normal｜major
	SoftwareVersion string    `json:"software_version" form:"software_version"`
	ProductClass    string    `json:"product_class" form:"product_class"`
	Oui             string    `json:"oui" form:"oui"`
	TaskTags        string    `gorm:"index" json:"task_tags" form:"task_tags"`     // task label
	Content         string    `json:"content" form:"content"`                      // script content
	ExecStatus      string    `gorm:"index" json:"exec_status" form:"exec_status"` // execution state  success | failure
	LastError       string    `json:"last_error" form:"last_error"`                // last execution error
	Timeout         int64     `json:"timeout" form:"timeout"`                      // execution timeout second
	ExecTime        time.Time `gorm:"index" json:"exec_time" form:"exec_time"`     // execution time
	RespTime        time.Time `json:"resp_time" form:"resp_time"`                  // Response time
	CreatedAt       time.Time `gorm:"index" json:"created_at"`
	UpdatedAt       time.Time `gorm:"index" json:"updated_at"`
}

type CwmpEventData struct {
	Session string       `json:"session"`
	Sn      string       `json:"sn"`
	Message cwmp.Message `json:"message"`
}

// CwmpFactoryReset factory settings script
type CwmpFactoryReset struct {
	ID              int64     `json:"id,string" form:"id"` // 主键 ID
	Oid             string    `json:"oid" form:"oid"`
	Name            string    `json:"name" form:"name"`
	SoftwareVersion string    `json:"software_version" form:"software_version"`
	ProductClass    string    `json:"product_class" form:"product_class"`
	Oui             string    `json:"oui" form:"oui"`
	Content         string    `json:"content" form:"content"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// CwmpFirmwareConfig firmware configuration
type CwmpFirmwareConfig struct {
	ID              int64     `json:"id,string" form:"id"` // 主键 ID
	Oid             string    `json:"oid" form:"oid"`
	Name            string    `json:"name" form:"name"`
	SoftwareVersion string    `json:"software_version" form:"software_version"`
	ProductClass    string    `json:"product_class" form:"product_class"`
	Oui             string    `json:"oui" form:"oui"`
	Content         string    `json:"content" form:"content"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type CwmpPreset struct {
	ID          int64     `json:"id,string" form:"id"` // 主键 ID
	Name        string    `json:"name" form:"name"`
	Priority    int       `json:"priority" form:"priority"`
	Event       string    `json:"event" form:"event"`
	SchedPolicy string    `json:"sched_policy" form:"sched_policy"`
	SchedKey    string    `json:"sched_key" form:"sched_key"`
	Interval    int       `json:"interval" form:"interval"`
	Content     string    `json:"content" form:"content"`
	TaskTags    string    `gorm:"index" json:"task_tags" form:"task_tags"` // 任务标签
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CwmpPresetTask struct {
	ID        int64     `json:"id,string"` // 主键 ID
	PresetId  int64     `json:"preset_id,string" gorm:"index"`
	Sn        string    `json:"sn" gorm:"index"`
	Name      string    `json:"name" `
	Oid       string    `json:"oid" gorm:"index"`
	Event     string    `json:"event" `
	Batch     string    `json:"batch" `
	Onfail    string    `json:"onfail" `
	Session   string    `json:"session" gorm:"index"`
	Request   string    `json:"request" `
	Response  string    `json:"response"`
	Content   string    `json:"content"`
	Status    string    `json:"status" gorm:"index"`
	ExecTime  time.Time `json:"exec_time"` // 执行时间
	RespTime  time.Time `json:"resp_time"` // 响应时间
	CreatedAt time.Time `json:"created_at" gorm:"index"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Preset

type CwmpPresetSched struct {
	Downloads          []CwmpPresetDownload       `yaml:"Downloads"`
	Uploads            []CwmpPresetUpload         `yaml:"Uploads"`
	GetParameterValues []string                   `yaml:"GetParameterValues"`
	SetParameterValues []CwmpPresetParameterValue `yaml:"SetParameterValues"`
}

type CwmpPresetContent struct {
	FactoryResetConfig *CwmpPresetFactoryResetConfig `yaml:"FactoryResetConfig"`
	FirmwareConfig     *CwmpPresetFirmwareConfig     `yaml:"FirmwareConfig"`
	Downloads          []CwmpPresetDownload          `yaml:"Downloads"`
	Uploads            []CwmpPresetUpload            `yaml:"Uploads"`
	GetParameterValues []string                      `yaml:"GetParameterValues"`
	SetParameterValues []CwmpPresetParameterValue    `yaml:"SetParameterValues"`
}

type CwmpPresetDownload struct {
	Oid     string `yaml:"oid"`
	Enabled bool   `yaml:"enabled"`
	Delay   int    `yaml:"delay"`
	OnFail  string `yaml:"onfail"`
}

type CwmpPresetUpload struct {
	FileType string `yaml:"filetype"`
	Enabled  bool   `yaml:"enabled"`
	OnFail   string `yaml:"onfail"`
}

type CwmpPresetFactoryResetConfig struct {
	Oid     string `yaml:"oid"`
	Enabled bool   `yaml:"enabled"`
	Delay   int    `yaml:"delay"`
	OnFail  string `yaml:"onfail"`
}

type CwmpPresetFirmwareConfig struct {
	Oid     string `yaml:"oid"`
	Enabled bool   `yaml:"enabled"`
	Delay   int    `yaml:"delay"`
	OnFail  string `yaml:"onfail"`
}

type CwmpPresetParameterValue struct {
	Name  string `yaml:"name"`
	Type  string `yaml:"type"`
	Value string `yaml:"value"`
}

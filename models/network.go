package models

import "time"

// 网络模块相关模型

// NetNode network node
type NetNode struct {
	ID        int64     `json:"id,string" form:"id"`
	Name      string    `json:"name" form:"name"`
	Remark    string    `json:"remark" form:"remark"`
	Tags      string    `json:"tags" form:"tags"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NetVpe VPE data model, VPE is usually a gateway-type device that can act as a BRAS device
type NetVpe struct {
	ID         int64     `json:"id,string" form:"id"`            // 主键 ID
	NodeId     int64     `json:"node_id,string" form:"node_id"`  // 节点ID
	LdapId     int64     `json:"ldap_id,string" form:"ldap_id"`  // LDAP ID
	Name       string    `json:"name" form:"name"`               // 设备名称
	Identifier string    `json:"identifier" form:"identifier"`   // 设备标识-RADIUS
	Hostname   string    `json:"hostname" form:"hostname"`       // 设备主机地址
	Ipaddr     string    `json:"ipaddr" form:"ipaddr"`           // 设备IP
	Secret     string    `json:"secret" form:"secret"`           // 设备 RADIUS 秘钥
	CoaPort    int       `json:"coa_port" form:"coa_port"`       // 设备 RADIUS COA 端口
	Model      string    `json:"model" form:"model"`             // 设备型号
	VendorCode string    `json:"vendor_code" form:"vendor_code"` // 设备厂商代码
	Status     string    `json:"status" form:"status"`           // 设备状态
	Tags       string    `json:"tags" form:"tags"`               // 标签
	Remark     string    `json:"remark" form:"remark"`           // 备注
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// NetCpe Cpe 数据模型
type NetCpe struct {
	ID              int64     `json:"id,string" form:"id"`                      // 主键 ID
	NodeId          int64     `json:"node_id,string" form:"node_id"`            // 节点ID
	SystemName      string    `json:"system_name" form:"system_name"`           // 设备系统名称
	CpeType         string    `json:"cpe_type" form:"cpe_type"`                 // 设备类型
	Sn              string    `gorm:"uniqueIndex" json:"sn" form:"sn"`          // 设备序列号
	Name            string    `json:"name" form:"name"`                         // 设备名称
	ArchName        string    `json:"arch_name" form:"arch_name"`               // 设备架构
	SoftwareVersion string    `json:"software_version" form:"software_version"` // 设备固件版本
	HardwareVersion string    `json:"hardware_version" form:"hardware_version"` // 设备版本
	Model           string    `json:"model" form:"model"`                       // 设备型号
	VendorCode      string    `json:"vendor_code" form:"vendor_code"`           // 设备厂商代码
	Oui             string    `json:"oui" form:"oui"`                           // 设备OUI
	Manufacturer    string    `json:"manufacturer" form:"manufacturer"`         // 设备制造商
	ProductClass    string    `json:"product_class" form:"product_class"`       // 设备类型
	Status          string    `gorm:"index" json:"status" form:"status"`        // 设备状态
	Tags            string    `json:"tags" form:"tags"`                         // 标签
	TaskTags        string    `gorm:"index" json:"task_tags" form:"task_tags"`  // 任务标签
	Remark          string    `json:"remark" form:"remark"`                     // 备注
	Uptime          int64     `json:"uptime" form:"uptime"`                     // UpTime
	MemoryTotal     int64     `json:"memory_total" form:"memory_total"`         // 内存总量
	MemoryFree      int64     `json:"memory_free" form:"memory_free"`           // 内存可用
	CPUUsage        int64     `json:"cpu_usage" form:"cpu_usage"`               // CPE 百分比
	CwmpStatus      string    `gorm:"index"  json:"cwmp_status"`                // cwmp 状态
	CwmpUrl         string    `json:"cwmp_url"`
	FactoryresetId  string    `json:"factoryreset_id" form:"factoryreset_id"`
	CwmpLastInform  time.Time `json:"cwmp_last_inform" ` // CWMP 最后检测时间
	CreatedAt       time.Time `json:"created_at" `
	UpdatedAt       time.Time `json:"updated_at"`
}

// NetCpeParam CPE 参数
type NetCpeParam struct {
	ID        string    `gorm:"primaryKey" json:"string"` // 主键 ID
	Sn        string    `gorm:"index" json:"sn"`          // 设备序列号
	Tag       string    `gorm:"index" json:"tag" `
	Name      string    `gorm:"index" json:"name" `
	Value     string    `json:"value" `
	Remark    string    `json:"remark"`
	Writable  string    `json:"writable"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type NetCpeTaskQue struct {
	ID     int64  `json:"id,string"` // 主键 ID
	Sn     string `json:"sn"`        // 设备序列号
	TaskId string `json:"task_id"`
}

type NetLdapServer struct {
	Id         int64     `json:"id,string" form:"id"`
	Tags       string    `json:"tags" form:"tags"`
	Name       string    `json:"name" form:"name"`
	Address    string    `json:"address" form:"address"`
	Password   string    `json:"password" form:"password"`
	Searchdn   string    `json:"searchdn" form:"searchdn"`
	Basedn     string    `json:"basedn" form:"basedn"`
	UserFilter string    `json:"user_filter" form:"user_filter"`
	Istls      string    `json:"istls" form:"istls"`
	Status     string    `json:"status" form:"status"`
	Remark     string    `json:"remark" form:"remark"`
	CreateTime time.Time `json:"create_time,string" `
	UpdateTime time.Time `json:"update_time,string" `
}

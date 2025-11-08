package domain

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

// NetNas NAS 设备数据模型，通常是网关类型的设备，可作为 BRAS 设备
type NetNas struct {
	ID         int64     `json:"id,string" form:"id"`            // 主键 ID
	NodeId     int64     `json:"node_id,string" form:"node_id"`  // 节点ID
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

// TableName 指定表名
func (NetNas) TableName() string {
	return "net_vpe"
}

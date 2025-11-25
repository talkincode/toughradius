package domain

import "time"

// Network module related models

// NetNode network node
type NetNode struct {
	ID        int64     `json:"id,string" form:"id"`
	Name      string    `json:"name" form:"name"`
	Remark    string    `json:"remark" form:"remark"`
	Tags      string    `json:"tags" form:"tags"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName Specify table name
func (NetNode) TableName() string {
	return "net_node"
}

// NetNas NAS device data model, typically gateway-type devices, can be used as BRAS equipment
type NetNas struct {
	ID         int64     `json:"id,string" form:"id"`            // Primary key ID
	NodeId     int64     `json:"node_id,string" form:"node_id"`  // Node ID
	Name       string    `json:"name" form:"name"`               // Device name
	Identifier string    `json:"identifier" form:"identifier"`   // Device identifier - RADIUS
	Hostname   string    `json:"hostname" form:"hostname"`       // Device host address
	Ipaddr     string    `json:"ipaddr" form:"ipaddr"`           // Device IP
	Secret     string    `json:"secret" form:"secret"`           // Device RADIUS Secret
	CoaPort    int       `json:"coa_port" form:"coa_port"`       // Device RADIUS COA Port
	Model      string    `json:"model" form:"model"`             // Device model
	VendorCode string    `json:"vendor_code" form:"vendor_code"` // Device vendor code
	Status     string    `json:"status" form:"status"`           // Device status
	Tags       string    `json:"tags" form:"tags"`               // Tags
	Remark     string    `json:"remark" form:"remark"`           // Remark
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// TableName Specify table name
func (NetNas) TableName() string {
	return "net_nas"
}

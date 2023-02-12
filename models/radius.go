package models

import (
	"time"
)

// RADIUS 相关模型

// RadiusProfile RADIUS 策略
type RadiusProfile struct {
	ID        int64     `json:"id,string" form:"id"`               // 主键 ID
	NodeId    int64     `json:"node_id,string" form:"node_id"`     // 节点ID
	Name      string    `json:"name" form:"name"`                  // 策略名称
	Status    string    `gorm:"index" json:"status" form:"status"` // 策略状态 0：禁用 1：正常
	AddrPool  string    `json:"addr_pool" form:"addr_pool"`        // 策略地址池
	ActiveNum int       `json:"active_num" form:"active_num"`      // 并发数
	UpRate    int       `json:"up_rate" form:"up_rate"`            // 上行速率
	DownRate  int       `json:"down_rate" form:"down_rate"`        // 下行速率
	Remark    string    `json:"remark" form:"remark"`              // 备注
	CreatedAt time.Time `json:"created_at" form:"created_at"`
	UpdatedAt time.Time `json:"updated_at" form:"updated_at"`
}

// RadiusUser RADIUS Authentication account
type RadiusUser struct {
	ID          int64     `json:"id,string" form:"id"`                              // 主键 ID
	NodeId      int64     `json:"node_id,string" form:"node_id"`                    // 节点ID
	ProfileId   int64     `gorm:"index" json:"profile_id,string" form:"profile_id"` // RADIUS 策略ID
	Realname    string    `json:"realname" form:"realname"`                         // 联系人姓名
	Mobile      string    `json:"mobile" form:"mobile"`                             // 联系人电话
	Username    string    `json:"username" gorm:"uniqueIndex" form:"username"`      // 账号名
	Password    string    `json:"password" form:"password"`                         // 密码
	AddrPool    string    `json:"addr_pool" form:"addr_pool"`                       // 策略地址池
	ActiveNum   int       `gorm:"index" json:"active_num" form:"active_num"`        // 并发数
	UpRate      int       `json:"up_rate" form:"up_rate"`                           // 上行速率
	DownRate    int       `json:"down_rate" form:"down_rate"`                       // 下行速率
	Vlanid1     int       `json:"vlanid1" form:"vlanid1"`                           // VLAN ID 1
	Vlanid2     int       `json:"vlanid2" form:"vlanid2"`                           // VLAN ID 2
	IpAddr      string    `json:"ip_addr" form:"ip_addr"`                           // 静态IP
	MacAddr     string    `json:"mac_addr" form:"mac_addr"`                         // MAC
	BindVlan    int       `json:"bind_vlan" form:"bind_vlan"`                       // 绑定VLAN
	BindMac     int       `json:"bind_mac" form:"bind_mac"`                         // 绑定MAC
	ExpireTime  time.Time `gorm:"index" json:"expire_time"`                         // 过期时间
	Status      string    `gorm:"index" json:"status" form:"status"`                // 状态 enabled | disabled
	Remark      string    `json:"remark" form:"remark"`                             // 备注
	OnlineCount int       `json:"online_count" gorm:"-:migration;<-:false"`
	LastOnline  time.Time `json:"last_online"`
	CreatedAt   time.Time `gorm:"index" json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// RadiusOnline
// Radius RadiusOnline Recode
type RadiusOnline struct {
	ID                int64     `json:"id,string"` // 主键 ID
	Username          string    `gorm:"index" json:"username"`
	NasId             string    `json:"nas_id"`
	NasAddr           string    `json:"nas_addr"`
	NasPaddr          string    `json:"nas_paddr"`
	SessionTimeout    int       `json:"session_timeout"`
	FramedIpaddr      string    `json:"framed_ipaddr"`
	FramedNetmask     string    `json:"framed_netmask"`
	MacAddr           string    `json:"mac_addr"`
	NasPort           int64     `json:"nas_port,string"`
	NasClass          string    `json:"nas_class"`
	NasPortId         string    `json:"nas_port_id"`
	NasPortType       int       `json:"nas_port_type"`
	ServiceType       int       `json:"service_type"`
	AcctSessionId     string    `gorm:"index" json:"acct_session_id"`
	AcctSessionTime   int       `json:"acct_session_time"`
	AcctInputTotal    int64     `json:"acct_input_total,string"`
	AcctOutputTotal   int64     `json:"acct_output_total,string"`
	AcctInputPackets  int       `json:"acct_input_packets"`
	AcctOutputPackets int       `json:"acct_output_packets"`
	AcctStartTime     time.Time `gorm:"index" json:"acct_start_time"`
	LastUpdate        time.Time `json:"last_update"`
}

// RadiusAccounting
// Radius Accounting Recode
type RadiusAccounting struct {
	ID                int64     `json:"id,string"` // 主键 ID
	Username          string    `gorm:"index" json:"username"`
	AcctSessionId     string    `gorm:"index" json:"acct_session_id"`
	NasId             string    `json:"nas_id"`
	NasAddr           string    `json:"nas_addr"`
	NasPaddr          string    `json:"nas_paddr"`
	SessionTimeout    int       `json:"session_timeout"`
	FramedIpaddr      string    `json:"framed_ipaddr"`
	FramedNetmask     string    `json:"framed_netmask"`
	MacAddr           string    `json:"mac_addr"`
	NasPort           int64     `json:"nas_port,string"`
	NasClass          string    `json:"nas_class"`
	NasPortId         string    `json:"nas_port_id"`
	NasPortType       int       `json:"nas_port_type"`
	ServiceType       int       `json:"service_type"`
	AcctSessionTime   int       `json:"acct_session_time"`
	AcctInputTotal    int64     `json:"acct_input_total,string"`
	AcctOutputTotal   int64     `json:"acct_output_total,string"`
	AcctInputPackets  int       `json:"acct_input_packets"`
	AcctOutputPackets int       `json:"acct_output_packets"`
	LastUpdate        time.Time `json:"last_update"`
	AcctStartTime     time.Time `gorm:"index" json:"acct_start_time"`
	AcctStopTime      time.Time `gorm:"index" json:"acct_stop_time"`
}

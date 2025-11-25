package domain

import (
	"time"
)

// RADIUS related models

// RadiusProfile RADIUS billing profile
type RadiusProfile struct {
	ID             int64     `json:"id,string" form:"id"`                      // Primary key ID
	NodeId         int64     `json:"node_id,string" form:"node_id"`            // Node ID
	Name           string    `json:"name" form:"name"`                         // Profile name
	Status         string    `gorm:"index" json:"status" form:"status"`        // Profile status: 0=disabled 1=enabled
	AddrPool       string    `json:"addr_pool" form:"addr_pool"`               // Address pool
	ActiveNum      int       `json:"active_num" form:"active_num"`             // Concurrent sessions
	UpRate         int       `json:"up_rate" form:"up_rate"`                   // Upload rate in Kb
	DownRate       int       `json:"down_rate" form:"down_rate"`               // Download rate in Kb
	Domain         string    `json:"domain" form:"domain"`                     // Domain, corresponds to NAS device domain attribute, e.g., Huawei domain_code
	IPv6PrefixPool string    `json:"ipv6_prefix_pool" form:"ipv6_prefix_pool"` // IPv6 prefix pool name for NAS-side allocation
	BindMac        int       `json:"bind_mac" form:"bind_mac"`                 // Bind MAC
	BindVlan       int       `json:"bind_vlan" form:"bind_vlan"`               // Bind VLAN
	Remark         string    `json:"remark" form:"remark"`                     // Remark
	CreatedAt      time.Time `json:"created_at" form:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" form:"updated_at"`
}

// TableName Specify table name
func (RadiusProfile) TableName() string {
	return "radius_profile"
}

// RadiusUser RADIUS Authentication account
type RadiusUser struct {
	ID              int64     `json:"id,string" form:"id"`                              // Primary key ID
	NodeId          int64     `json:"node_id,string" form:"node_id"`                    // Node ID
	ProfileId       int64     `gorm:"index" json:"profile_id,string" form:"profile_id"` // RADIUS profile ID
	Realname        string    `json:"realname" form:"realname"`                         // Contact name
	Mobile          string    `json:"mobile" form:"mobile"`                             // Contact phone
	Username        string    `json:"username" gorm:"uniqueIndex" form:"username"`      // Account name
	Password        string    `json:"password" form:"password"`                         // Password
	AddrPool        string    `json:"addr_pool" form:"addr_pool"`                       // Address pool
	ActiveNum       int       `gorm:"index" json:"active_num" form:"active_num"`        // Concurrent sessions
	UpRate          int       `json:"up_rate" form:"up_rate"`                           // Upload rate
	DownRate        int       `json:"down_rate" form:"down_rate"`                       // Download rate
	Vlanid1         int       `json:"vlanid1" form:"vlanid1"`                           // VLAN ID 1
	Vlanid2         int       `json:"vlanid2" form:"vlanid2"`                           // VLAN ID 2
	IpAddr          string    `json:"ip_addr" form:"ip_addr"`                           // Static IP
	IpV6Addr        string    `json:"ipv6_addr" form:"ipv6_addr"`                       // Static IPv6 address
	MacAddr         string    `json:"mac_addr" form:"mac_addr"`                         // MAC address
	Domain          string    `json:"domain" form:"domain"`                             // Domain name for vendor-specific features (e.g., Huawei domain)
	IPv6PrefixPool  string    `json:"ipv6_prefix_pool" form:"ipv6_prefix_pool"`         // IPv6 prefix pool name (inherited from profile or user-specific)
	BindVlan        int       `json:"bind_vlan" form:"bind_vlan"`                       // Bind VLAN
	BindMac         int       `json:"bind_mac" form:"bind_mac"`                         // Bind MAC
	ProfileLinkMode int       `json:"profile_link_mode" form:"profile_link_mode"`       // 0=static (snapshot), 1=dynamic (real-time from profile)
	ExpireTime      time.Time `gorm:"index" json:"expire_time"`                         // Expiration time
	Status          string    `gorm:"index" json:"status" form:"status"`                // Status: enabled | disabled
	Remark          string    `json:"remark" form:"remark"`                             // Remark
	OnlineCount     int       `json:"online_count" gorm:"-:migration;<-:false"`
	LastOnline      time.Time `json:"last_online"`
	CreatedAt       time.Time `gorm:"index" json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// TableName Specify table name
func (RadiusUser) TableName() string {
	return "radius_user"
}

// RadiusOnline
// Radius RadiusOnline Recode
type RadiusOnline struct {
	ID                  int64     `json:"id,string"` // Primary key ID
	Username            string    `gorm:"index" json:"username"`
	NasId               string    `json:"nas_id"`
	NasAddr             string    `json:"nas_addr"`
	NasPaddr            string    `json:"nas_paddr"`
	SessionTimeout      int       `json:"session_timeout"`
	FramedIpaddr        string    `json:"framed_ipaddr"`
	FramedNetmask       string    `json:"framed_netmask"`
	FramedIpv6Prefix    string    `json:"framed_ipv6_prefix"`
	FramedIpv6Address   string    `json:"framed_ipv6_address"`
	DelegatedIpv6Prefix string    `json:"delegated_ipv6_prefix"`
	MacAddr             string    `json:"mac_addr"`
	NasPort             int64     `json:"nas_port,string"`
	NasClass            string    `json:"nas_class"`
	NasPortId           string    `json:"nas_port_id"`
	NasPortType         int       `json:"nas_port_type"`
	ServiceType         int       `json:"service_type"`
	AcctSessionId       string    `gorm:"index" json:"acct_session_id"`
	AcctSessionTime     int       `json:"acct_session_time"`
	AcctInputTotal      int64     `json:"acct_input_total,string"`
	AcctOutputTotal     int64     `json:"acct_output_total,string"`
	AcctInputPackets    int       `json:"acct_input_packets"`
	AcctOutputPackets   int       `json:"acct_output_packets"`
	AcctStartTime       time.Time `gorm:"index" json:"acct_start_time"`
	LastUpdate          time.Time `json:"last_update"`
}

// TableName Specify table name
func (RadiusOnline) TableName() string {
	return "radius_online"
}

// RadiusAccounting
// Radius Accounting Recode
type RadiusAccounting struct {
	ID                  int64     `json:"id,string"` // Primary key ID
	Username            string    `gorm:"index" json:"username"`
	AcctSessionId       string    `gorm:"index" json:"acct_session_id"`
	NasId               string    `json:"nas_id"`
	NasAddr             string    `json:"nas_addr"`
	NasPaddr            string    `json:"nas_paddr"`
	SessionTimeout      int       `json:"session_timeout"`
	FramedIpaddr        string    `json:"framed_ipaddr"`
	FramedNetmask       string    `json:"framed_netmask"`
	FramedIpv6Prefix    string    `json:"framed_ipv6_prefix"`
	FramedIpv6Address   string    `json:"framed_ipv6_address"`
	DelegatedIpv6Prefix string    `json:"delegated_ipv6_prefix"`
	MacAddr             string    `json:"mac_addr"`
	NasPort             int64     `json:"nas_port,string"`
	NasClass            string    `json:"nas_class"`
	NasPortId           string    `json:"nas_port_id"`
	NasPortType         int       `json:"nas_port_type"`
	ServiceType         int       `json:"service_type"`
	AcctSessionTime     int       `json:"acct_session_time"`
	AcctInputTotal      int64     `json:"acct_input_total,string"`
	AcctOutputTotal     int64     `json:"acct_output_total,string"`
	AcctInputPackets    int       `json:"acct_input_packets"`
	AcctOutputPackets   int       `json:"acct_output_packets"`
	LastUpdate          time.Time `json:"last_update"`
	AcctStartTime       time.Time `gorm:"index" json:"acct_start_time"`
	AcctStopTime        time.Time `gorm:"index" json:"acct_stop_time"`
}

// TableName Specify table name
func (RadiusAccounting) TableName() string {
	return "radius_accounting"
}

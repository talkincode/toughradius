package domain

import (
	"time"
)

type SysConfig struct {
	ID        int64     `json:"id,string"   form:"id"`
	Sort      int       `json:"sort"  form:"sort"`
	Type      string    `gorm:"index" json:"type" form:"type"`
	Name      string    `gorm:"index" json:"name" form:"name"`
	Value     string    `json:"value" form:"value"`
	Remark    string    `json:"remark" form:"remark"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName Specify table name
func (SysConfig) TableName() string {
	return "sys_config"
}

type SysOpr struct {
	ID        int64     `json:"id,string" form:"id"`
	Realname  string    `json:"realname" form:"realname"`
	Mobile    string    `json:"mobile" form:"mobile"`
	Email     string    `json:"email" form:"email"`
	Username  string    `json:"username" form:"username"`
	Password  string    `json:"password" form:"password"`
	Level     string    `json:"level" form:"level"`
	Status    string    `json:"status" form:"status"`
	Remark    string    `json:"remark" form:"remark"`
	LastLogin time.Time `json:"last_login" form:"last_login"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName Specify table name
func (SysOpr) TableName() string {
	return "sys_opr"
}

// SysCert is a locally managed X.509 certificate. It stores a PEM-encoded
// certificate together with an optional PEM-encoded private key so operators can
// import, name, and reference certificates from the management UI instead of
// editing on-disk file paths. Server certificates (CertType "server") carry a
// private key and are used as the EAP/TLS server identity; CA certificates
// (CertType "ca") hold a trust anchor or bundle used to verify EAP-TLS clients.
//
// The PrivateKey field is intentionally tagged json:"-" so it is never disclosed
// through the REST API; the parsed metadata fields (Subject, Issuer, Serial,
// Fingerprint, NotBefore, NotAfter) are derived from Cert when a certificate is
// imported. HasKey is computed at read time (gorm:"-") to report whether a
// private key is present without exposing the key material itself.
type SysCert struct {
	ID          int64     `json:"id,string" form:"id"`
	Name        string    `gorm:"uniqueIndex;size:128" json:"name" form:"name"`
	CertType    string    `gorm:"index;size:16" json:"cert_type" form:"cert_type"`
	Cert        string    `gorm:"type:text" json:"cert" form:"cert"`
	PrivateKey  string    `gorm:"type:text" json:"-" form:"private_key"`
	Subject     string    `gorm:"size:512" json:"subject" form:"subject"`
	Issuer      string    `gorm:"size:512" json:"issuer" form:"issuer"`
	Serial      string    `gorm:"size:128" json:"serial" form:"serial"`
	Fingerprint string    `gorm:"size:128" json:"fingerprint" form:"fingerprint"`
	NotBefore   time.Time `json:"not_before" form:"not_before"`
	NotAfter    time.Time `json:"not_after" form:"not_after"`
	HasKey      bool      `gorm:"-" json:"has_key"`
	Remark      string    `gorm:"size:512" json:"remark" form:"remark"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName Specify table name
func (SysCert) TableName() string {
	return "sys_cert"
}

type SysOprLog struct {
	ID        int64     `json:"id,string"`
	OprName   string    `json:"opr_name"`
	OprIp     string    `json:"opr_ip"`
	OptAction string    `json:"opt_action"`
	OptDesc   string    `json:"opt_desc"`
	OptTime   time.Time `json:"opt_time"`
}

// TableName Specify table name
func (SysOprLog) TableName() string {
	return "sys_opr_log"
}

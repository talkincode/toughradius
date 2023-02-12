package models

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

type SysOprLog struct {
	ID        int64     `json:"id,string"`
	OprName   string    `json:"opr_name"`
	OprIp     string    `json:"opr_ip"`
	OptAction string    `json:"opt_action"`
	OptDesc   string    `json:"opt_desc"`
	OptTime   time.Time `json:"opt_time"`
}

// SysApiToken  API 令牌
type SysApiToken struct {
	ID         string    `json:"string"`
	Uid        string    `json:"uid"`
	Name       string    `json:"name"`
	Level      string    `json:"level"`
	Token      string    `json:"token"`
	Remark     string    `json:"remark"`
	ExpireTime time.Time `json:"expire_time"`
	CreatedAt  time.Time `json:"created_at"`
}

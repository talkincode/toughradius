package domain

import (
	"encoding/json"
	"time"

	"github.com/araddon/dateparse"
	"github.com/talkincode/toughradius/v9/pkg/timeutil"
)

func (d SysOprLog) MarshalJSON() ([]byte, error) {
	type Alias SysOprLog
	return json.Marshal(&struct {
		Alias
		OptTime string `json:"opt_time"`
	}{
		Alias:   (Alias)(d),
		OptTime: d.OptTime.Format(time.RFC3339),
	})
}

func (d *RadiusUser) MarshalJSON() ([]byte, error) {
	type Alias RadiusUser
	return json.Marshal(&struct {
		*Alias
		ExpireTime string `json:"expire_time"`
		LastOnline string `json:"last_online"`
	}{
		Alias:      (*Alias)(d),
		LastOnline: d.LastOnline.Format(timeutil.YYYYMMDDHHMM_LAYOUT),
		ExpireTime: d.ExpireTime.Format(timeutil.YYYYMMDDHHMMSS_LAYOUT),
	})
}

func (d *RadiusUser) UnmarshalJSON(b []byte) error {
	type Alias RadiusUser
	var tmp = struct {
		*Alias
		ExpireTime string `json:"expire_time"`
		LastOnline string `json:"last_online"`
	}{
		Alias: (*Alias)(d),
	}
	if err := json.Unmarshal(b, &tmp); err != nil {
		return err
	}
	if t, err := time.ParseInLocation(timeutil.YYYYMMDDHHMMSS_LAYOUT, tmp.ExpireTime, time.Local); err == nil {
		d.ExpireTime = t
	} else {
		if len(tmp.ExpireTime) >= 10 {
			d.ExpireTime, _ = dateparse.ParseAny(tmp.ExpireTime[:10] + " 23:59:59")
		} else if tmp.ExpireTime != "" {
			d.ExpireTime, _ = dateparse.ParseAny(tmp.ExpireTime)
		}
	}
	d.LastOnline, _ = dateparse.ParseAny(tmp.LastOnline)
	return nil
}

/*
 * Copyright (c) 2024-2025 TalkingCode
 * Licensed under the MIT License. See LICENSE file in the project root for details.
 */

package timeutil

import (
	"strings"
	"time"

	"github.com/spf13/cast"
)

const (
	Datetime14Layout      = "20060102150405"
	Datetime8Layout       = "20060102"
	Datetime6Layout       = "200601"
	YYYYMMDDHHMMSS_LAYOUT = "2006-01-02 15:04:05"
	YYYYMMDDHHMM_LAYOUT   = "2006-01-02 15:04"
	YYYYMMDD_LAYOUT       = "2006-01-02"
)

var (
	ShangHaiLOC, _ = time.LoadLocation("Asia/Shanghai")
	EmptyTime, _   = time.Parse("2006-01-02 15:04:05 Z0700 MST", "1979-11-30 00:00:00 +0000 GMT")
)

type LocalTime time.Time

func (t *LocalTime) UnmarshalParam(src string) error {
	ts, err := time.Parse(YYYYMMDDHHMMSS_LAYOUT, src)
	*t = LocalTime(ts)
	return err
}

func (t *LocalTime) MarshalParam() string {
	lt := time.Time(*t)
	return lt.Format(YYYYMMDDHHMMSS_LAYOUT)
}

func (t *LocalTime) UnmarshalJSON(data []byte) (err error) {
	now, err := time.ParseInLocation(`"`+YYYYMMDDHHMMSS_LAYOUT+`"`, string(data), time.Local)
	*t = LocalTime(now)
	return
}

func (t LocalTime) MarshalJSON() ([]byte, error) {
	b := make([]byte, 0, len(YYYYMMDDHHMMSS_LAYOUT)+2)
	b = append(b, '"')
	b = time.Time(t).AppendFormat(b, YYYYMMDDHHMMSS_LAYOUT)
	b = append(b, '"')
	return b, nil
}

func (t LocalTime) MarshalCSV() (string, error) {
	b := make([]byte, 0, len(YYYYMMDDHHMMSS_LAYOUT)+2)
	b = append(b, '"')
	b = time.Time(t).AppendFormat(b, YYYYMMDDHHMMSS_LAYOUT)
	b = append(b, '"')
	return string(b), nil
}

// yyyy-MM-dd hh:mm:ss (year-month-day hour:minute:second)
func FmtDatetimeString(t time.Time) string {
	return t.Format(YYYYMMDDHHMMSS_LAYOUT)
}

// yyyy-MM-dd hh:mm (year-month-day hour:minute)
func FmtDatetimeMString(t time.Time) string {
	return t.Format(YYYYMMDDHHMM_LAYOUT)
}

// yy-MM-dd (year-month-day)
func FmtDateString(t time.Time) string {
	return t.Format(YYYYMMDD_LAYOUT)
}

// yyyyMMddhhmmss (yearmonthdayhourminutesecond)
func FmtDatetime14String(t time.Time) string {
	return t.Format(Datetime14Layout)
}

// yyyyMMdd (yearmonthday)
func FmtDatetime8String(t time.Time) string {
	return t.Format(Datetime8Layout)
}

// yyyyMM (yearmonth)
func FmtDatetime6String(t time.Time) string {
	return t.Format(Datetime6Layout)
}

func ComputeEndTime(times int, unit string) time.Time {
	ctime := time.Now()
	switch unit {
	case "second":
		return ctime.Add(time.Second * time.Duration(times))
	case "minute":
		return ctime.Add(time.Minute * time.Duration(times))
	case "hour":
		return ctime.Add(time.Hour * time.Duration(times))
	case "day":
		return ctime.Add(time.Hour * 24 * time.Duration(times))
	case "week":
		return ctime.Add(time.Hour * 24 * 7 * time.Duration(times))
	case "month":
		return ctime.Add(time.Hour * 24 * 30 * time.Duration(times))
	case "year":
		return ctime.Add(time.Hour * 24 * 365 * time.Duration(times))
	default:
		return ctime
	}
}

func ParseTimeDesc(timestr string) string {
	switch {
	case strings.HasPrefix(timestr, "now-") && strings.HasSuffix(timestr, "hour"):
		v := cast.ToInt(timestr[4 : len(timestr)-4])
		return time.Now().Add(time.Hour * time.Duration(v*-1)).Format(time.RFC3339)
	case strings.HasPrefix(timestr, "now-") && strings.HasSuffix(timestr, "min"):
		v := cast.ToInt(timestr[4 : len(timestr)-3])
		return time.Now().Add(time.Minute * time.Duration(v*-1)).Format(time.RFC3339)
	case strings.HasPrefix(timestr, "now-") && strings.HasSuffix(timestr, "sec"):
		v := cast.ToInt(timestr[4 : len(timestr)-3])
		return time.Now().Add(time.Second * time.Duration(v*-1)).Format(time.RFC3339)
	case strings.HasPrefix(timestr, "now-") && strings.HasSuffix(timestr, "day"):
		v := cast.ToInt(timestr[4 : len(timestr)-3])
		return time.Now().Add(time.Hour * 24 * time.Duration(v*-1)).Format(time.RFC3339)
	case timestr == "now":
		return time.Now().Format(time.RFC3339)
	default:
		return time.Now().Format(time.RFC3339)
	}
}

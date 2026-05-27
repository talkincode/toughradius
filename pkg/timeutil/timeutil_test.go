/*
 * Copyright (c) 2024-2025 TalkingCode
 * Licensed under the MIT License. See LICENSE file in the project root for details.
 */

package timeutil

import (
	"testing"
	"time"
)

func TestFormatLenTime(t *testing.T) {
	t.Log(FmtDatetime14String(time.Now()))
	t.Log(FmtDatetime8String(time.Now()))
	t.Log(FmtDatetime6String(time.Now()))
	t.Log(FmtDateString(time.Now()))
	t.Log(FmtDatetimeString(time.Now()))
	t.Log(FmtDatetimeMString(time.Now()))
}

func TestTZ(t *testing.T) {
	tz, err := time.LoadLocation("Etc/UTC")
	t.Log(tz, err)
	t.Log(time.Now().In(tz).Format("2006-01-02 15:04:05"))

}

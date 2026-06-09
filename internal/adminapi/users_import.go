package adminapi

import (
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/spf13/cast"

	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/webserver"
	"github.com/talkincode/toughradius/v9/pkg/common"
)

// ImportUserError describes one failed input row during a batch user import.
//
// Row is 1-based within the parsed payload and Username may be empty when the
// failure occurs before a username can be resolved (for example missing column).
type ImportUserError struct {
	Row      int    `json:"row"`
	Username string `json:"username"`
	Message  string `json:"message"`
}

// ImportUserResult summarizes the outcome of a batch user import request.
//
// Success and Failed are per-row counts and Errors carries only failed rows so
// callers can present actionable feedback without re-reading the source file.
type ImportUserResult struct {
	Total   int               `json:"total"`
	Success int               `json:"success"`
	Failed  int               `json:"failed"`
	Errors  []ImportUserError `json:"errors"`
}

// importMapString reads a trimmed string value from the row map using the first
// candidate key that yields a non-empty value.
func importMapString(item map[string]interface{}, keys ...string) string {
	for _, k := range keys {
		if v, ok := item[k]; ok {
			s := strings.TrimSpace(cast.ToString(v))
			if s != "" {
				return s
			}
		}
	}
	return ""
}

// importRadiusUsers handles batch import of RADIUS users from an uploaded file.
// Supported formats: Excel (.xlsx), CSV (.csv) and line-delimited JSON (.json).
func importRadiusUsers(c echo.Context) error {
	items, err := webserver.ImportData(c, "Sheet1")
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_FILE", "Unable to read import file", err.Error())
	}

	result := ImportUserResult{Errors: make([]ImportUserError, 0)}
	db := GetDB(c)

	for idx, item := range items {
		result.Total++
		username := importMapString(item, "username", "Username", "USERNAME")
		if username == "" {
			result.Failed++
			result.Errors = append(result.Errors, ImportUserError{
				Row:     idx + 1,
				Message: "Username is required",
			})
			continue
		}

		password := importMapString(item, "password", "Password", "PASSWORD")
		if password == "" {
			result.Failed++
			result.Errors = append(result.Errors, ImportUserError{
				Row:      idx + 1,
				Username: username,
				Message:  "Password is required",
			})
			continue
		}

		profileID := cast.ToInt64(importMapString(item, "profile_id", "profileid", "ProfileId", "profile"))
		if profileID == 0 {
			result.Failed++
			result.Errors = append(result.Errors, ImportUserError{
				Row:      idx + 1,
				Username: username,
				Message:  "Billing profile is required",
			})
			continue
		}

		var profile domain.RadiusProfile
		if err := db.Where("id = ?", profileID).First(&profile).Error; err != nil {
			result.Failed++
			result.Errors = append(result.Errors, ImportUserError{
				Row:      idx + 1,
				Username: username,
				Message:  "Associated billing profile not found",
			})
			continue
		}

		var exists int64
		db.Model(&domain.RadiusUser{}).Where("username = ?", username).Count(&exists)
		if exists > 0 {
			result.Failed++
			result.Errors = append(result.Errors, ImportUserError{
				Row:      idx + 1,
				Username: username,
				Message:  "Username already exists",
			})
			continue
		}

		expire, err := parseTimeInput(importMapString(item, "expire_time", "ExpireTime"), time.Now().AddDate(1, 0, 0))
		if err != nil {
			result.Failed++
			result.Errors = append(result.Errors, ImportUserError{
				Row:      idx + 1,
				Username: username,
				Message:  "Invalid expire time format",
			})
			continue
		}

		status := strings.ToLower(importMapString(item, "status", "Status"))
		if status == "" {
			status = common.ENABLED
		}

		addrPool := importMapString(item, "addr_pool", "AddrPool")
		if addrPool == "" {
			addrPool = profile.AddrPool
		}
		userDomain := importMapString(item, "domain", "Domain")
		if userDomain == "" {
			userDomain = profile.Domain
		}

		user := &domain.RadiusUser{
			ID:              common.UUIDint64(),
			NodeId:          cast.ToInt64(importMapString(item, "node_id", "NodeId")),
			ProfileId:       profileID,
			Realname:        importMapString(item, "realname", "Realname"),
			Email:           importMapString(item, "email", "Email"),
			Mobile:          importMapString(item, "mobile", "Mobile"),
			Address:         importMapString(item, "address", "Address"),
			Username:        username,
			Password:        password,
			AddrPool:        addrPool,
			ActiveNum:       profile.ActiveNum,
			UpRate:          profile.UpRate,
			DownRate:        profile.DownRate,
			Vlanid1:         cast.ToInt(importMapString(item, "vlanid1", "Vlanid1")),
			Vlanid2:         cast.ToInt(importMapString(item, "vlanid2", "Vlanid2")),
			IpAddr:          importMapString(item, "ip_addr", "IpAddr"),
			MacAddr:         importMapString(item, "mac_addr", "MacAddr"),
			Domain:          userDomain,
			IPv6PrefixPool:  profile.IPv6PrefixPool,
			BindVlan:        profile.BindVlan,
			BindMac:         profile.BindMac,
			ProfileLinkMode: domain.ProfileLinkModeStatic,
			ExpireTime:      expire,
			Status:          status,
			Remark:          importMapString(item, "remark", "Remark"),
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}

		if err := db.Create(user).Error; err != nil {
			result.Failed++
			result.Errors = append(result.Errors, ImportUserError{
				Row:      idx + 1,
				Username: username,
				Message:  err.Error(),
			})
			continue
		}
		result.Success++
	}

	return ok(c, result)
}

package adminapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/webserver"
	"github.com/talkincode/toughradius/v9/pkg/common"
)

// backupVersion identifies the backup payload schema.
const backupVersion = "9.0"

// supportedBackupMajor is the only schema major version restoreSystem accepts.
const supportedBackupMajor = "9"

// maxRestoreRecords caps the number of records accepted per table on restore,
// bounding the work a single (potentially malicious) payload can trigger.
const maxRestoreRecords = 100000

// SystemBackup is the on-disk JSON snapshot exchanged by backupSystem and
// restoreSystem.
//
// The payload is versioned by Version and contains the core configuration
// tables that define runtime behavior (nodes, NAS, profiles, users, configs,
// and operators). The snapshot intentionally preserves primary keys so restore
// can upsert records deterministically.
//
// Sensitive data notice: the payload includes security-relevant credentials
// (for example RadiusUser passwords and SysOpr password hashes). Callers must
// treat serialized backups as secrets at rest and in transit.
type SystemBackup struct {
	// Version is the backup schema version in "major.minor" form.
	Version string `json:"version"`
	// CreatedAt is the server timestamp when the backup was generated.
	CreatedAt time.Time `json:"created_at"`
	// Nodes stores exported network node definitions.
	Nodes []domain.NetNode `json:"nodes"`
	// Nas stores exported NAS device records.
	Nas []domain.NetNas `json:"nas"`
	// Profiles stores exported RADIUS profile definitions.
	Profiles []domain.RadiusProfile `json:"profiles"`
	// Users stores exported RADIUS user records.
	Users []domain.RadiusUser `json:"users"`
	// Configs stores exported dynamic system configuration items.
	Configs []domain.SysConfig `json:"configs"`
	// Operators stores exported admin operator accounts.
	Operators []domain.SysOpr `json:"operators"`
}

func registerSystemBackupRoutes() {
	webserver.ApiGET("/system/backup", backupSystem, requireAdmin())
	webserver.ApiPOST("/system/restore", restoreSystem, requireAdmin())
}

// backupSystem exports the core configuration tables as a downloadable JSON file.
// A copy is also written to the configured backup directory when available.
//
// SECURITY: the exported payload contains sensitive credentials in clear form —
// RADIUS user passwords are stored in plaintext (required for PAP/CHAP) and the
// operators table includes admin password hashes. Both the downloaded file and
// the on-disk copy in the backup directory must be handled and stored securely.
func backupSystem(c echo.Context) error {
	db := GetDB(c)

	backup := SystemBackup{
		Version:   backupVersion,
		CreatedAt: time.Now(),
	}

	if err := db.Find(&backup.Nodes).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to export nodes", err.Error())
	}
	if err := db.Find(&backup.Nas).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to export NAS", err.Error())
	}
	if err := db.Find(&backup.Profiles).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to export profiles", err.Error())
	}
	if err := db.Find(&backup.Users).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to export users", err.Error())
	}
	if err := db.Find(&backup.Configs).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to export configs", err.Error())
	}
	if err := db.Find(&backup.Operators).Error; err != nil {
		return fail(c, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to export operators", err.Error())
	}

	bs, err := json.MarshalIndent(backup, "", "  ")
	if err != nil {
		return fail(c, http.StatusInternalServerError, "ENCODE_ERROR", "Failed to encode backup", err.Error())
	}

	filename := fmt.Sprintf("toughradius-backup-%s.json", backup.CreatedAt.Format("20060102-150405"))

	// Best-effort: persist a copy of the backup to the backup directory.
	if cfg := GetAppContext(c).Config(); cfg != nil {
		if dir := cfg.GetBackupDir(); dir != "" {
			if mkErr := os.MkdirAll(dir, 0750); mkErr == nil {
				_ = os.WriteFile(filepath.Join(dir, filename), bs, 0600) //nolint:errcheck
			}
		}
	}

	c.Response().Header().Set("Content-Disposition", "attachment;filename="+filename)
	return c.JSONBlob(http.StatusOK, bs)
}

// SystemRestoreResult reports how many records restoreSystem upserted into each
// table during a successful restore transaction.
//
// A zero value for a field means either the table was absent from the payload
// or the payload contained no records for that table.
type SystemRestoreResult struct {
	// Nodes is the number of net_node records restored.
	Nodes int `json:"nodes"`
	// Nas is the number of net_nas records restored.
	Nas int `json:"nas"`
	// Profiles is the number of radius_profile records restored.
	Profiles int `json:"profiles"`
	// Users is the number of radius_user records restored.
	Users int `json:"users"`
	// Configs is the number of sys_config records restored.
	Configs int `json:"configs"`
	// Operators is the number of sys_opr records restored.
	Operators int `json:"operators"`
}

// restoreSystem imports a previously exported backup file, upserting records
// into the core configuration tables inside a single transaction.
func restoreSystem(c echo.Context) error {
	file, err := c.FormFile("upload")
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_FILE", "Backup file is required", err.Error())
	}
	src, err := file.Open()
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_FILE", "Unable to open backup file", err.Error())
	}
	defer func() { _ = src.Close() }() //nolint:errcheck

	bs, err := io.ReadAll(src)
	if err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_FILE", "Unable to read backup file", err.Error())
	}

	var backup SystemBackup
	if err := json.Unmarshal(bs, &backup); err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_BACKUP", "Invalid backup file format", err.Error())
	}

	// Strong validation: reject arbitrary, incompatible, or malformed payloads
	// before touching the database.
	if err := validateBackup(&backup); err != nil {
		return fail(c, http.StatusBadRequest, "INVALID_BACKUP", "Backup failed validation", err.Error())
	}

	// Restoring the operators table can rewrite admin password hashes and
	// privilege levels, so it is a potential privilege-escalation vector.
	// Restrict it to super operators even though admins may restore other tables.
	if len(backup.Operators) > 0 {
		current, err := resolveOperatorFromContext(c)
		if err != nil {
			return fail(c, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required", nil)
		}
		if !strings.EqualFold(current.Level, LevelSuper) {
			return fail(c, http.StatusForbidden, "PERMISSION_DENIED",
				"Only super operators may restore the operators table", nil)
		}
	}

	result := SystemRestoreResult{}
	upsert := clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		UpdateAll: true,
	}

	err = GetDB(c).Transaction(func(tx *gorm.DB) error {
		if len(backup.Nodes) > 0 {
			if err := tx.Clauses(upsert).Create(&backup.Nodes).Error; err != nil {
				return err
			}
			result.Nodes = len(backup.Nodes)
		}
		if len(backup.Nas) > 0 {
			if err := tx.Clauses(upsert).Create(&backup.Nas).Error; err != nil {
				return err
			}
			result.Nas = len(backup.Nas)
		}
		if len(backup.Profiles) > 0 {
			if err := tx.Clauses(upsert).Create(&backup.Profiles).Error; err != nil {
				return err
			}
			result.Profiles = len(backup.Profiles)
		}
		if len(backup.Users) > 0 {
			if err := tx.Clauses(upsert).Create(&backup.Users).Error; err != nil {
				return err
			}
			result.Users = len(backup.Users)
		}
		if len(backup.Configs) > 0 {
			if err := tx.Clauses(upsert).Create(&backup.Configs).Error; err != nil {
				return err
			}
			result.Configs = len(backup.Configs)
		}
		if len(backup.Operators) > 0 {
			if err := tx.Clauses(upsert).Create(&backup.Operators).Error; err != nil {
				return err
			}
			result.Operators = len(backup.Operators)
		}
		return nil
	})
	if err != nil {
		return fail(c, http.StatusInternalServerError, "RESTORE_ERROR", "Failed to restore backup", err.Error())
	}

	return ok(c, result)
}

// validateBackup performs strong, write-free validation of a restore payload.
// It rejects incompatible schema versions, oversized payloads, and records that
// are missing required identifiers or carry invalid enum-like values, ensuring a
// malformed or malicious backup cannot be upserted into the configuration tables.
func validateBackup(b *SystemBackup) error {
	major := b.Version
	if i := strings.IndexByte(major, '.'); i >= 0 {
		major = major[:i]
	}
	if major != supportedBackupMajor {
		return fmt.Errorf("incompatible backup version %q, expected %s.x", b.Version, supportedBackupMajor)
	}

	for name, n := range map[string]int{
		"nodes":     len(b.Nodes),
		"nas":       len(b.Nas),
		"profiles":  len(b.Profiles),
		"users":     len(b.Users),
		"configs":   len(b.Configs),
		"operators": len(b.Operators),
	} {
		if n > maxRestoreRecords {
			return fmt.Errorf("table %q has %d records, exceeding the limit of %d", name, n, maxRestoreRecords)
		}
	}

	for i := range b.Nodes {
		if b.Nodes[i].ID == 0 || strings.TrimSpace(b.Nodes[i].Name) == "" {
			return fmt.Errorf("nodes[%d]: id and name are required", i)
		}
	}
	for i := range b.Nas {
		if b.Nas[i].ID == 0 || strings.TrimSpace(b.Nas[i].Name) == "" {
			return fmt.Errorf("nas[%d]: id and name are required", i)
		}
		if !isValidStatus(b.Nas[i].Status) {
			return fmt.Errorf("nas[%d]: invalid status %q", i, b.Nas[i].Status)
		}
	}
	for i := range b.Profiles {
		if b.Profiles[i].ID == 0 || strings.TrimSpace(b.Profiles[i].Name) == "" {
			return fmt.Errorf("profiles[%d]: id and name are required", i)
		}
	}
	for i := range b.Users {
		if b.Users[i].ID == 0 || strings.TrimSpace(b.Users[i].Username) == "" {
			return fmt.Errorf("users[%d]: id and username are required", i)
		}
		if !isValidStatus(b.Users[i].Status) {
			return fmt.Errorf("users[%d]: invalid status %q", i, b.Users[i].Status)
		}
	}
	for i := range b.Configs {
		if b.Configs[i].ID == 0 || strings.TrimSpace(b.Configs[i].Name) == "" {
			return fmt.Errorf("configs[%d]: id and name are required", i)
		}
	}
	for i := range b.Operators {
		if b.Operators[i].ID == 0 || strings.TrimSpace(b.Operators[i].Username) == "" {
			return fmt.Errorf("operators[%d]: id and username are required", i)
		}
		if !isValidLevel(b.Operators[i].Level) {
			return fmt.Errorf("operators[%d]: invalid level %q", i, b.Operators[i].Level)
		}
		if !isValidStatus(b.Operators[i].Status) {
			return fmt.Errorf("operators[%d]: invalid status %q", i, b.Operators[i].Status)
		}
	}
	return nil
}

// isValidStatus reports whether s is an accepted account/device status. An empty
// status is allowed because callers default it elsewhere.
func isValidStatus(s string) bool {
	switch s {
	case "", common.ENABLED, common.DISABLED:
		return true
	default:
		return false
	}
}

// isValidLevel reports whether l is a recognized operator privilege level.
func isValidLevel(l string) bool {
	switch l {
	case LevelSuper, LevelAdmin, LevelOperator:
		return true
	default:
		return false
	}
}

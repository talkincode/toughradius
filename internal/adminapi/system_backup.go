package adminapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/internal/webserver"
)

// backupVersion identifies the backup payload schema.
const backupVersion = "9.0"

// SystemBackup is the serialized snapshot of the core configuration tables.
type SystemBackup struct {
	Version   string                 `json:"version"`
	CreatedAt time.Time              `json:"created_at"`
	Nodes     []domain.NetNode       `json:"nodes"`
	Nas       []domain.NetNas        `json:"nas"`
	Profiles  []domain.RadiusProfile `json:"profiles"`
	Users     []domain.RadiusUser    `json:"users"`
	Configs   []domain.SysConfig     `json:"configs"`
	Operators []domain.SysOpr        `json:"operators"`
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

// SystemRestoreResult summarizes how many records were restored per table.
type SystemRestoreResult struct {
	Nodes     int `json:"nodes"`
	Nas       int `json:"nas"`
	Profiles  int `json:"profiles"`
	Users     int `json:"users"`
	Configs   int `json:"configs"`
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

	// Guard against arbitrary or incompatible JSON: a valid backup always carries
	// a version stamp written by backupSystem.
	if backup.Version == "" {
		return fail(c, http.StatusBadRequest, "INVALID_BACKUP", "Backup file is missing a version and may be incompatible", nil)
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

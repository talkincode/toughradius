/*
 * ToughRADIUS Admin Password Reset Tool
 *
 * Usage:
 *   go run . -c toughradius.yml -u admin -p newpassword
 */

package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/talkincode/toughradius/v9/config"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"github.com/talkincode/toughradius/v9/pkg/common"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

func main() {
	var (
		configFile string
		username   string
		password   string
	)

	flag.StringVar(&configFile, "c", "toughradius.yml", "Configuration file path")
	flag.StringVar(&username, "u", "admin", "Username to reset")
	flag.StringVar(&password, "p", "toughradius", "New password")
	flag.Parse()

	// Load configuration
	cfg := config.LoadConfig(configFile)
	if cfg == nil {
		fmt.Printf("Error: Failed to load configuration from %s\n", configFile)
		os.Exit(1)
	}

	// Connect to database
	db, err := openDatabase(cfg)
	if err != nil {
		fmt.Printf("Error: Failed to connect to database: %v\n", err)
		os.Exit(1)
	}

	// Hash the new password
	hashedPassword := common.Sha256HashWithSalt(password, common.GetSecretSalt())

	// Update the password
	result := db.Model(&domain.SysOpr{}).
		Where("username = ?", username).
		Update("password", hashedPassword)

	if result.Error != nil {
		fmt.Printf("Error: Failed to update password: %v\n", result.Error)
		os.Exit(1)
	}

	if result.RowsAffected == 0 {
		fmt.Printf("Warning: User '%s' not found, creating new admin user...\n", username)

		// Create new admin user
		newAdmin := &domain.SysOpr{
			ID:       common.UUIDint64(),
			Username: username,
			Password: hashedPassword,
			Level:    "super",
			Status:   common.ENABLED,
		}

		if err := db.Create(newAdmin).Error; err != nil {
			fmt.Printf("Error: Failed to create admin user: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Success: Created new admin user '%s'\n", username)
	} else {
		fmt.Printf("Success: Password updated for user '%s'\n", username)
	}

	fmt.Println("\nNew credentials:")
	fmt.Printf("  Username: %s\n", username)
	fmt.Printf("  Password: %s\n", password)
}

func openDatabase(cfg *config.AppConfig) (*gorm.DB, error) {
	dbType := strings.ToLower(cfg.Database.Type)
	gormCfg := &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		SkipDefaultTransaction:                   true,
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
		Logger: logger.New(nil, logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Silent,
			IgnoreRecordNotFoundError: true,
		}),
	}

	switch dbType {
	case "sqlite":
		dbPath := cfg.Database.Name
		if dbPath != ":memory:" && !path.IsAbs(dbPath) {
			dbPath = path.Join(cfg.System.Workdir, "data", dbPath)
		}
		return gorm.Open(sqlite.Open(dbPath), gormCfg)

	case "postgres", "postgresql":
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable",
			cfg.Database.Host,
			cfg.Database.User,
			cfg.Database.Passwd,
			cfg.Database.Name,
			cfg.Database.Port)
		return gorm.Open(postgres.Open(dsn), gormCfg)

	default:
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}
}

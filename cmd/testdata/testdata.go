// Package main implements testdata, a database fixture seeding tool for ToughRADIUS benchmarks.
//
// testdata manages fixed test records in the database for benchmark and integration testing.
// It provides two operations:
//   - apply: Creates or updates test user, profile, NAS, and node records
//   - clear: Removes test data inserted by apply
//
// Test fixtures created:
//   - Default node (ID: AutoRegisterPopNodeId)
//   - Test profile "bmtest-profile" (2048 Kbps up/down)
//   - Test user "test1" with password "111111"
//   - Test NAS "bmtest-nas" with configurable IP and secret
//
// Example usage:
//
//	testdata apply -c toughradius.yml --nas-ip 127.0.0.1 --nas-secret testing123
//	testdata clear -c toughradius.yml
//
// Database support:
//   - SQLite (pure Go, no CGO)
//   - PostgreSQL
package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/talkincode/toughradius/v9/config"
	"github.com/talkincode/toughradius/v9/internal/app"
	"github.com/talkincode/toughradius/v9/internal/domain"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

const (
	defaultConfigPath = "toughradius.yml" // Default configuration file path
	defaultNasIP      = "127.0.0.1"       // Default NAS IP for test fixtures
	defaultNasSecret  = "testing123"      // Standard test shared secret

	// Fixed test fixture identifiers (must match benchmark tool expectations)
	testProfileName   = "bmtest-profile"    // Profile name for benchmark tests
	testNasIdentifier = "bmtest-nas"        // NAS identifier for test NAS
	testNasName       = "bmtest-nas"        // NAS display name
	testUsername      = "test1"             // Test user login name
	testPassword      = "111111"            // Test user password (plain text)
	testUserIP        = "172.16.0.10"       // Allocated IP for test user
	testUserMAC       = "AA-BB-CC-DD-EE-FF" // Test user MAC address

	// Supported commands
	actionApply = "apply" // Create/update test data
	actionClear = "clear" // Remove test data
)

// commandOptions holds parsed command-line arguments.
type commandOptions struct {
	ConfigPath string // Path to toughradius.yml configuration
	NasIP      string // NAS IP address for test NAS record
	NasSecret  string // Shared secret for test NAS authentication
}

// seedOptions contains parameters for test data seeding operations.
type seedOptions struct {
	NasIP     string // NAS IP address to populate in database
	NasSecret string // NAS shared secret to populate in database
}

func main() {
	log.SetFlags(0)
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	action := strings.ToLower(os.Args[1])
	var opts commandOptions
	switch action {
	case actionApply:
		opts = parseApplyFlags(os.Args[2:])
	case actionClear:
		opts = parseClearFlags(os.Args[2:])
	default:
		printUsage()
		os.Exit(1)
	}

	if err := run(action, opts); err != nil {
		log.Fatalf("%s failed: %v", action, err)
	}
	log.Printf("%s completed successfully", action)
}

// printUsage displays command-line help text to stdout.
func printUsage() {
	fmt.Println("testdata seeds or clears fixed benchmark fixtures.")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  apply [-c toughradius.yml] [--nas-ip 127.0.0.1] [--nas-secret testing123]")
	fmt.Println("  clear [-c toughradius.yml]")
}

// parseApplyFlags parses command-line arguments for the "apply" command.
// Includes NAS IP and secret flags for customizing test NAS configuration.
//
// Parameters:
//   - args: Command-line arguments (excluding command name)
//
// Returns:
//   - commandOptions: Parsed configuration with NAS parameters
func parseApplyFlags(args []string) commandOptions {
	fs := flag.NewFlagSet(actionApply, flag.ExitOnError)
	return parseFlags(fs, args, true)
}

// parseClearFlags parses command-line arguments for the "clear" command.
// Only requires configuration path; NAS parameters are not used.
//
// Parameters:
//   - args: Command-line arguments (excluding command name)
//
// Returns:
//   - commandOptions: Parsed configuration
func parseClearFlags(args []string) commandOptions {
	fs := flag.NewFlagSet(actionClear, flag.ExitOnError)
	return parseFlags(fs, args, false)
}

// parseFlags is the common flag parser for both apply and clear commands.
//
// Parameters:
//   - fs: FlagSet to register flags on
//   - args: Command-line arguments to parse
//   - includeNas: Whether to include NAS-specific flags (--nas-ip, --nas-secret)
//
// Returns:
//   - commandOptions: Parsed options with defaults applied
func parseFlags(fs *flag.FlagSet, args []string, includeNas bool) commandOptions {
	var opts commandOptions
	fs.StringVar(&opts.ConfigPath, "c", defaultConfigPath, "Path to the configuration file")
	fs.StringVar(&opts.ConfigPath, "config", defaultConfigPath, "Path to the configuration file")
	if includeNas {
		fs.StringVar(&opts.NasIP, "nas-ip", defaultNasIP, "NAS IP address for the generated record")
		fs.StringVar(&opts.NasSecret, "nas-secret", defaultNasSecret, "Shared secret for the generated NAS record")
	} else {
		opts.NasIP = defaultNasIP
		opts.NasSecret = defaultNasSecret
	}
	if err := fs.Parse(args); err != nil {
		os.Exit(2)
	}
	if opts.ConfigPath == "" {
		opts.ConfigPath = defaultConfigPath
	}
	if includeNas {
		if opts.NasIP == "" {
			opts.NasIP = defaultNasIP
		}
		if opts.NasSecret == "" {
			opts.NasSecret = defaultNasSecret
		}
	}
	return opts
}

// run executes the requested command (apply or clear) against the database.
//
// Parameters:
//   - action: Command to execute ("apply" or "clear")
//   - opts: Parsed command options including config path
//
// Returns:
//   - error: Database connection, migration, or seeding error; nil on success
//
// Side effects:
//   - Loads configuration from file
//   - Opens database connection
//   - Runs GORM auto-migration on all domain tables
//   - Creates or deletes test records via transaction
func run(action string, opts commandOptions) error {
	cfg := config.LoadConfig(opts.ConfigPath)
	db, err := openDatabase(cfg)
	if err != nil {
		return fmt.Errorf("open database: %w", err)
	}
	defer closeDatabase(db)

	if err := db.AutoMigrate(domain.Tables...); err != nil {
		return fmt.Errorf("auto migrate: %w", err)
	}

	switch action {
	case actionApply:
		return applyTestData(db, seedOptions{NasIP: opts.NasIP, NasSecret: opts.NasSecret})
	case actionClear:
		return clearTestData(db)
	default:
		return fmt.Errorf("unsupported action %q", action)
	}
}

// openDatabase establishes a GORM database connection based on configuration.
// Supports both SQLite (default) and PostgreSQL.
//
// Parameters:
//   - cfg: Application configuration with database settings
//
// Returns:
//   - *gorm.DB: Connected database instance
//   - error: Connection or configuration error; nil on success
func openDatabase(cfg *config.AppConfig) (*gorm.DB, error) {
	dbType := strings.ToLower(strings.TrimSpace(cfg.Database.Type))
	if dbType == "" {
		dbType = "sqlite"
	}
	switch dbType {
	case "sqlite":
		return openSqliteDatabase(cfg)
	case "postgres", "postgresql":
		return openPostgresDatabase(cfg)
	default:
		return nil, fmt.Errorf("unsupported database type %q", cfg.Database.Type)
	}
}

// openSqliteDatabase opens a SQLite database connection.
// Uses pure Go implementation (github.com/glebarez/sqlite), no CGO required.
//
// Parameters:
//   - cfg: Application configuration with database settings
//
// Returns:
//   - *gorm.DB: SQLite database instance
//   - error: File creation or connection error; nil on success
//
// Side effects:
//   - Creates database directory if it doesn't exist
//   - Sets MaxIdleConns=1, MaxOpenConns=1 for SQLite limitations
func openSqliteDatabase(cfg *config.AppConfig) (*gorm.DB, error) {
	dbPath := cfg.Database.Name
	if dbPath == "" {
		dbPath = "toughradius.db"
	}
	if dbPath != ":memory:" && !filepath.IsAbs(dbPath) {
		dbPath = filepath.Join(cfg.System.Workdir, "data", dbPath)
	}
	if dbPath != ":memory:" {
		if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
			return nil, fmt.Errorf("prepare sqlite path: %w", err)
		}
	}
	gormCfg := &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		SkipDefaultTransaction:                   true,
		PrepareStmt:                              true,
		NamingStrategy:                           schema.NamingStrategy{SingularTable: true},
		Logger:                                   logger.Default.LogMode(selectLogLevel(cfg.Database.Debug)),
	}
	db, err := gorm.Open(sqlite.Open(dbPath), gormCfg)
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxIdleConns(1)
	sqlDB.SetMaxOpenConns(1)
	return db, nil
}

// openPostgresDatabase opens a PostgreSQL database connection.
//
// Parameters:
//   - cfg: Application configuration with PostgreSQL credentials
//
// Returns:
//   - *gorm.DB: PostgreSQL database instance
//   - error: Connection or authentication error; nil on success
//
// Side effects:
//   - Configures connection pool based on cfg.Database.IdleConn and MaxConn
func openPostgresDatabase(cfg *config.AppConfig) (*gorm.DB, error) {
	dbCfg := cfg.Database
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Asia/Shanghai",
		dbCfg.Host,
		dbCfg.User,
		dbCfg.Passwd,
		dbCfg.Name,
		dbCfg.Port,
	)
	gormCfg := &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		SkipDefaultTransaction:                   true,
		PrepareStmt:                              true,
		NamingStrategy:                           schema.NamingStrategy{SingularTable: true},
		Logger:                                   logger.Default.LogMode(selectLogLevel(dbCfg.Debug)),
	}
	db, err := gorm.Open(postgres.Open(dsn), gormCfg)
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	if dbCfg.IdleConn > 0 {
		sqlDB.SetMaxIdleConns(dbCfg.IdleConn)
	}
	if dbCfg.MaxConn > 0 {
		sqlDB.SetMaxOpenConns(dbCfg.MaxConn)
	}
	return db, nil
}

// selectLogLevel converts boolean debug flag to GORM log level.
//
// Parameters:
//   - debug: Whether to enable verbose SQL logging
//
// Returns:
//   - logger.LogLevel: Info if debug=true, Silent otherwise
func selectLogLevel(debug bool) logger.LogLevel {
	if debug {
		return logger.Info
	}
	return logger.Silent
}

// closeDatabase closes the underlying database connection.
// Safe to call multiple times; errors are silently ignored.
//
// Parameters:
//   - db: GORM database instance to close
func closeDatabase(db *gorm.DB) {
	sqlDB, err := db.DB()
	if err != nil {
		return
	}
	_ = sqlDB.Close()
}

// applyTestData creates or updates all test fixtures in a single transaction.
// Ensures referential integrity by creating records in dependency order:
//  1. Default node (required for all other records)
//  2. Test profile (required for user)
//  3. Test user (references profile)
//  4. Test NAS
//
// Parameters:
//   - db: GORM database instance
//   - opts: Seeding options with NAS IP and secret
//
// Returns:
//   - error: Transaction error or nil on success
//
// Side effects:
//   - Uses database transaction for atomicity
//   - Performs upsert (creates if not exists, updates if exists)
//   - Logs seeding progress
func applyTestData(db *gorm.DB, opts seedOptions) error {
	log.Printf("Seeding test data (user=%s, nas_ip=%s)", testUsername, opts.NasIP)
	return db.Transaction(func(tx *gorm.DB) error {
		if err := ensureDefaultNode(tx); err != nil {
			return err
		}
		profile, err := ensureTestProfile(tx)
		if err != nil {
			return err
		}
		if err := ensureTestUser(tx, profile.ID); err != nil {
			return err
		}
		if err := ensureTestNas(tx, opts); err != nil {
			return err
		}
		return nil
	})
}

// clearTestData removes all test fixtures created by applyTestData.
// Deletion order respects foreign key constraints.
//
// Parameters:
//   - db: GORM database instance
//
// Returns:
//   - error: Transaction error or nil on success
//
// Side effects:
//   - Deletes test user by username
//   - Deletes test NAS by identifier
//   - Deletes test profile only if no other users reference it
//   - Uses database transaction for atomicity
func clearTestData(db *gorm.DB) error {
	log.Printf("Removing test data (user=%s)", testUsername)
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("username = ?", testUsername).Delete(&domain.RadiusUser{}).Error; err != nil {
			return err
		}
		if err := tx.Where("identifier = ?", testNasIdentifier).Delete(&domain.NetNas{}).Error; err != nil {
			return err
		}
		var profile domain.RadiusProfile
		err := tx.Where("name = ?", testProfileName).First(&profile).Error
		if err == nil {
			var userCount int64
			if err := tx.Model(&domain.RadiusUser{}).Where("profile_id = ?", profile.ID).Count(&userCount).Error; err != nil {
				return err
			}
			if userCount == 0 {
				if err := tx.Delete(&profile).Error; err != nil {
					return err
				}
			}
		} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		return nil
	})
}

// ensureDefaultNode creates the default node if it doesn't exist.
// Required for all other test records with foreign key to net_node.
//
// Parameters:
//   - tx: GORM transaction instance
//
// Returns:
//   - error: Database error or nil on success
func ensureDefaultNode(tx *gorm.DB) error {
	var node domain.NetNode
	err := tx.Where("id = ?", app.AutoRegisterPopNodeId).First(&node).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		node = domain.NetNode{
			ID:     app.AutoRegisterPopNodeId,
			Name:   "default",
			Remark: "Created by cmd/testdata",
			Tags:   "system,testdata",
		}
		return tx.Create(&node).Error
	}
	return err
}

// ensureTestProfile creates or updates the test profile.
// Profile defines bandwidth limits and session constraints for test user.
//
// Parameters:
//   - tx: GORM transaction instance
//
// Returns:
//   - *domain.RadiusProfile: Created or updated profile
//   - error: Database error or nil on success
//
// Configuration:
//   - UpRate/DownRate: 2048 Kbps (2 Mbps)
//   - ActiveNum: 32 (max concurrent sessions)
//   - Status: "enabled"
func ensureTestProfile(tx *gorm.DB) (*domain.RadiusProfile, error) {
	var profile domain.RadiusProfile
	err := tx.Where("name = ?", testProfileName).First(&profile).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		profile = domain.RadiusProfile{
			NodeId:    app.AutoRegisterPopNodeId,
			Name:      testProfileName,
			Status:    "enabled",
			ActiveNum: 32,
			UpRate:    2048,
			DownRate:  2048,
			Remark:    "Test profile generated by cmd/testdata",
		}
		if err := tx.Create(&profile).Error; err != nil {
			return nil, err
		}
		return &profile, nil
	}
	if err != nil {
		return nil, err
	}
	profile.NodeId = app.AutoRegisterPopNodeId
	profile.Status = "enabled"
	profile.Remark = "Test profile generated by cmd/testdata"
	if err := tx.Save(&profile).Error; err != nil {
		return nil, err
	}
	return &profile, nil
}

// ensureTestUser creates or updates the test user account.
//
// Parameters:
//   - tx: GORM transaction instance
//   - profileID: ID of the test profile to associate with user
//
// Returns:
//   - error: Database error or nil on success
//
// User configuration:
//   - Username: "test1", Password: "111111" (plain text)
//   - IP: 172.16.0.10, MAC: AA-BB-CC-DD-EE-FF
//   - Expiration: 1 year from creation
//   - Status: "enabled"
func ensureTestUser(tx *gorm.DB, profileID int64) error {
	var user domain.RadiusUser
	err := tx.Where("username = ?", testUsername).First(&user).Error
	expire := time.Now().AddDate(1, 0, 0)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		user = domain.RadiusUser{
			NodeId:     app.AutoRegisterPopNodeId,
			ProfileId:  profileID,
			Realname:   "Benchmark User",
			Mobile:     "00000000000",
			Username:   testUsername,
			Password:   testPassword,
			ActiveNum:  1,
			UpRate:     2048,
			DownRate:   2048,
			IpAddr:     testUserIP,
			MacAddr:    testUserMAC,
			BindMac:    0,
			BindVlan:   0,
			Status:     "enabled",
			ExpireTime: expire,
			Remark:     "Created by cmd/testdata",
		}
		return tx.Create(&user).Error
	}
	if err != nil {
		return err
	}
	user.ProfileId = profileID
	user.Password = testPassword
	user.Status = "enabled"
	user.ExpireTime = expire
	user.IpAddr = testUserIP
	user.MacAddr = testUserMAC
	user.Remark = "Created by cmd/testdata"
	return tx.Save(&user).Error
}

// ensureTestNas creates or updates the test NAS record.
// NAS represents a Network Access Server that will send RADIUS requests.
//
// Parameters:
//   - tx: GORM transaction instance
//   - opts: Seeding options with NAS IP and shared secret
//
// Returns:
//   - error: Database error or nil on success
//
// NAS configuration:
//   - Identifier: "bmtest-nas" (must match benchmark tool)
//   - IP/Hostname: Configurable via --nas-ip flag
//   - Secret: Configurable via --nas-secret flag
//   - CoA Port: 3799 (RFC 5176 Change of Authorization)
//   - Model: "software", Vendor: "generic"
func ensureTestNas(tx *gorm.DB, opts seedOptions) error {
	var nas domain.NetNas
	err := tx.Where("identifier = ?", testNasIdentifier).First(&nas).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		nas = domain.NetNas{
			NodeId:     app.AutoRegisterPopNodeId,
			Name:       testNasName,
			Identifier: testNasIdentifier,
			Hostname:   opts.NasIP,
			Ipaddr:     opts.NasIP,
			Secret:     opts.NasSecret,
			CoaPort:    3799,
			Model:      "software",
			VendorCode: "generic",
			Status:     "enabled",
			Tags:       "testdata",
			Remark:     "Created by cmd/testdata",
		}
		return tx.Create(&nas).Error
	}
	if err != nil {
		return err
	}
	nas.Hostname = opts.NasIP
	nas.Ipaddr = opts.NasIP
	nas.Secret = opts.NasSecret
	nas.Status = "enabled"
	nas.Remark = "Created by cmd/testdata"
	return tx.Save(&nas).Error
}

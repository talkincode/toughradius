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
	defaultConfigPath = "toughradius.yml"
	defaultNasIP      = "127.0.0.1"
	defaultNasSecret  = "testing123"

	testProfileName   = "bmtest-profile"
	testNasIdentifier = "bmtest-nas"
	testNasName       = "bmtest-nas"
	testUsername      = "test1"
	testPassword      = "111111"
	testUserIP        = "172.16.0.10"
	testUserMAC       = "AA-BB-CC-DD-EE-FF"

	actionApply = "apply"
	actionClear = "clear"
)

type commandOptions struct {
	ConfigPath string
	NasIP      string
	NasSecret  string
}

type seedOptions struct {
	NasIP     string
	NasSecret string
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

func printUsage() {
	fmt.Println("testdata seeds or clears fixed benchmark fixtures.")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  apply [-c toughradius.yml] [--nas-ip 127.0.0.1] [--nas-secret testing123]")
	fmt.Println("  clear [-c toughradius.yml]")
}

func parseApplyFlags(args []string) commandOptions {
	fs := flag.NewFlagSet(actionApply, flag.ExitOnError)
	return parseFlags(fs, args, true)
}

func parseClearFlags(args []string) commandOptions {
	fs := flag.NewFlagSet(actionClear, flag.ExitOnError)
	return parseFlags(fs, args, false)
}

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

func selectLogLevel(debug bool) logger.LogLevel {
	if debug {
		return logger.Info
	}
	return logger.Silent
}

func closeDatabase(db *gorm.DB) {
	sqlDB, err := db.DB()
	if err != nil {
		return
	}
	_ = sqlDB.Close()
}

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

package app

import (
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/talkincode/toughradius/v9/config"
	"github.com/talkincode/toughradius/v9/pkg/common"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

// getDatabase 根据配置类型获取数据库连接
func getDatabase(config config.DBConfig) *gorm.DB {
	dbType := strings.ToLower(config.Type)
	switch dbType {
	case "sqlite":
		return getSqliteDatabase(config)
	case "postgres", "postgresql":
		return getPgDatabase(config)
	default:
		zap.S().Fatalf("不支持的数据库类型: %s，支持的类型: postgres, sqlite", config.Type)
		return nil
	}
}

// getSqliteDatabase 获取 SQLite 数据库连接
func getSqliteDatabase(config config.DBConfig) *gorm.DB {
	// 如果 Name 不是绝对路径且不是内存数据库,则放在 workdir/data 目录下
	dbPath := config.Name
	if dbPath != ":memory:" && !path.IsAbs(dbPath) {
		dbPath = path.Join(GConfig().System.Workdir, "data", dbPath)
	}

	zap.S().Infof("SQLite 数据库路径: %s", dbPath)

	pool, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		SkipDefaultTransaction:                   true,
		PrepareStmt:                              true,
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
		Logger: logger.New(
			zap.NewStdLog(zap.L()),
			logger.Config{
				SlowThreshold:             time.Millisecond * 200,
				LogLevel:                  common.If(config.Debug, logger.Info, logger.Silent).(logger.LogLevel),
				IgnoreRecordNotFoundError: true,
				Colorful:                  false,
			},
		),
	})
	common.Must(err)

	sqlDB, err := pool.DB()
	common.Must(err)

	// SQLite 连接池设置
	sqlDB.SetMaxIdleConns(1)
	sqlDB.SetMaxOpenConns(1)

	return pool
}

// getPgDatabase 获取 PostgreSQL 数据库连接
func getPgDatabase(config config.DBConfig) *gorm.DB {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=Asia/Shanghai",
		config.Host,
		config.User,
		config.Passwd,
		config.Name,
		config.Port)
	pool, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		SkipDefaultTransaction:                   true,
		PrepareStmt:                              true,
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // use singular table name, table for `User` would be `user` with this option enabled
		},
		Logger: logger.New(
			zap.NewStdLog(zap.L()), // io writer
			logger.Config{
				SlowThreshold:             time.Millisecond * 200,                                                // Slow SQL threshold
				LogLevel:                  common.If(config.Debug, logger.Info, logger.Silent).(logger.LogLevel), // Log level
				IgnoreRecordNotFoundError: true,                                                                  // Ignore ErrRecordNotFound error for logger
				Colorful:                  false,                                                                 // Disable color
			},
		),
	})
	common.Must(err)
	sqlDB, err := pool.DB()
	common.Must(err)
	// SetMaxIdleConns 设置空闲连接池中连接的最大数量
	sqlDB.SetMaxIdleConns(config.IdleConn)
	// SetMaxOpenConns 设置打开数据库连接的最大数量。
	sqlDB.SetMaxOpenConns(config.MaxConn)
	// SetConnMaxLifetime 设置了连接可复用的最大时间。
	// sqlDB.SetConnMaxLifetime(time.Hour * 8)
	return pool
}

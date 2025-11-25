package app

import (
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/talkincode/toughradius/v9/config"
	"github.com/talkincode/toughradius/v9/pkg/common"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

// getDatabase returns a database connection based on the configuration type
func getDatabase(dbConfig config.DBConfig, workdir string) *gorm.DB {
	dbType := strings.ToLower(dbConfig.Type)
	switch dbType {
	case "sqlite":
		return getSqliteDatabase(dbConfig, workdir)
	case "postgres", "postgresql":
		return getPgDatabase(dbConfig)
	default:
		zap.S().Fatalf("Unsupported database type: %s, supported types: postgres, sqlite", dbConfig.Type)
		return nil
	}
}

// getSqliteDatabase returns a SQLite database connection
func getSqliteDatabase(config config.DBConfig, workdir string) *gorm.DB {
	// e.g., if the name is not an absolute path and not an in-memory DB, store it under workdir/data
	dbPath := config.Name
	if dbPath != ":memory:" && !path.IsAbs(dbPath) {
		dbPath = path.Join(workdir, "data", dbPath)
	}

	zap.S().Infof("SQLite database path: %s", dbPath)

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

	// SQLite connection pool settings
	sqlDB.SetMaxIdleConns(1)
	sqlDB.SetMaxOpenConns(1)

	return pool
}

// getPgDatabase returns a PostgreSQL database connection
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
	// SetMaxIdleConns sets the maximum number of idle connections in the pool
	sqlDB.SetMaxIdleConns(config.IdleConn)
	// SetMaxOpenConns sets the maximum number of open database connections
	sqlDB.SetMaxOpenConns(config.MaxConn)
	// SetConnMaxLifetime sets the maximum lifetime a connection can be reused
	// sqlDB.SetConnMaxLifetime(time.Hour * 8)
	return pool
}

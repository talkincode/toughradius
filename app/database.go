package app

import (
	"fmt"
	"time"

	"github.com/talkincode/toughradius/common"
	"github.com/talkincode/toughradius/config"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

// getPgDatabase 获取数据库连接，执行一次
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

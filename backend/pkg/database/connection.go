package database

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Init 初始化数据库连接
func SetupDatabase(config DatabaseConfig, gormOption ...gorm.Option) (*gorm.DB, error) {
	var db *gorm.DB
	var err error

	// 根据不同的数据库驱动初始化连接
	switch config.Driver {
	case "mysql":
		db, err = initMySQL(config.MySQL, gormOption...)
	case "postgres":
		db, err = initPostgres(config.Postgres, gormOption...)
	case "sqlite":
		db, err = initSQLite(config.SQLite, gormOption...)
	default:
		return nil, fmt.Errorf("不支持的数据库驱动: %s", config.Driver)
	}

	if err != nil {
		return nil, fmt.Errorf("数据库连接失败: %v", err)
	}

	return db, nil
}

// initMySQL 初始化MySQL连接
func initMySQL(config MysqlConfig, gormOption ...gorm.Option) (*gorm.DB, error) {
	// 构建DSN (Data Source Name)
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=%t&loc=%s",
		config.Username,
		config.Password,
		config.Host,
		config.Port,
		config.Database,
		config.Charset,
		config.ParseTime,
		config.Loc,
	)

	// 打开数据库连接
	db, err := gorm.Open(mysql.Open(dsn), gormOption...)
	if err != nil {
		return nil, err
	}

	// 获取底层sql.DB以设置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// 设置连接池参数
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime)

	return db, nil
}

// initPostgres 初始化PostgreSQL连接
func initPostgres(config PgConfig, gormOption ...gorm.Option) (*gorm.DB, error) {
	// 构建DSN
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
		config.Host,
		config.Port,
		config.Username,
		config.Password,
		config.Database,
		config.SSLMode,
		config.Timezone,
	)

	// 打开数据库连接
	db, err := gorm.Open(postgres.Open(dsn), gormOption...)
	if err != nil {
		return nil, err
	}

	// 获取底层sql.DB以设置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// 设置连接池参数
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime)

	return db, nil
}

// initSQLite 初始化SQLite连接
func initSQLite(config SqliteConfig, gormOption ...gorm.Option) (*gorm.DB, error) {
	// 构建DSN
	dsn := config.Path
	if config.ForeignKeys {
		dsn += "?foreign_keys=on"
	}

	// 打开数据库连接
	db, err := gorm.Open(sqlite.Open(dsn), gormOption...)
	if err != nil {
		return nil, err
	}

	// 获取底层sql.DB以设置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// 设置连接池参数
	// SQLite默认只允许一个写连接，所以通常MaxOpenConns设为1
	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime)

	return db, nil
}

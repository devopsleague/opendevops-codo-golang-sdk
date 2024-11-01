package mysql

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/XSAM/otelsql"
	_ "github.com/go-sql-driver/mysql"
	"go.opentelemetry.io/otel"

	"go.opentelemetry.io/otel/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
)

type DBConfigOption func(c *DBConfig)

type DBConfig struct {
	Host   string `json:"host" yaml:"host" env:"DEFAULT_MYSQL_HOST"`
	Port   uint32 `json:"port" yaml:"port" env:"DEFAULT_MYSQL_PORT"`
	User   string `json:"user" yaml:"user" env:"DEFAULT_MYSQL_USER"`
	Pass   string `json:"pass" yaml:"pass" env:"DEFAULT_MYSQL_PASS"`
	DBName string `json:"db_name" yaml:"dbName" env:"DEFAULT_MYSQL_DB_NAME"`

	ConnMaxIdleTime uint32 `json:"conn_max_idle_time" yaml:"connMaxIdleTime" env:"DEFAULT_MYSQL_CONN_MAX_IDLE_TIME"`
	ConnMaxLifetime uint32 `json:"conn_max_lifetime" yaml:"connMaxLifetime" env:"DEFAULT_MYSQL_CONN_MAX_LIFETIME"`
	MaxIdleConns    uint32 `json:"max_idle_conns" yaml:"maxIdleConns" env:"DEFAULT_MYSQL_MAX_IDLE_CONNS"`
	MaxOpenConns    uint32 `json:"max_open_conns" yaml:"maxOpenConns" env:"DEFAULT_MYSQL_MAX_OPEN_CONNS"`

	TracerProvider trace.TracerProvider `json:"-"`
	MeterProvider  metric.MeterProvider `json:"-"`
}

func defaultConfig() DBConfig {
	emptyOr := func(str string, defaultVal string) string {
		str = strings.TrimSpace(str)
		if str == "" {
			return defaultVal
		}
		return str
	}
	host := emptyOr(os.Getenv("DEFAULT_MYSQL_HOST"), "127.0.0.1")
	port := emptyOr(os.Getenv("DEFAULT_MYSQL_PORT"), "3306")
	user := emptyOr(os.Getenv("DEFAULT_MYSQL_USER"), "admin")
	pass := emptyOr(os.Getenv("DEFAULT_MYSQL_PASS"), "123456")
	dbName := emptyOr(os.Getenv("DEFAULT_MYSQL_DB_NAME"), "default")
	connMaxIdleTime := emptyOr(os.Getenv("DEFAULT_MYSQL_CONN_MAX_IDLE_TIME"), "300")
	connMaxLifetime := emptyOr(os.Getenv("DEFAULT_MYSQL_CONN_MAX_LIFETIME"), "300")
	maxIdleConns := emptyOr(os.Getenv("DEFAULT_MYSQL_MAX_IDLE_CONNS"), "60")
	maxOpenConns := emptyOr(os.Getenv("DEFAULT_MYSQL_MAX_OPEN_CONNS"), "60")

	i64Port, _ := strconv.Atoi(port)
	i64ConnMaxIdleTime, _ := strconv.Atoi(connMaxIdleTime)
	i64ConnMaxLifetime, _ := strconv.Atoi(connMaxLifetime)
	i64MaxIdleConns, _ := strconv.Atoi(maxIdleConns)
	i64MaxOpenConns, _ := strconv.Atoi(maxOpenConns)
	return DBConfig{
		Host:            host,
		Port:            uint32(i64Port),
		User:            user,
		Pass:            pass,
		DBName:          dbName,
		ConnMaxIdleTime: uint32(i64ConnMaxIdleTime),
		ConnMaxLifetime: uint32(i64ConnMaxLifetime),
		MaxIdleConns:    uint32(i64MaxIdleConns),
		MaxOpenConns:    uint32(i64MaxOpenConns),

		TracerProvider: otel.GetTracerProvider(),
		MeterProvider:  otel.GetMeterProvider(),
	}
}

func NewMysql(opts ...DBConfigOption) (*sql.DB, func(), error) {
	cfg := defaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}
	netAddr := fmt.Sprintf("tcp(%s:%d)", cfg.Host, cfg.Port)
	dsn := fmt.Sprintf("%s:%s@%s/%s?timeout=30s&charset=utf8mb4", cfg.User, cfg.Pass, netAddr, cfg.DBName)

	driverName, err := otelsql.Register(
		"mysql",
		otelsql.WithTracerProvider(cfg.TracerProvider),
		otelsql.WithAttributes(
			semconv.DBSystemMySQL,
		),
		otelsql.WithMeterProvider(cfg.MeterProvider),
	)
	if err != nil {
		return nil, nil, err
	}

	db, err := sql.Open(driverName, dsn)
	if err != nil {
		return nil, nil, err
	}

	db.SetConnMaxIdleTime(time.Duration(cfg.ConnMaxIdleTime) * time.Second)
	db.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)
	db.SetMaxIdleConns(int(cfg.MaxIdleConns))
	db.SetMaxOpenConns(int(cfg.MaxOpenConns))
	return db, func() {
		db.Close()
	}, nil
}

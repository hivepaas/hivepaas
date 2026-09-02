package config

import (
	"fmt"
	"time"
)

type DB struct {
	Host     string `toml:"host" env:"HP_DB_HOST"`
	Port     int    `toml:"port" env:"HP_DB_PORT"`
	User     string `toml:"user" env:"HP_DB_USER"`
	Password string `toml:"password" env:"HP_DB_PASSWORD"`
	DBName   string `toml:"db_name" env:"HP_DB_DB_NAME"`
	// NOTE: keep MaxIdleConns equal to MaxOpenConns. A smaller idle pool makes bursty load close
	// and reopen connections, paying TCP + TLS + auth each time.
	//
	// The budget per process is roughly: task queue concurrency (HP_TASKS_QUEUE_CONCURRENCY, 10 by
	// default, in the same process under run mode app+worker) + concurrent HTTP handlers + the
	// connections long-running jobs pin, such as a backup repository cleanup holding an advisory
	// lock. Postgres allows 100 connections by default, 97 of them non-superuser, so
	// instances x MaxOpenConns has to stay under that.
	MaxOpenConns int `toml:"max_open_conns" env:"HP_DB_MAX_OPEN_CONNS" default:"30"`
	MaxIdleConns int `toml:"max_idle_conns" env:"HP_DB_MAX_IDLE_CONNS" default:"30"`
	// ConnMaxIdleTime returns connections the app is not using, so an idle instance does not sit
	// on the whole pool until ConnMaxLifetime expires.
	ConnMaxIdleTime time.Duration `toml:"conn_max_idle_time" env:"HP_DB_MAX_IDLE_TIME" default:"10m"`
	ConnMaxLifetime time.Duration `toml:"conn_max_lifetime" env:"HP_DB_MAX_LIFETIME" default:"60m"`
	SSLMode         string        `toml:"ssl_mode" env:"HP_DB_SSL_MODE" default:"require"`
}

func (c *DB) GetDSN() string {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host,
		c.Port,
		c.User,
		c.Password,
		c.DBName,
		c.SSLMode,
	)
	return dsn
}

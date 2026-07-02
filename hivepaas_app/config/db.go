package config

import (
	"fmt"
	"time"
)

type DB struct {
	Host            string        `toml:"host" env:"HP_DB_HOST"`
	Port            int           `toml:"port" env:"HP_DB_PORT"`
	User            string        `toml:"user" env:"HP_DB_USER"`
	Password        string        `toml:"password" env:"HP_DB_PASSWORD"`
	DBName          string        `toml:"db_name" env:"HP_DB_DB_NAME"`
	MaxOpenConns    int           `toml:"max_open_conns" env:"HP_DB_MAX_OPEN_CONNS" default:"20"`
	MaxIdleConns    int           `toml:"max_idle_conns" env:"HP_DB_MAX_IDLE_CONNS" default:"20"`
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

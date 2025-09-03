package config

import (
	"fmt"
	"net/url"
)

type PostgresDB struct {
	Addr            string `validate:"required" usage:"storage address in format ip:port"`
	DBName          string `yaml:"db_name" env:"DB_NAME" validate:"required"  usage:"storage user"`
	User            string `validate:"required" secret:"true" usage:"storage user password"`
	Password        string `validate:"required" secret:"true" usage:"storage db name"`
	ConnMaxLifetime int    `yaml:"conn_max_lifetime" default:"180"`
	MaxOpenConns    int32  `yaml:"max_open_conns" default:"100"`
	MaxIdleConns    int32  `yaml:"max_idle_conns" default:"100"`
	MinOpenConns    int32  `yaml:"min_open_conns" default:"6"`
}

func (conf PostgresDB) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", conf.User, url.QueryEscape(conf.Password), conf.Addr, conf.DBName)
}

func (conf PostgresDB) Engine() string {
	return "postgres"
}

type RedisDB struct {
	Addr     string `validate:"required" usage:"storage address in format ip:port"`
	User     string `secret:"true" usage:"storage user password"`
	Password string `secret:"true" usage:"storage db name"`
	DBIndex  int    `yaml:"db_index" usage:"index of database"`
}

func (r RedisDB) URL() string {
	return fmt.Sprintf("redis://%s/%d?protocol=3", r.Addr, r.DBIndex)
}

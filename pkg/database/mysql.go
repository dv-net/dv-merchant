package database

import (
	"database/sql"
	"time"
)

type MySQLClient struct {
	DB *sql.DB
}

func NewMysqlClient(dsn string, maxLifetime int, maxOpenConns int, maxIdleConns int) (*MySQLClient, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	db.SetConnMaxLifetime(time.Duration(maxLifetime) * time.Second)
	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &MySQLClient{DB: db}, nil
}

func (c *MySQLClient) Close() error {
	return c.DB.Close()
}

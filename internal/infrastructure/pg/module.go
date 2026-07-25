package pg

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/url"

	_ "github.com/lib/pq"
)

// Config — настройки подключения к PostgreSQL.
type Config struct {
	Host     string `envconfig:"HOST" default:"localhost"`
	Port     string `envconfig:"PORT" default:"5434"`
	User     string `envconfig:"USER" default:"postgres"`
	Password string `envconfig:"PASSWORD" default:"postgres"`
	DBName   string `envconfig:"NAME" default:"roadmap"`
	SSLMode  string `envconfig:"SSLMODE" default:"disable"`
}

// DSN возвращает строку подключения для lib/pq.
func (c *Config) DSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode)
}

// URL возвращает postgres:// URL для golang-migrate.
func (c *Config) URL() string {
	u := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(c.User, c.Password),
		Host:   net.JoinHostPort(c.Host, c.Port),
		Path:   c.DBName,
	}
	q := url.Values{}
	q.Set("sslmode", c.SSLMode)
	u.RawQuery = q.Encode()
	return u.String()
}

// DB — тонкая обёртка над database/sql (драйвер PostgreSQL).
type DB struct {
	conn *sql.DB
}

// New открывает пул соединений и проверяет его пингом.
func New(cfg *Config) (*DB, error) {
	conn, err := sql.Open("postgres", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("pg open: %w", err)
	}
	if err := conn.Ping(); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("pg ping: %w", err)
	}
	return &DB{conn: conn}, nil
}

// Close закрывает пул.
func (db *DB) Close() error {
	return db.conn.Close()
}

// Ping проверяет соединение.
func (db *DB) Ping(ctx context.Context) error {
	return db.conn.PingContext(ctx)
}

// ExecContext выполняет запрос без возврата строк.
func (db *DB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return db.conn.ExecContext(ctx, query, args...)
}

// QueryContext выполняет запрос и возвращает строки.
func (db *DB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return db.conn.QueryContext(ctx, query, args...)
}

// QueryRowContext выполняет запрос, ожидающий одну строку.
func (db *DB) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return db.conn.QueryRowContext(ctx, query, args...)
}

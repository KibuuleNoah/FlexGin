package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/joho/godotenv/autoload"
)

type Config struct {
	Env      string // dev | staging | prod
	LogLevel string
	HTTP     HTTP
	DB       DB
	CORS     CORS
}

type HTTP struct {
	Port            int
	ReadTimeout     time.Duration
	ReadHeaderTO    time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	RequestTimeout  time.Duration // per-request context deadline (middleware)
	ShutdownTimeout time.Duration
}

type DB struct {
	Host            string
	Port            int
	Name            string
	User            string
	Password        string
	Schema          string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

type CORS struct {
	AllowedOrigins   []string
	AllowCredentials bool
}

func (c Config) IsProd() bool { return c.Env == "prod" }

// DSN builds a pgx-compatible URL for database/sql.
func (d DB) DSN() string {
	u := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(d.User, d.Password),
		Host:   fmt.Sprintf("%s:%d", d.Host, d.Port),
		Path:   d.Name,
	}
	q := url.Values{}
	q.Set("sslmode", d.SSLMode)
	if d.Schema != "" {
		q.Set("search_path", d.Schema)
	}
	u.RawQuery = q.Encode()
	return u.String()
}

// Load reads and validates the environment. It reports every problem at once.
func Load() (Config, error) {
	var l loader

	cfg := Config{
		Env:      l.str("APP_ENV", "dev"),
		LogLevel: l.str("LOG_LEVEL", "info"),
		HTTP: HTTP{
			Port:            l.integer("PORT", 8080),
			ReadTimeout:     l.duration("HTTP_READ_TIMEOUT", 10*time.Second),
			ReadHeaderTO:    l.duration("HTTP_READ_HEADER_TIMEOUT", 5*time.Second),
			WriteTimeout:    l.duration("HTTP_WRITE_TIMEOUT", 30*time.Second),
			IdleTimeout:     l.duration("HTTP_IDLE_TIMEOUT", 60*time.Second),
			RequestTimeout:  l.duration("HTTP_REQUEST_TIMEOUT", 15*time.Second),
			ShutdownTimeout: l.duration("HTTP_SHUTDOWN_TIMEOUT", 15*time.Second),
		},
		DB: DB{
			Host:            l.required("DB_HOST"),
			Port:            l.integer("DB_PORT", 5432),
			Name:            l.required("DB_DATABASE"),
			User:            l.required("DB_USERNAME"),
			Password:        l.required("DB_PASSWORD"),
			Schema:          l.str("DB_SCHEMA", "public"),
			SSLMode:         l.str("DB_SSLMODE", "disable"),
			MaxOpenConns:    l.integer("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    l.integer("DB_MAX_IDLE_CONNS", 25),
			ConnMaxLifetime: l.duration("DB_CONN_MAX_LIFETIME", 30*time.Minute),
			ConnMaxIdleTime: l.duration("DB_CONN_MAX_IDLE_TIME", 5*time.Minute),
		},
		CORS: CORS{
			AllowedOrigins:   l.list("CORS_ALLOWED_ORIGINS", []string{"http://localhost:5173"}),
			AllowCredentials: l.boolean("CORS_ALLOW_CREDENTIALS", true),
		},
	}

	switch cfg.Env {
	case "dev", "staging", "prod":
	default:
		l.errs = append(l.errs, fmt.Errorf("APP_ENV: must be dev, staging or prod, got %q", cfg.Env))
	}
	if cfg.HTTP.Port < 1 || cfg.HTTP.Port > 65535 {
		l.errs = append(l.errs, fmt.Errorf("PORT: out of range: %d", cfg.HTTP.Port))
	}
	if cfg.IsProd() && cfg.DB.SSLMode == "disable" {
		l.errs = append(l.errs, errors.New("DB_SSLMODE: must not be 'disable' in prod"))
	}
	if cfg.IsProd() && containsWildcard(cfg.CORS.AllowedOrigins) {
		l.errs = append(l.errs, errors.New("CORS_ALLOWED_ORIGINS: wildcard not allowed in prod"))
	}

	if err := errors.Join(l.errs...); err != nil {
		return Config{}, fmt.Errorf("invalid config:\n%w", err)
	}
	return cfg, nil
}

func containsWildcard(s []string) bool {
	for _, v := range s {
		if v == "*" {
			return true
		}
	}
	return false
}

// loader accumulates parse errors so Load can report them together.
type loader struct{ errs []error }

func (l *loader) str(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}

func (l *loader) required(key string) string {
	v := os.Getenv(key)
	if v == "" {
		l.errs = append(l.errs, fmt.Errorf("%s: required", key))
	}
	return v
}

func (l *loader) integer(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		l.errs = append(l.errs, fmt.Errorf("%s: not an integer: %q", key, v))
		return def
	}
	return n
}

func (l *loader) boolean(key string, def bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		l.errs = append(l.errs, fmt.Errorf("%s: not a bool: %q", key, v))
		return def
	}
	return b
}

func (l *loader) duration(key string, def time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		l.errs = append(l.errs, fmt.Errorf("%s: not a duration: %q", key, v))
		return def
	}
	return d
}

func (l *loader) list(key string, def []string) []string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

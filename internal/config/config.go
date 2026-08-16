package config

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all application settings loaded from the environment.
type Config struct {
	App       App
	Log       Log
	DB        DB
	CH        CH
	Redis     Redis
	JWT       JWT
	GitHub    GitHub
	RateLimit RateLimit
}

// RateLimit contains request rate limiting settings.
type RateLimit struct {
	Global  int
	PerUser int
	Window  time.Duration
}

// App contains HTTP server settings.
type App struct {
	Env          string
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// Log contains logging settings.
type Log struct {
	Level string
}

// DB contains PostgreSQL connection settings.
type DB struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  string
}

// DSN returns the lib/pq-style connection string for PostgreSQL.
func (d DB) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode,
	)
}

// CH contains ClickHouse connection settings.
type CH struct {
	Host     string
	Port     int
	Database string
	User     string
	Password string
}

// DSN returns the clickhouse connection string used by the native driver.
func (c CH) DSN() string {
	return fmt.Sprintf(
		"clickhouse://%s:%s@%s:%d/%s",
		c.User, c.Password, c.Host, c.Port, c.Database,
	)
}

// Redis contains connection settings for Redis.
type Redis struct {
	Host     string
	Port     int
	Password string
	DB       int
}

// JWT contains token signing and lifetime settings.
type JWT struct {
	AccessSecret  string
	RefreshSecret string
	AccessTTL     time.Duration
	RefreshTTL    time.Duration
}

// GitHub contains OAuth application settings.
type GitHub struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

// Load reads configuration from environment variables and the given .env files.
func Load(paths ...string) (*Config, error) {
	if len(paths) > 0 {
		if err := godotenv.Load(paths...); err != nil {
			return nil, fmt.Errorf("load env file: %w", err)
		}
	} else {
		_ = godotenv.Load()
	}

	cfg := &Config{}

	cfg.App = App{
		Env:          get("APP_ENV", "development"),
		Port:         getInt("APP_PORT", 8080),
		ReadTimeout:  getDuration("APP_READ_TIMEOUT", 10*time.Second),
		WriteTimeout: getDuration("APP_WRITE_TIMEOUT", 30*time.Second),
		IdleTimeout:  getDuration("APP_IDLE_TIMEOUT", 60*time.Second),
	}

	cfg.Log = Log{
		Level: get("LOG_LEVEL", "info"),
	}

	cfg.DB = DB{
		Host:     get("POSTGRES_HOST", "localhost"),
		Port:     getInt("POSTGRES_PORT", 5432),
		User:     get("POSTGRES_USER", "levelup"),
		Password: get("POSTGRES_PASSWORD", "levelup"),
		Name:     get("POSTGRES_DB", "levelup"),
		SSLMode:  get("POSTGRES_SSLMODE", "disable"),
	}

	cfg.CH = CH{
		Host:     get("CLICKHOUSE_HOST", "localhost"),
		Port:     getInt("CLICKHOUSE_PORT", 9000),
		Database: get("CLICKHOUSE_DATABASE", "levelup"),
		User:     get("CLICKHOUSE_USER", "default"),
		Password: get("CLICKHOUSE_PASSWORD", ""),
	}

	cfg.Redis = Redis{
		Host:     get("REDIS_HOST", "localhost"),
		Port:     getInt("REDIS_PORT", 6379),
		Password: get("REDIS_PASSWORD", ""),
		DB:       getInt("REDIS_DB", 0),
	}

	cfg.JWT = JWT{
		AccessSecret:  get("JWT_ACCESS_SECRET", "change-me-access-secret"),
		RefreshSecret: get("JWT_REFRESH_SECRET", "change-me-refresh-secret"),
		AccessTTL:     getDuration("JWT_ACCESS_TTL", 15*time.Minute),
		RefreshTTL:    getDuration("JWT_REFRESH_TTL", 720*time.Hour),
	}

	cfg.GitHub = GitHub{
		ClientID:     get("GITHUB_CLIENT_ID", ""),
		ClientSecret: get("GITHUB_CLIENT_SECRET", ""),
		RedirectURL:  get("GITHUB_REDIRECT_URL", ""),
	}

	cfg.RateLimit = RateLimit{
		Global:  getInt("RATE_LIMIT_GLOBAL", 2000),
		PerUser: getInt("RATE_LIMIT_PER_USER", 120),
		Window:  getDuration("RATE_LIMIT_WINDOW", time.Minute),
	}

	return cfg, nil
}

func get(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		var n int
		if _, err := fmt.Sscanf(v, "%d", &n); err == nil {
			return n
		}
	}
	return fallback
}

func getDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}

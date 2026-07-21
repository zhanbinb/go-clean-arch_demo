// Package config loads runtime configuration via Viper from YAML files
// and environment variables. Supports multi-environment overrides via
// APP_ENV (e.g. APP_ENV=local → loads config.local.yaml on top).
package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config is the root configuration.
type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	Database  DatabaseConfig  `mapstructure:"database"`
	JWT       JWTConfig       `mapstructure:"jwt"`
	Log       LogConfig       `mapstructure:"log"`
	CORS      CORSConfig      `mapstructure:"cors"`
	RateLimit RateLimitConfig `mapstructure:"ratelimit"`
	Swagger   SwaggerConfig   `mapstructure:"swagger"`
}

// ServerConfig holds HTTP + gRPC server settings.
type ServerConfig struct {
	HTTPPort int    `mapstructure:"http_port"`
	GRPCPort int    `mapstructure:"grpc_port"`
	Mode     string `mapstructure:"mode"` // debug | release | test
}

// HTTPAddr returns the HTTP listen address.
func (s ServerConfig) HTTPAddr() string { return fmt.Sprintf(":%d", s.HTTPPort) }

// GRPCAddr returns the gRPC listen address.
func (s ServerConfig) GRPCAddr() string { return fmt.Sprintf(":%d", s.GRPCPort) }

// DatabaseConfig holds MySQL connection settings.
type DatabaseConfig struct {
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	User            string `mapstructure:"user"`
	Password        string `mapstructure:"password"`
	Name            string `mapstructure:"name"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"` // seconds
	LogLevel        string `mapstructure:"log_level"`         // silent | error | warn | info
}

// DSN builds a GORM-compatible MySQL DSN.
func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=true&loc=Local",
		d.User, d.Password, d.Host, d.Port, d.Name)
}

// JWTConfig holds JWT signing settings.
type JWTConfig struct {
	Secret     string        `mapstructure:"secret"`
	TTL        time.Duration `mapstructure:"ttl"`
	RefreshTTL time.Duration `mapstructure:"refresh_ttl"`
}

// LogConfig holds logger settings.
type LogConfig struct {
	Level  string `mapstructure:"level"`  // debug | info | warn | error
	Format string `mapstructure:"format"` // json | console
}

// CORSConfig holds CORS settings.
type CORSConfig struct {
	AllowOrigins     []string `mapstructure:"allow_origins"`
	AllowMethods     []string `mapstructure:"allow_methods"`
	AllowHeaders     []string `mapstructure:"allow_headers"`
	ExposeHeaders    []string `mapstructure:"expose_headers"`
	AllowCredentials bool     `mapstructure:"allow_credentials"`
	MaxAge           int      `mapstructure:"max_age"` // seconds
}

// RateLimitConfig holds rate-limiter settings.
type RateLimitConfig struct {
	Enabled         bool    `mapstructure:"enabled"`
	RPS             float64 `mapstructure:"rps"`
	Burst           int     `mapstructure:"burst"`
	CleanupInterval int     `mapstructure:"cleanup_interval"` // seconds
	Dimension       string  `mapstructure:"dimension"`        // ip | user
}

// SwaggerConfig holds Swagger UI settings.
type SwaggerConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Path    string `mapstructure:"path"`
}

// Load reads configs/config.yaml plus the optional config.<env>.yaml override,
// then applies environment variables prefixed with APP_.
func Load(env string) (*Config, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("./configs")
	v.AddConfigPath(".")

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config.yaml: %w", err)
	}
	if env != "" {
		v.SetConfigName("config." + env)
		if err := v.MergeInConfig(); err != nil {
			// env file is optional
			if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
				return nil, fmt.Errorf("merge config.%s.yaml: %w", env, err)
			}
		}
	}

	v.SetEnvPrefix("APP")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	if cfg.Database.MaxOpenConns == 0 {
		cfg.Database.MaxOpenConns = 50
	}
	if cfg.Database.MaxIdleConns == 0 {
		cfg.Database.MaxIdleConns = 10
	}
	if cfg.Database.ConnMaxLifetime == 0 {
		cfg.Database.ConnMaxLifetime = 3600
	}
	if cfg.JWT.TTL == 0 {
		cfg.JWT.TTL = time.Hour
	}
	if cfg.JWT.RefreshTTL == 0 {
		cfg.JWT.RefreshTTL = 24 * time.Hour
	}
	return &cfg, nil
}

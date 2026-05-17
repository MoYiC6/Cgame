package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	App             AppConfig             `yaml:"app"`
	Server          ServerConfig          `yaml:"server"`
	Log             LogConfig             `yaml:"log"`
	DB              DBConfig              `yaml:"db"`
	Redis           RedisConfig           `yaml:"redis"`
	MQ              MQConfig              `yaml:"mq"`
	Observability   ObservabilityConfig   `yaml:"observability"`
	CORS            CORSConfig            `yaml:"cors"`
	SecurityHeaders SecurityHeadersConfig `yaml:"security_headers"`
	RateLimit       RateLimitConfig       `yaml:"rate_limit"`
	Metrics         MetricsConfig         `yaml:"metrics"`
}

type AppConfig struct {
	Name string `yaml:"name"`
	Env  string `yaml:"env"`
}

type ServerConfig struct {
	Addr string `yaml:"addr"`
}

type LogConfig struct {
	Level string `yaml:"level"`
}

type DBConfig struct {
	Driver              string `yaml:"driver"`
	DSN                 string `yaml:"dsn"`
	MaxOpenConns        int    `yaml:"max_open_conns"`
	MaxIdleConns        int    `yaml:"max_idle_conns"`
	ConnMaxLifetimeSecs int    `yaml:"conn_max_lifetime_secs"`
}

type ObservabilityConfig struct {
	TraceExporterType     string `yaml:"trace_exporter_type"`
	TraceExporterEndpoint string `yaml:"trace_exporter_endpoint"`
	ServiceName           string `yaml:"service_name"`
	ServiceVersion        string `yaml:"service_version"`
	Environment           string `yaml:"environment"`
}

type RedisConfig struct {
	Addr string `yaml:"addr"`
}

type MQConfig struct {
	Driver      string `yaml:"driver"`
	TopicPrefix string `yaml:"topic_prefix"`
}

type MetricsConfig struct {
	Enabled bool `yaml:"enabled"`
}

type CORSConfig struct {
	AllowedOrigins   []string `yaml:"allowed_origins"`
	AllowedMethods   []string `yaml:"allowed_methods"`
	AllowedHeaders   []string `yaml:"allowed_headers"`
	AllowCredentials bool     `yaml:"allow_credentials"`
	MaxAgeSecs       int      `yaml:"max_age_secs"`
}

type SecurityHeadersConfig struct {
	FrameOptions       string `yaml:"frame_options"`
	ContentTypeOptions bool   `yaml:"content_type_options"`
	ReferrerPolicy     string `yaml:"referrer_policy"`
}

type RateLimitConfig struct {
	Requests   int `yaml:"requests"`
	WindowSecs int `yaml:"window_secs"`
}

func Load(path string) (Config, error) {
	return loadFromPath(path)
}

func LoadConfig(env string) (*Config, error) {
	env = normalizeEnv(env)
	path := os.Getenv("APP_CONFIG_PATH")
	if strings.TrimSpace(path) == "" {
		path = filepath.Join("configs", fmt.Sprintf("config.%s.yaml", env))
	}

	cfg, err := loadFromPath(path)
	if err != nil {
		return nil, err
	}
	applyEnvOverrides(&cfg)
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func loadFromPath(path string) (Config, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(content, &cfg); err != nil {
		return Config{}, fmt.Errorf("unmarshal config: %w", err)
	}

	if cfg.App.Env == "" {
		cfg.App.Env = "local"
	}
	if cfg.Log.Level == "" {
		cfg.Log.Level = "info"
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func normalizeEnv(env string) string {
	env = strings.TrimSpace(env)
	if env == "" {
		env = strings.TrimSpace(os.Getenv("APP_ENV"))
	}
	if env == "" {
		return "local"
	}
	return env
}

func applyEnvOverrides(cfg *Config) {
	if cfg == nil {
		return
	}
	if value := os.Getenv("APP_NAME"); strings.TrimSpace(value) != "" {
		cfg.App.Name = value
	}
	if value := os.Getenv("APP_ENV"); strings.TrimSpace(value) != "" {
		cfg.App.Env = value
	}
	if value := os.Getenv("SERVER_ADDR"); strings.TrimSpace(value) != "" {
		cfg.Server.Addr = value
	}
	if value := os.Getenv("LOG_LEVEL"); strings.TrimSpace(value) != "" {
		cfg.Log.Level = value
	}
	if value := os.Getenv("DB_DRIVER"); strings.TrimSpace(value) != "" {
		cfg.DB.Driver = value
	}
	if value := os.Getenv("DB_DSN"); strings.TrimSpace(value) != "" {
		cfg.DB.DSN = value
	}
	if value := os.Getenv("DB_MAX_OPEN_CONNS"); strings.TrimSpace(value) != "" {
		if parsed, err := strconv.Atoi(strings.TrimSpace(value)); err == nil {
			cfg.DB.MaxOpenConns = parsed
		}
	}
	if value := os.Getenv("DB_MAX_IDLE_CONNS"); strings.TrimSpace(value) != "" {
		if parsed, err := strconv.Atoi(strings.TrimSpace(value)); err == nil {
			cfg.DB.MaxIdleConns = parsed
		}
	}
	if value := os.Getenv("DB_CONN_MAX_LIFETIME_SECS"); strings.TrimSpace(value) != "" {
		if parsed, err := strconv.Atoi(strings.TrimSpace(value)); err == nil {
			cfg.DB.ConnMaxLifetimeSecs = parsed
		}
	}
	if value := os.Getenv("OTEL_TRACE_EXPORTER_TYPE"); strings.TrimSpace(value) != "" {
		cfg.Observability.TraceExporterType = value
	}
	if value := os.Getenv("OTEL_TRACE_EXPORTER_ENDPOINT"); strings.TrimSpace(value) != "" {
		cfg.Observability.TraceExporterEndpoint = value
	}
	if value := os.Getenv("OTEL_SERVICE_NAME"); strings.TrimSpace(value) != "" {
		cfg.Observability.ServiceName = value
	}
	if value := os.Getenv("OTEL_SERVICE_VERSION"); strings.TrimSpace(value) != "" {
		cfg.Observability.ServiceVersion = value
	}
	if value := os.Getenv("OTEL_ENVIRONMENT"); strings.TrimSpace(value) != "" {
		cfg.Observability.Environment = value
	}
	if value := os.Getenv("MQ_DRIVER"); strings.TrimSpace(value) != "" {
		cfg.MQ.Driver = value
	}
	if value := os.Getenv("MQ_TOPIC_PREFIX"); strings.TrimSpace(value) != "" {
		cfg.MQ.TopicPrefix = value
	}
}

func (c Config) Validate() error {
	if strings.TrimSpace(c.App.Name) == "" {
		return fmt.Errorf("app.name is required")
	}
	if strings.TrimSpace(c.Server.Addr) == "" {
		return fmt.Errorf("server.addr is required")
	}
	if c.DB.MaxOpenConns < 0 {
		return fmt.Errorf("db.max_open_conns must be >= 0")
	}
	if c.DB.MaxIdleConns < 0 {
		return fmt.Errorf("db.max_idle_conns must be >= 0")
	}
	if c.DB.ConnMaxLifetimeSecs < 0 {
		return fmt.Errorf("db.conn_max_lifetime_secs must be >= 0")
	}
	if c.DB.MaxOpenConns > 0 && c.DB.MaxIdleConns > c.DB.MaxOpenConns {
		return fmt.Errorf("db.max_idle_conns must be <= db.max_open_conns")
	}
	traceExporterType := strings.TrimSpace(c.Observability.TraceExporterType)
	if traceExporterType != "" && traceExporterType != "none" && traceExporterType != "otlp" {
		return fmt.Errorf("observability.trace_exporter_type must be one of: none, otlp")
	}
	if traceExporterType == "otlp" && strings.TrimSpace(c.Observability.TraceExporterEndpoint) == "" {
		return fmt.Errorf("observability.trace_exporter_endpoint is required when trace exporter type is otlp")
	}
	if strings.TrimSpace(c.Observability.ServiceName) == "" {
		return fmt.Errorf("observability.service_name is required")
	}
	return nil
}

func (c Config) MaskedSummary() map[string]string {
	return map[string]string{
		"app_name":                  c.App.Name,
		"app_env":                   c.App.Env,
		"server":                    c.Server.Addr,
		"log_level":                 c.Log.Level,
		"db_driver":                 c.DB.Driver,
		"db_dsn":                    maskSecret(c.DB.DSN),
		"db_max_open_conns":         strconv.Itoa(c.DB.MaxOpenConns),
		"db_max_idle_conns":         strconv.Itoa(c.DB.MaxIdleConns),
		"db_conn_max_lifetime_secs": strconv.Itoa(c.DB.ConnMaxLifetimeSecs),
		"redis":                     maskSecret(c.Redis.Addr),
		"mq_driver":                 c.MQ.Driver,
		"mq_topic":                  c.MQ.TopicPrefix,
		"otel_trace_exporter_type":  c.Observability.TraceExporterType,
		"otel_exporter_endpoint":    maskSecret(c.Observability.TraceExporterEndpoint),
		"otel_service_name":         c.Observability.ServiceName,
		"otel_service_version":      c.Observability.ServiceVersion,
		"otel_environment":          c.Observability.Environment,
		"metrics_enabled":           strconv.FormatBool(c.Metrics.Enabled),
		"cors_allowed_origins":      strings.Join(c.CORS.AllowedOrigins, ","),
		"rate_limit_requests":       strconv.Itoa(c.RateLimit.Requests),
		"rate_limit_window_secs":    strconv.Itoa(c.RateLimit.WindowSecs),
	}
}

func maskSecret(value string) string {
	if value == "" {
		return ""
	}
	if len(value) <= 4 {
		return "****"
	}
	return value[:2] + strings.Repeat("*", len(value)-4) + value[len(value)-2:]
}

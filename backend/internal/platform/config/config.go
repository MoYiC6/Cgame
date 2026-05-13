package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	App    AppConfig    `yaml:"app"`
	Server ServerConfig `yaml:"server"`
	Log    LogConfig    `yaml:"log"`
	DB     DBConfig     `yaml:"db"`
	Redis  RedisConfig  `yaml:"redis"`
	MQ     MQConfig     `yaml:"mq"`
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
	Driver string `yaml:"driver"`
	DSN    string `yaml:"dsn"`
}

type RedisConfig struct {
	Addr string `yaml:"addr"`
}

type MQConfig struct {
	Driver      string `yaml:"driver"`
	TopicPrefix string `yaml:"topic_prefix"`
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
	if value := os.Getenv("REDIS_ADDR"); strings.TrimSpace(value) != "" {
		cfg.Redis.Addr = value
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
	return nil
}

func (c Config) MaskedSummary() map[string]string {
	return map[string]string{
		"app_name":  c.App.Name,
		"app_env":   c.App.Env,
		"server":    c.Server.Addr,
		"log_level": c.Log.Level,
		"db_driver": c.DB.Driver,
		"db_dsn":    maskSecret(c.DB.DSN),
		"redis":     maskSecret(c.Redis.Addr),
		"mq_driver": c.MQ.Driver,
		"mq_topic":  c.MQ.TopicPrefix,
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

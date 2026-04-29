package config

import (
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server       ServerConfig       `mapstructure:"server"`
	Database     DatabaseConfig     `mapstructure:"database"`
	Redis        RedisConfig        `mapstructure:"redis"`
	Milvus       MilvusConfig       `mapstructure:"milvus"`
	Storage      StorageConfig      `mapstructure:"storage"`
	LLM          LLMConfig          `mapstructure:"llm"`
	Worker       WorkerConfig       `mapstructure:"worker"`
	XFileStorage XFileStorageConfig `mapstructure:"xfile_storage"`
	Log          LogConfig          `mapstructure:"log"`
}

type LogConfig struct {
	Level    string `mapstructure:"level"`
	Output   string `mapstructure:"output"`
	FilePath string `mapstructure:"file_path"`
}

type XFileStorageConfig struct {
	BaseURL string `mapstructure:"base_url"`
	APIKey  string `mapstructure:"api_key"`
}

type ServerConfig struct {
	Port             int             `mapstructure:"port"`
	GRPCPort         int             `mapstructure:"grpc_port"`
	JWTSecret        string          `mapstructure:"jwt_secret"`
	CredentialSecret string          `mapstructure:"credential_secret"`
	RateLimit        RateLimitConfig `mapstructure:"rate_limit"`
}

type RateLimitConfig struct {
	GlobalRate  float64 `mapstructure:"global_rate"`
	GlobalBurst int     `mapstructure:"global_burst"`
	IPRate      float64 `mapstructure:"ip_rate"`
	IPBurst     int     `mapstructure:"ip_burst"`
}

type DatabaseConfig struct {
	Driver       string `mapstructure:"driver"`
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	User         string `mapstructure:"user"`
	Password     string `mapstructure:"password"`
	DBName       string `mapstructure:"dbname"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
	LogLevel     string `mapstructure:"log_level"`
}

type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

type MilvusConfig struct {
	Addr       string `mapstructure:"addr"`
	Collection string `mapstructure:"collection"`
	Dimension  int    `mapstructure:"dimension"`
}

type StorageConfig struct {
	Type       string `mapstructure:"type"`
	LocalPath  string `mapstructure:"local_path"`
	S3Bucket   string `mapstructure:"s3_bucket"`
	S3Region   string `mapstructure:"s3_region"`
	S3Endpoint string `mapstructure:"s3_endpoint"`
}

type LLMConfig struct {
	Providers       map[string]LLMProviderConfig `mapstructure:"providers"`
	DefaultProvider string                       `mapstructure:"default_provider"`
	MaxRetries      int                          `mapstructure:"max_retries"`
	Timeout         int                          `mapstructure:"timeout"`
}

type LLMProviderConfig struct {
	APIKey       string `mapstructure:"api_key"`
	BaseURL      string `mapstructure:"base_url"`
	DefaultModel string `mapstructure:"default_model"`
}

type WorkerConfig struct {
	Concurrency int `mapstructure:"concurrency"`
	MaxRetry    int `mapstructure:"max_retry"`
}

var C *Config

func Load(path string) error {
	viper.SetConfigFile(path)
	viper.SetEnvPrefix("NEXUS")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		return err
	}
	C = &Config{}
	return viper.Unmarshal(C)
}

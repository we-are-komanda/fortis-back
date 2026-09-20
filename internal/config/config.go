package config

import (
	"fmt"
	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v3"
	"log/slog"
)

type Documents struct {
	Enabled           bool   `yaml:"enabled"`
	RootDir           string `yaml:"root_dir" envconfig:"root_dir"`
	ScannerExecutable string `yaml:"scanner_executable" envconfig:"scanner_executable"`
}

type Config struct {
	Documents    Documents    `yaml:"documents" envconfig:"documents"`
	WorkMode     WorkMode     `yaml:"work_mode" envconfig:"work_mode"`
	Postgres     Postgres     `envconfig:"db"`
	Access       Access       `envconfig:"access"`
	Cors         Cors         `yaml:"cors" envconfig:"cors"`
	Auth         Auth         `yaml:"auth" envconfig:"auth"`
	DemoRequests DemoRequests `yaml:"demo_requests" envconfig:"demo"`
	Environment  string
}

func NewConfig(env string) *Config {
	return &Config{
		Environment: env,
	}
}

// Disabled by default. Recipient and credentials are server configuration only.
type DemoRequests struct {
	Enabled           bool     `yaml:"enabled"`
	ConsentVersions   []string `yaml:"consent_versions" envconfig:"consent_versions"`
	AllowedOrigins    []string `yaml:"allowed_origins" envconfig:"allowed_origins"`
	TrustedProxyCIDRs []string `yaml:"trusted_proxy_cidrs" envconfig:"trusted_proxy_cidrs"`
	SMTPAddress       string   `yaml:"smtp_address" envconfig:"smtp_address"`
	SMTPUsername      string   `yaml:"smtp_username" envconfig:"smtp_username"`
	SMTPPassword      string   `yaml:"-" envconfig:"smtp_password"`
	SMTPFrom          string   `yaml:"smtp_from" envconfig:"smtp_from"`
	SMTPTo            string   `yaml:"smtp_to" envconfig:"smtp_to"`
}

type WorkMode string

type Postgres struct {
	Host           string `yaml:"host"`
	User           string `yaml:"user"`
	Password       string `yaml:"password"`
	DbName         string `yaml:"dbname"`
	Port           int    `yaml:"port"`
	RequestTimeout int    `yaml:"request_timeout"`
	MaxOpenSlots   int    `yaml:"max_open_slots"`
}

type Access struct {
	AuthUrl          string   `yaml:"auth_url"`
	ValidityEndpoint string   `yaml:"validity_endpoint"`
	WhiteList        []string `yaml:"whitelist"`
	ReTry            int      `yaml:"retry"`
}

type Auth struct {
	JWTSecret string `yaml:"jwt_secret"`
	JWTExpiry int    `yaml:"jwt_expiry"` // hours
}

type Cors struct {
	Enable           bool     `yaml:"enable"`
	AllowCredentials bool     `yaml:"allow_credentials"`
	OriginRegex      bool     `yaml:"origin_regex"`
	AllowOrigins     []string `yaml:"allow_origins"`
	AllowMethods     []string `yaml:"allow_methods"`
	AllowHeaders     []string `yaml:"allow_headers"`
	ExposeHeaders    []string `yaml:"expose_headers"`
	MaxAge           int      `yaml:"max_age"`
}

func (cnf *Config) ReadConfig(configPath string, readerFunc func(string) ([]byte, error)) error {
	slog.Info(fmt.Sprintf("environment=%s", cnf.Environment))

	configRaw, err := readerFunc(fmt.Sprintf("%s/config.%s.yml", configPath, cnf.Environment))
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(configRaw, cnf)
	if err != nil {
		return fmt.Errorf("config parse error: %w", err)
	}

	err = envconfig.Process("APP", cnf)
	if err != nil {
		return fmt.Errorf("env config parse error: %w", err)
	}

	return nil
}

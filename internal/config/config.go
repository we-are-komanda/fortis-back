package config

import (
	"fmt"
	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v3"
	"log/slog"
)

type Config struct {
	WorkMode    WorkMode `yaml:"work_mode" envconfig:"work_mode"`
	Postgres    Postgres `envconfig:"db"`
	Access      Access   `envconfig:"access"`
	Cors        Cors     `yaml:"cors" envconfig:"cors"`
	Auth        Auth     `yaml:"auth" envconfig:"auth"`
	Environment string
}

func NewConfig(env string) *Config {
	return &Config{
		Environment: env,
	}
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

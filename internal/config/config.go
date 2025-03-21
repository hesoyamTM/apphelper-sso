package config

import (
	"errors"
	"flag"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Env        string `yaml:"env" env-required:"true"`
	PrivateKey string `yaml:"public_key" env-required:"true" env:"PRIVATE_KEY"`

	Grpc  GRPC  `yaml:"grpc"`
	Psql  PSQL  `yaml:"psql"`
	Redis Redis `yaml:"redis"`

	Report   ReportClient   `yaml:"report_client"`
	Schedule ScheduleClient `yaml:"schedule_client"`
}

type GRPC struct {
	Host            string        `yaml:"host" env-required:"true"`
	Port            int           `yaml:"port" env-required:"true"`
	AccessTokenTTL  time.Duration `yaml:"access_token_ttl" env-required:"true"`
	RefreshTokenTTL time.Duration `yaml:"refresh_token_ttl" env-required:"true"`
}

type PSQL struct {
	Host     string `yaml:"host" env-required:"true"`
	Port     int    `yaml:"port" env-required:"true"`
	User     string `yaml:"user" env-required:"true"`
	Password string `yaml:"password" env-required:"true"`
	DB       string `yaml:"db" env-required:"true"`
}

type Redis struct {
	Host     string `yaml:"host" env-required:"true"`
	Port     int    `yaml:"port" env-required:"true"`
	Password string `yaml:"password" env-required:"true"`
}

type ReportClient struct {
	Addr string `yaml:"addr" env-required:"true"`
}

type ScheduleClient struct {
	Addr string `yaml:"addr" env-required:"true"`
}

func MustLoad() *Config {
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	cfgPath := fetchConfigPath()

	if cfgPath == "" {
		panic("config path is empty")
	}

	return MustLoadByPath(cfgPath)
}

func MustLoadByPath(cfgPath string) *Config {
	var cfg Config

	if _, err := os.Stat(cfgPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			panic("file does not exist: " + cfgPath)
		}
		panic(err)
	}

	if err := cleanenv.ReadConfig(cfgPath, &cfg); err != nil {
		panic("failed to read config: " + err.Error())
	}

	return &cfg
}

func fetchConfigPath() string {
	var res string

	flag.StringVar(&res, "config", "", "path to config file")
	flag.Parse()

	if res == "" {
		res = os.Getenv("CONFIG_PATH")
	}

	return res
}

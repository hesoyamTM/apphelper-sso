package config

import (
	"errors"
	"flag"
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Env        string `yaml:"env" env-required:"true" env:"ENV"`
	PrivateKey string `yaml:"public_key" env-required:"true" env:"PRIVATE_KEY"`

	AccessTokenTTL  time.Duration `yaml:"access_token_ttl" env-required:"true" env:"ACCESS_TOKEN_TTL"`
	RefreshTokenTTL time.Duration `yaml:"refresh_token_ttl" env-required:"true" env:"REFRESH_TOKEN_TTL"`

	Grpc  GRPC  `yaml:"grpc"`
	Psql  PSQL  `yaml:"psql"`
	Redis Redis `yaml:"redis"`
}

type GRPC struct {
	Host string `yaml:"host" env-required:"true" env:"GRPC_HOST"`
	Port int    `yaml:"port" env-required:"true" env:"GRPC_PORT"`
}

type PSQL struct {
	Host     string `yaml:"host" env-required:"true" env:"PSQL_HOST"`
	Port     int    `yaml:"port" env-required:"true" env:"PSQL_PORT"`
	User     string `yaml:"user" env-required:"true" env:"PSQL_USER"`
	Password string `yaml:"password" env-required:"true" env:"PSQL_PASSWORD"`
	DB       string `yaml:"db" env-required:"true" env:"PSQL_DATABASE"`
}

type Redis struct {
	Host     string `yaml:"host" env-required:"true" env:"REDIS_HOST"`
	Port     int    `yaml:"port" env-required:"true" env:"REDIS_PORT"`
	Password string `yaml:"password" env-required:"true" env:"REDIS_PASSWORD"`
}

func fetchConfigPath() string {
	var cfgPath string

	flag.StringVar(&cfgPath, "config", "", "config path")
	flag.Parse()

	if cfgPath == "" {
		cfgPath = os.Getenv("CONFIG_PATH")
	}

	return cfgPath
}

func MustLoad() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("failed to load .env")
	}

	cfgPath := fetchConfigPath()
	if cfgPath != "" {
		return MustLoadByPath(cfgPath)
	}

	return MustLoadEnv()
}

func MustLoadEnv() *Config {
	var cfg Config

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		panic(err)
	}

	return &cfg
}

func MustLoadByPath(cfgPath string) *Config {
	if cfgPath == "" {
		panic("config path is empty")
	}

	if _, err := os.Stat(cfgPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			panic("config file does not exist: " + err.Error())
		}

		panic(err)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(cfgPath, &cfg); err != nil {
		panic(err)
	}

	return &cfg
}

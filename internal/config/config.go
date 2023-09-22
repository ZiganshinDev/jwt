package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env        string `yaml:"env" env-default:"local"`
	HTTPServer `yaml:"http_server"`
	Mongo
	JWT `yaml:"jwt"`
}

type HTTPServer struct {
	Address     string        `yaml:"address" env-default:"localhost:8080"`
	Timeout     time.Duration `yaml:"timeout" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
}

type Mongo struct {
	URI      string
	User     string
	Password string
	Name     string
	Database string
}

type JWT struct {
	AccessTokenTTL  time.Duration `yaml:"access_token_ttl"`
	RefreshTokenTTL time.Duration `yaml:"refresh_token_ttl"`
	SigningKey      string
}

func MustLoad() *Config {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("CONFIG_PATH is not set")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config: %s", err)
	}

	setFromEnv(&cfg)

	return &cfg
}

func setFromEnv(cfg *Config) {
	if cfg.Env == "local" {
		cfg.Mongo.URI = "mongodb://localhost:27017"
	} else {
		cfg.Mongo.URI = os.Getenv("MONGO_URI")
	}
	// cfg.Mongo.User = os.Getenv("MONGO_USER")
	// cfg.Mongo.Password = os.Getenv("MONGO_PASSWORD")
	cfg.Mongo.Database = os.Getenv("MONGO_DATABASE")

	cfg.JWT.SigningKey = os.Getenv("JWT_SIGNING_KEY")
}

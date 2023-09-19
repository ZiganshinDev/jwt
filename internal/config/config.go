package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env         string `yaml:"env" env-default:"local"`
	StoragePath string `yaml:"storage_path" env-required:"true"`
	HTTPServer  `yaml:"http_server"`
	Mongo       `yaml:"mongo"`
	Auth        `yaml:"auth"`
}

type HTTPServer struct {
	Address     string        `yaml:"address" env-default:"localhost:8080"`
	Timeout     time.Duration `yaml:"timeout" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
}

// TODO change to yaml
type Mongo struct {
	URI      string
	User     string
	Password string
	Name     string
}

type Auth struct {
	JWT                    JWTConfig `yaml:"jwt"`
	PasswordSalt           string
	VerificationCodeLength int
}

type JWTConfig struct {
	AccessTokenTTL  time.Duration `yaml:"access_token_ttl"`
	RefreshTokenTTL time.Duration `yaml:"refresh_token_ttl"`
	SigningKey      string
}

func MustLoad() *Config {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("CONFIG_PATH is not set")
	}

	// check if file exists
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
	cfg.Mongo.URI = os.Getenv("MONGO_URI")

	cfg.Auth.JWT.SigningKey = os.Getenv("JWT_SIGNING_KEY")
}

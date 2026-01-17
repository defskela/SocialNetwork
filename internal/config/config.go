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
	Postgres   `yaml:"postgres"`
	JWT        `yaml:"jwt"`
}

type HTTPServer struct {
	Address     string        `yaml:"address" env:"HTTP_SERVER_ADDRESS" env-default:"localhost:8080"`
	Timeout     time.Duration `yaml:"timeout" env:"HTTP_SERVER_TIMEOUT" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env:"HTTP_SERVER_IDLE_TIMEOUT" env-default:"60s"`
}

type Postgres struct {
	Host     string `yaml:"host" env:"POSTGRES_HOST" env-default:"localhost"`
	Port     int    `yaml:"port" env:"POSTGRES_PORT" env-required:"true"`
	User     string `yaml:"user" env:"POSTGRES_USER" env-required:"true"`
	Password string `yaml:"password" env:"POSTGRES_PASSWORD" env-required:"true"`
	DBName   string `yaml:"db_name" env:"POSTGRES_DB" env-required:"true"`
	SSLMode  string `yaml:"ssl_mode" env:"POSTGRES_SSL_MODE" env-default:"disable"`
}

type JWT struct {
	PrivateKeyPath string `yaml:"private_key_path" env:"JWT_PRIVATE_KEY_PATH" env-required:"true"`
	PublicKeyPath  string `yaml:"public_key_path" env:"JWT_PUBLIC_KEY_PATH" env-required:"true"`
}

func MustLoad() *Config {
	var cfg Config

	// Try loading .env file if it exists
	if _, err := os.Stat(".env"); err == nil {
		if err := cleanenv.ReadConfig(".env", &cfg); err != nil {
			log.Printf("cannot read .env file: %s", err)
		}
	}

	// Override with environment variables
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Fatalf("cannot read environment variables: %s", err)
	}

	return &cfg
}

func MustLoadPath(configPath string) *Config {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config: %s", err)
	}

	// Override with environment variables
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Fatalf("cannot read environment variables: %s", err)
	}

	return &cfg
}

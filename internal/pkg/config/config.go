package config

import (
	"log"
	"os"
	"time"

	"github.com/drone/envsubst"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Env                string         `yaml:"env"`
	PostgresConnString string         `yaml:"postgres_conn_string"`
	HTTPServer         HTTPServer     `yaml:"http_server"`
	Log                Log            `yaml:"log"`
	JWT                JWT            `yaml:"jwt"`
	GoogleOAuth        GoogleOAuth    `yaml:"google_oauth"`
	SMTP               SMTP           `yaml:"smtp"`
	Redis              RedisConfig    `yaml:"redis"`
	RabbitMQ           RabbitMQConfig `yaml:"rabbitmq"`
	MinioConfig        MinioConfig    `yaml:"minio"`
}

type HTTPServer struct {
	Address     string        `yaml:"address"`
	Timeout     time.Duration `yaml:"timeout"`
	IdleTimeout time.Duration `yaml:"idle_timeout"`
}

type Log struct {
	Level      string            `yaml:"level"`
	File       string            `yaml:"file"`
	LokiURL    string            `yaml:"loki_url"`
	LokiLabels map[string]string `yaml:"loki_labels"`
}

type JWT struct {
	Secret string `yaml:"secret"`
}

type GoogleOAuth struct {
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
	RedirectURL  string `yaml:"redirect_url"`
}

type SMTP struct {
	Host        string `yaml:"host"`
	Port        string `yaml:"port"`
	User        string `yaml:"user"`
	Password    string `yaml:"password"`
	FromName    string `yaml:"from_name"`
	FromAddress string `yaml:"from_address"`
}

type RedisConfig struct {
	Address  string `yaml:"address"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type RabbitMQConfig struct {
	URL       string `yaml:"url"`
	QueueName string `yaml:"queue_name"`
}

type MinioConfig struct {
	Endpoint  string `yaml:"endpoint"`
	AccessKey string `yaml:"access_key"`
	SecretKey string `yaml:"secret_key"`
	UseSSL    bool   `yaml:"use_ssl"`
}

func MustLoad() *Config {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config/config.yaml"
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatalf("cannot read config file: %s", err)
	}

	expandedConfig, err := envsubst.EvalEnv(string(data))
	if err != nil {
		log.Fatalf("cannot substitute env variables: %s", err)
	}

	var cfg Config

	if err := yaml.Unmarshal([]byte(expandedConfig), &cfg); err != nil {
		log.Fatalf("cannot parse config: %s", err)
	}

	return &cfg
}

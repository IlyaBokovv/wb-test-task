package configs

import (
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Db    DbConfig
	Kafka KafkaConfig
	Http  HttpConfig
}

type DbConfig struct {
	Dsn string
}

type HttpConfig struct {
	Port string
}

type KafkaConfig struct {
	Brokers []string
	Topic   string
	GroupId string
}

const DefaultEnvFileName = ".env.example"
const DefaultHttpPort = "8085"

func LoadConfig() *Config {
	err := godotenv.Load(DefaultEnvFileName)
	if err != nil {
		log.Println("Error loading .env file, using default config")
	}
	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = DefaultHttpPort
	}
	return &Config{
		Db: DbConfig{
			Dsn: os.Getenv("DSN"),
		},
		Kafka: KafkaConfig{
			Brokers: strings.Split(os.Getenv("KAFKA_BROKERS"), ","),
			Topic:   os.Getenv("KAFKA_TOPIC"),
			GroupId: os.Getenv("KAFKA_GROUP_ID"),
		},
		Http: HttpConfig{
			Port: httpPort,
		},
	}
}

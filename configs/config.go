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
}

type DbConfig struct {
	Dsn string
}

type KafkaConfig struct {
	Brokers []string
	Topic   string
	GroupId string
}

const DefaultEnvFileName = ".env.example"

func LoadConfig() *Config {
	err := godotenv.Load(DefaultEnvFileName)
	if err != nil {
		log.Println("Error loading .env file, using default config")
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
	}
}

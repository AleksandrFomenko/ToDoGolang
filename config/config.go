package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port   string
	DBfile string
}

func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("не нашел файл .env")
	}

	return &Config{
		Port:   getEnv("TODO_PORT", "7540"),
		DBfile: getEnv("TODO_DBFILE", "scheduler.db"),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

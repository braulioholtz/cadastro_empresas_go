package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port            string
	MongoURI        string
	MongoDB         string
	MongoCollection string
	RabbitURL       string
}

func Load() Config {
	_ = godotenv.Load()
	cfg := Config{
		Port:            get("PORT", "8080"),
		MongoURI:        get("MONGODB_URI", "mongodb://mongo:27017"),
		MongoDB:         get("MONGODB_DB", "matriz"),
		MongoCollection: get("MONGODB_COLLECTION", "empresas"),
		RabbitURL:       get("RABBITMQ_URL", "amqp://guest:guest@rabbitmq:5672/"),
	}
	log.Printf("config loaded: port=%s db=%s", cfg.Port, cfg.MongoDB)
	return cfg
}

func get(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}

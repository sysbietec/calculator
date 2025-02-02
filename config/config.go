package config

import (
	"os"

	"github.com/joho/godotenv"
)

// Config estrutura para armazenar configurações do sistema
type Config struct {
	PostgresURL  string
	FirebirdURL  string
	SQLServerURL string
}

// Load carrega as variáveis de ambiente do arquivo .env
func Load() *Config {
	// Carrega o arquivo .env
	_ = godotenv.Load()

	postgresURL := os.Getenv("POSTGRES_URL")

	return &Config{
		PostgresURL:  postgresURL,
		FirebirdURL:  os.Getenv("FIREBIRD_URL"),
		SQLServerURL: os.Getenv("SQLSERVER_URL"),
	}
}

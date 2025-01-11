package main

import (
	"net/http"
	"os"

	"calculator/config"
	"calculator/internal/container"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
    "github.com/joho/godotenv"
)

func init() {
	// Configuração do logrus para JSON
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetOutput(os.Stdout) // Para visualizar no console
	logrus.SetLevel(logrus.InfoLevel)

	// Redirecionar logs para um arquivo
	file, err := os.OpenFile("logs.json", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		logrus.SetOutput(file)
	} else {
		logrus.Warn("Falha ao abrir o arquivo de logs, escrevendo para o console.")
	}
}

func main() {    
    
    if err := godotenv.Load("../.env"); err != nil {
        logrus.Warn("Não foi possível carregar o arquivo .env:", err)
    }

	// Inicializa as configurações do sistema
	config := config.Load()

	// Inicializa o container de dependências
	cont, err := container.NewContainer(config)
	if err != nil {
		logrus.Fatal("Erro ao inicializar o container de dependências:", err)
	}
	defer cont.Close()

	// Configura o roteamento
	r := mux.NewRouter()
	r.HandleFunc("/calcAlpha", cont.PriceController.CalculateAlphaHandler).Methods("GET")

	logrus.Info("Servidor na porta 8080...")
	if err := http.ListenAndServe(":8080", r); err != nil {
		logrus.Fatal("Erro ao iniciar o servidor:", err)
	}
}

package controllers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"calculator/domain/usecase"
)

// PriceController disponibiliza endpoints para os cálculos
type PriceController struct {
	priceUC usecase.PriceUseCase
}

// NewPriceController cria uma nova instância de PriceController
func NewPriceController(uc usecase.PriceUseCase) *PriceController {
	return &PriceController{priceUC: uc}
}

// /calcAlpha?sku=1234&userPrice=100.50
func (pc *PriceController) CalculateAlphaHandler(w http.ResponseWriter, r *http.Request) {
	// Obter o SKU da query string
	sku := r.URL.Query().Get("sku")
	if sku == "" {
		http.Error(w, "sku is required", http.StatusBadRequest)
		return
	}

	// Obter o preço personalizado (userPrice) da query string
	userPriceStr := r.URL.Query().Get("userPrice")
	var userPrice float64
	var err error

	// Se um preço personalizado for fornecido, convertê-lo para float
	if userPriceStr != "" {
		userPrice, err = strconv.ParseFloat(userPriceStr, 64)
		if err != nil {
			http.Error(w, "invalid userPrice value", http.StatusBadRequest)
			return
		}
	}

	// Chamar o caso de uso correto com o preço personalizado
	valorFinal, detailsJSON, err := pc.priceUC.CalculateAlphaPriceWithUserPrice(sku, userPrice)
	if err != nil {
		log.Println("Error calculating alpha:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Montar a resposta com o valor final e os detalhes
	resp := map[string]interface{}{
		"sku":         sku,
		"valor_final": valorFinal,
		"detalhes":    json.RawMessage(detailsJSON),
	}

	// Configurar o cabeçalho da resposta e enviar a resposta em JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

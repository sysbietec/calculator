package firebird

import (
	"database/sql"
	"fmt"

	"github.com/sirupsen/logrus"
)

type ProductService struct {
	db *sql.DB
}

// NewProductService criando nova instancia do serviço
func NewProductService(db *sql.DB) *ProductService {
	return &ProductService{db: db}
}

// CalculateIcmsEfetivoEDifal calcula o ICMS efetivo e o Difal baseado no perfil de imposto e origem do produto
func (ps *ProductService) CalculateIcmsEfetivoEDifal(produto int) (float64, float64, error) {
	// Consulta o valor total de RED_ICMS
	queryRedIcms := `
		SELECT SUM(RED_ICMS)
		FROM IMPOSTOS_PERFIL
		WHERE perfil_imposto = (
			SELECT perfil_imposto
			FROM produtos
			WHERE produto = ?
		)
	`

	var totalIcms sql.NullFloat64
	row := ps.db.QueryRow(queryRedIcms, produto)
	err := row.Scan(&totalIcms)
	if err != nil {
		return 0, 0, fmt.Errorf("erro ao consultar RED_ICMS: %w", err)
	}
	logrus.WithField("total_icms", totalIcms.Float64).Info("Valor total de ICMS calculado")

	// Consulta a origem do produto
	queryOrigemProd := `
		SELECT origem_prod
		FROM produtos
		WHERE produto = ?
	`

	var origemProd string
	row = ps.db.QueryRow(queryOrigemProd, produto)
	err = row.Scan(&origemProd)
	if err != nil {
		return 0, 0, fmt.Errorf("erro ao consultar origem_prod: %w", err)
	}
	logrus.WithField("origem_prod", origemProd).Info("Origem do produto consultada")

	// Determina se o produto é estrangeiro ou nacional
	origemUnimarcas := "NACIONAL"
	if origemProd == "1" || origemProd == "2" || origemProd == "3" || origemProd == "8" {
		origemUnimarcas = "ESTRANGEIRO"
	}
	logrus.WithField("origemUnimarcas", origemUnimarcas).Info("Origem Unimarcas determinada")

	// Aplica a lógica do cálculo de ICMS efetivo
	var icmsEfetivo float64
	if totalIcms.Valid && totalIcms.Float64 > 0 {
		icmsEfetivo = 0.088
	} else {
		icmsEfetivo = 0.18
	}
	logrus.WithField("icmsEfetivo", icmsEfetivo).Info("ICMS Efetivo calculado")

	// Aplica a lógica do cálculo de Difal
	var difal float64
	if origemUnimarcas == "NACIONAL" {
		if icmsEfetivo == 0.088 {
			difal = 0
		} else {
			difal = icmsEfetivo - 0.12
		}
	} else {
		difal = icmsEfetivo - 0.04
	}
	logrus.WithField("difal", difal).Info("Difal calculado")

	return icmsEfetivo, difal, nil
}

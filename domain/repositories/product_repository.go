package repositories

import (
	"calculator/domain/entities"
)

// ProductRepository define as operações para obter dados necessários
// ao cálculo do preço. Você pode expandir conforme precisar.
type ProductRepository interface {
	// Busca dados do 'productscmp' (Postgres)
	GetProductCmpValues(sku string) (entities.PriceInput, error)

	// Busca parâmetros padrão (Parameters) do Postgres
	GetParameters() (entities.Parameters, error)

	// Busca custo Firebird (departamento, comissao, frete) no Firebird
	GetCostFire(sku string) (entities.CostFire, error)
	
	// Se precisar, define também GetIcmsEfetivo() ou etc.
}

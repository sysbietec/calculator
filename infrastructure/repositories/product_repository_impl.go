package repositories

import (
	"database/sql"
	"fmt"
	// "strconv"
	"github.com/sirupsen/logrus"
	"calculator/domain/entities"
	"calculator/domain/repositories"
 )
 var log = logrus.New()

 // Configurando o logrus para saída em JSON
 func init() {
	 log.SetFormatter(&logrus.JSONFormatter{})
	 log.SetLevel(logrus.InfoLevel)
 }

 
// productRepositoryImpl implementa ProductRepository
type productRepositoryImpl struct {
	postgresDB  *sql.DB
	firebirdDB  *sql.DB
	sqlServerDB *sql.DB
}

// NewProductRepository constrói um repository com as três conexões
func NewProductRepository(pg *sql.DB, fb *sql.DB, ms *sql.DB) repositories.ProductRepository {
	return &productRepositoryImpl{
		postgresDB:  pg,
		firebirdDB:  fb,
		sqlServerDB: ms,
	}
}

// GetProductCmpValues → alimenta PriceInput (partial)
func (r *productRepositoryImpl) GetProductCmpValues(sku string) (entities.PriceInput, error) {
	q := `SELECT cmp_icms, cmp_pis_cofins, cmp, cmp_nf 
			FROM productscmp 
			WHERE produto = $1 
			AND index = (
				SELECT max_index 
				FROM productscmp 
				WHERE produto = $1 
				LIMIT 1
			)`
	row := r.postgresDB.QueryRow(q, sku)

	var pi entities.PriceInput
	err := row.Scan(&pi.IcmsMedio, &pi.PisCofinsMedio, &pi.CustoMedioLiq, &pi.CustoMedioNF)
	if err != nil {
		return pi, fmt.Errorf("GetProductCmpValues scan: %w", err)
	}
	return pi, nil
}

// GetParameters → carrega Parameters do Postgres
func (r *productRepositoryImpl) GetParameters() (entities.Parameters, error) {
	q := `SELECT lucro_adicional_desejado, lucro_padrao_desejado, imposto_federal, operacao,
	             custo_fixo, aliquota_pis, aliquota_cofins, rebate, fcp, redutor_padrao
	       FROM config_params 
	       WHERE id=1`
	row := r.postgresDB.QueryRow(q)

	var pm entities.Parameters
	err := row.Scan(
		&pm.LucroAdicionalDesejado,
		&pm.LucroPadraoDesejado,
		&pm.ImpostoFederal,
		&pm.Operacao,
		&pm.CustoFixo,
		&pm.AliquotaPis,
		&pm.AliquotaCofins,
		&pm.Rebate,
		&pm.Fcp,
		&pm.RedutorPadrao,
	)
	if err != nil {
		return pm, fmt.Errorf("GetParameters scan: %w", err)
	}
	return pm, nil
}

// GetCostFire → busca no Firebird (departamento, comissao, frete)
	func (r *productRepositoryImpl) GetCostFire(sku string) (entities.CostFire, error) {
		stringSku := "_0_0_U"
		product := sku + stringSku
		log.WithFields(logrus.Fields{
			"sku":sku,
			"product":product,
		}).Info("Gerando sku para consulta")

		// preciso inserir um log para identificar o que esta sendo montado em product
		q := `select n.sku,n.cod_produto, p.departamento, n.comissao, n.preco FROM 
		np_comissao_frete n join produtos p on p.cod_produto = n.cod_produto WHERE n.sku = ?`
		row := r.firebirdDB.QueryRow(q, product)

		var cf entities.CostFire
		err := row.Scan(&cf.Sku, &cf.Cod_produto, &cf.Departamento, &cf.Comissao, &cf.Frete)
		if err != nil {
			log.WithFields(logrus.Fields{
				"sku": product,
			}).Error("Erro ao escanear resultado:", err)
			return cf, fmt.Errorf("GetCostFire scan: %w", err)
		}
		log.WithFields(logrus.Fields{
			"departamento": cf.Departamento,
			"comissao":     cf.Comissao,
			"frete":        cf.Frete,
		}).Info("Dados recuperados com sucesso")

		return cf, nil
	}


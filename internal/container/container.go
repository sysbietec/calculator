package container

import (
	"database/sql"

	"calculator/config"
	"calculator/infrastructure/db"
	"calculator/infrastructure/repositories"
	"calculator/internal/firebird"
	"calculator/domain/usecase"
	"calculator/interface/controllers"
)

// Container estrutura para gerenciar dependências
type Container struct {
	PriceController *controllers.PriceController
	postgresDB      *sql.DB
	firebirdDB      *sql.DB
	sqlServerDB     *sql.DB
}

// NewContainer cria uma nova instância de Container
func NewContainer(cfg *config.Config) (*Container, error) {
	// Conexões com os bancos de dados
	postgresDB, err := db.NewPostgresConn(cfg.PostgresURL)
	if err != nil {
		return nil, err
	}

	firebirdDB, err := db.NewFirebirdConn(cfg.FirebirdURL)
	if err != nil {
		return nil, err
	}

	sqlServerDB, err := db.NewSQLServerConn(cfg.SQLServerURL)
	if err != nil {
		return nil, err
	}

	// Repositórios e serviços
	productRepo := repositories.NewProductRepository(postgresDB, firebirdDB, sqlServerDB)
	productService := firebird.NewProductService(firebirdDB)

	// UseCases e Controllers
	priceUC := usecase.NewPriceUseCase(productRepo, productService)
	priceCtrl := controllers.NewPriceController(priceUC)

	return &Container{
		PriceController: priceCtrl,
		postgresDB:      postgresDB,
		firebirdDB:      firebirdDB,
		sqlServerDB:     sqlServerDB,
	}, nil
}

// Close fecha as conexões de banco de dados
func (c *Container) Close() {
	if c.postgresDB != nil {
		c.postgresDB.Close()
	}
	if c.firebirdDB != nil {
		c.firebirdDB.Close()
	}
	if c.sqlServerDB != nil {
		c.sqlServerDB.Close()
	}
}

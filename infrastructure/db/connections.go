package db

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	_ "github.com/nakagami/firebirdsql"
	_ "github.com/denisenkom/go-mssqldb"
)

// NewPostgresConn cria uma nova conexão com o banco de dados Postgres
func NewPostgresConn(connStr string) (*sql.DB, error) {
    db, err := sql.Open("postgres", connStr)
    if err != nil {
        return nil, fmt.Errorf("erro ao conectar ao Postgres: %w", err)
    }
    if err := db.Ping(); err != nil {
        return nil, fmt.Errorf("erro ping Postgres: %w", err)
    }
    return db, nil
}

// NewFirebirdConn cria uma nova conexão com o banco de dados Firebird
func NewFirebirdConn(connStr string) (*sql.DB, error) {
	db, err := sql.Open("firebirdsql", connStr)
	if err != nil {
		return nil, fmt.Errorf("erro ao conectar ao Firebird: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("erro ping Firebird: %w", err)
	}
	return db, nil
}

// NewSQLServerConn cria uma nova conexão com o banco de dados SQL Server
func NewSQLServerConn(connStr string) (*sql.DB, error) {
	db, err := sql.Open("sqlserver", connStr)
	if err != nil {
		return nil, fmt.Errorf("erro ao conectar ao SQL Server: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("erro ping SQL Server: %w", err)
	}
	return db, nil
}

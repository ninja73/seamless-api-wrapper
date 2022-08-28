package postgres

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"seamless-api-wrapper/internal/config"
	"time"

	_ "github.com/lib/pq"
)

func InitDB(postgres *config.Postgres) (*sqlx.DB, error) {
	pqInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		postgres.Host, postgres.Port, postgres.User, postgres.Password, postgres.DB)

	db, err := sqlx.Connect("postgres", pqInfo)
	if err != nil {
		return nil, fmt.Errorf("error open db: %v", err)
	}

	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(postgres.PoolSize)
	db.SetConnMaxLifetime(15 * time.Second)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error ping db: %v", err)
	}

	return db, nil
}

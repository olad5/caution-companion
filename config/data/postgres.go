package data

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

func StartPostgres(DatabaseUrl string, l *zap.Logger) *sqlx.DB {
	connection, err := sqlx.Connect("postgres", DatabaseUrl)
	if err != nil {
		l.Fatal("Failed to create PostgresConnection: ", zap.Error(err))
	}
	return connection
}

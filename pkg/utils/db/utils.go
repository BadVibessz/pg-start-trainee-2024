package db

import (
	"database/sql"
	"errors"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"time"
)

func TryToConnectToDB(connectionURL, driver string, retires, interval int, logger *logrus.Logger) (*sqlx.DB, error) {
	var db *sqlx.DB
	var conn *sql.DB
	var err error

	for i := 0; i < retires; i++ {
		conn, err = sql.Open("pgx", connectionURL)
		if err != nil {
			return nil, errors.Join(ErrCannotOpenConnection, err)
		} else {
			db = sqlx.NewDb(conn, driver)

			if err = db.Ping(); err != nil {
				logger.Errorf("can't ping database: %v\nconnection string: %v", err, connectionURL)
				logger.Infof("retrying in %v sec...", interval)
				logger.Infof("retry %v of %v", i+1, retires)

				time.Sleep(time.Duration(interval) * time.Second)
			} else {
				return db, nil
			}
		}
	}

	return nil, err
}

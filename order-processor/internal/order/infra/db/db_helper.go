package db

import (
	"database/sql"
)

func InitialiazeDb(connString string) (*sql.DB, error) {
	if connString == "" {
		connString = ":memory:?_fk=on"
	}
	db, err := sql.Open("sqlite3", connString)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	_, err = db.Exec(
		`CREATE TABLE orders (
			id varchar(255) NOT NULL PRIMARY KEY ,
			price float NOT NULL,
			tax float NOT NULL,
			final_price float NOT NULL
			)`)
	if err != nil {
		return nil, err
	}
	return db, nil
}

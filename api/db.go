package main

import "database/sql"

type DB struct {
	*sql.DB
	ReadOnly *sql.DB
}

func NewDB(main string, readOnly string, maxConn int) (*DB, error) {
	dbMain, err := sql.Open("mysql", main)
	if err != nil {
		return nil, err
	}
	dbMain.SetMaxOpenConns(maxConn)
	if err := dbMain.Ping(); err != nil {
		return nil, err
	}

	db := &DB{DB: dbMain}

	if readOnly != "" {
		dbReadOnly, err := sql.Open("mysql", readOnly)
		if err != nil {
			return nil, err
		}
		dbReadOnly.SetMaxOpenConns(maxConn)
		if err := dbReadOnly.Ping(); err != nil {
			return nil, err
		}

		db.ReadOnly = dbReadOnly
	}

	if db.ReadOnly == nil {
		db.ReadOnly = dbMain
	}

	return db, nil
}

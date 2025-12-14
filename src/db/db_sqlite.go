package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

type LiteDB struct {
	connDb   *sql.DB
	dBPath   string
	debugSQL bool
}

func OpenSqliteDatabase(dbPath string, debugSql bool) (*LiteDB, error) {
	_, err := os.Stat(dbPath)
	if err != nil {
		return nil, err
	}
	ld := &LiteDB{
		dBPath:   dbPath,
		debugSQL: debugSql,
	}
	if err := ld.openSqliteDatabase(); err != nil {
		log.Println("[OpenSqliteDatabase] error")
		return nil, err
	}
	return ld, nil
}

func (ld *LiteDB) openSqliteDatabase() error {
	var err error
	dbname := ld.dBPath
	log.Println("Using the sqlite file: ", dbname)
	// Open database with foreign keys enabled
	ld.connDb, err = sql.Open("sqlite3", fmt.Sprintf("%s?_foreign_keys=1", dbname))
	if err != nil {
		return err
	}
	return nil
}

func (ld *LiteDB) GetTransaction() (*sql.Tx, error) {
	return ld.connDb.Begin()
}

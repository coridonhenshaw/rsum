package main

import (
	"bytes"
	"database/sql"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

type SqliteStruct struct {
	Filename string
	DB       *sql.DB
}

func DBConnect() (SqliteStruct, error) {

	var o SqliteStruct

	var SQL string
	var err error

	o.Filename = "rsum.sqlite3"

	if !WorkingDatabaseInCWD {
		Path, err := os.UserCacheDir()
		if err != nil {
			log.Panic(err)
		}
		o.Filename = filepath.Join(Path, o.Filename)
	}

	o.DB, err = sql.Open("sqlite3", o.Filename+"?cache=shared&_journal=WAL")
	if err != nil {
		log.Panic(err)
	}

	SQL = `PRAGMA SYNCHRONOUS = OFF`
	_, err = o.DB.Exec(SQL)
	if err != nil {
		log.Panic(err)
	}

	SQL = `PRAGMA journal_mode = WAL`
	_, err = o.DB.Exec(SQL)
	if err != nil {
		log.Panic(err)
	}

	_, err = o.DB.Exec("PRAGMA optimize")
	if err != nil {
		log.Panic(err)
	}

	return o, nil
}

func (o *SqliteStruct) Close() {
	if o.DB != nil {
		o.DB.Close()
	}
	if RemoveWorkingDatabase {
		o.Expunge()
	}
}

func (o *SqliteStruct) Expunge() {
	os.Remove(o.Filename)
}

//
//
//

type ResumeTableStruct struct {
	table           string
	AddStmt         *sql.Stmt
	GetStmt         *sql.Stmt
	DB              *SqliteStruct
	Mutex           sync.Mutex
	ResumeAvailable bool
}

func (o *SqliteStruct) CreateResumeTable(Identifier string) (*ResumeTableStruct, error) {
	var rc = new(ResumeTableStruct)

	var err error
	var SQL string

	rc.DB = o

	rc.table = "T" + Identifier + rc.table

	SQL = `CREATE TABLE IF NOT EXISTS ` + rc.table + ` (
		Filename TEXT,
		Size INTEGER,
		ModifiedTime INTEGER,
		Hash BINARY
		)`

	_, err = o.DB.Exec(SQL)
	if err != nil {
		log.Panic(err)
	}

	SQL = "INSERT INTO " + rc.table + ` (Filename, Size, ModifiedTime, Hash) VALUES (?, ?, ?, ?)`
	rc.AddStmt, err = o.DB.Prepare(SQL)
	if err != nil {
		log.Panic(err)
	}

	var Rows int
	SQL = "SELECT COUNT(Filename) FROM " + rc.table
	o.DB.QueryRow(SQL).Scan(&Rows)

	if Rows > 0 {
		rc.ResumeAvailable = true

		SQL = "CREATE INDEX IF NOT EXISTS Filename ON " + rc.table + " (Filename)"
		_, err = o.DB.Exec(SQL)
		if err != nil {
			log.Panic(err)
		}

		SQL = "SELECT Hash FROM " + rc.table + " WHERE Filename = ? AND Size = ? AND ModifiedTime = ?"
		rc.GetStmt, err = o.DB.Prepare(SQL)
		if err != nil {
			log.Panic(err)
		}
	}

	return rc, nil
}

func (o *ResumeTableStruct) GetTable() string { return o.table }

func (o *ResumeTableStruct) Close() {
	var SQL string
	var err error

	SQL = `DROP TABLE IF EXISTS ` + o.table
	_, err = o.DB.DB.Exec(SQL)
	if err != nil {
		log.Panic(err)
	}

	SQL = `VACUUM`
	_, err = o.DB.DB.Exec(SQL)
	if err != nil {
		log.Panic(err)
	}
}

// func (o *ResumeTableStruct) SetStem(Stem string) { o.stem = Stem }
// func (o *ResumeTableStruct) GetStem() string     { return o.stem }

//
//
//

type SQLTemplateStruct struct {
	//	Table   string
	Results string
	Left    string
	Right   string
	// Source  string
}

func (Data *SQLTemplateStruct) Render(Template string) string {

	var tpl bytes.Buffer

	t := template.Must(template.New("sqltemplate").Parse(Template))
	err := t.Execute(&tpl, Data)

	if err != nil {
		log.Panic(err)
	}

	return tpl.String()
}

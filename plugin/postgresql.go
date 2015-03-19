package plugin

import (
	"bytes"
	"database/sql"
	"time"

	"github.com/vektra/cypress"
)

const cEnableHstore = `
CREATE EXTENSION hstore
`

const cCreateTable = `
CREATE TABLE cypress_messages (
	timestamp TIMESTAMP,
	version INTEGER,
	type INTEGER,
	session_id TEXT,
	attributes HSTORE,
	tags HSTORE
)`

const cAddRow = `
INSERT INTO cypress_messages (
	timestamp,
	version,
	type,
	session_id,
	attributes,
	tags
) VALUES ($1, $2, $3, $4, $5, $6)`

type DBInterface interface {
	Ping() error
	Exec(query string, args ...interface{}) (sql.Result, error)
}

type PostgreSQL struct {
	Username string
	Password string
	Host     string
	Port     string
	DBName   string
	DB       DBInterface
}

func (p *PostgreSQL) dataSourceName() string {
	var buf bytes.Buffer
	buf.WriteString(p.Username)
	buf.WriteString(":")
	buf.WriteString(p.Password)
	buf.WriteString("@tcp(")
	buf.WriteString(p.Host)
	buf.WriteString(":")
	buf.WriteString(p.Port)
	buf.WriteString(")")
	buf.WriteString(p.DBName)
	return buf.String()
}

func (p *PostgreSQL) Init(db DBInterface) {
	p.DB = db
}

func (p *PostgreSQL) SetupDB() error {
	err := p.DB.Ping()
	if err != nil {
		return err
	}

	// TODO: first check if already enabled
	_, err = p.DB.Exec(cEnableHstore)
	if err != nil {
		// return err
	}

	// TODO: first check if already created
	_, err = p.DB.Exec(cCreateTable)
	if err != nil {
		// return err
	}

	// TODO: alter table if schema doesnt match

	return nil
}

func (p *PostgreSQL) Receive(m *cypress.Message) error {
	_, err := p.DB.Exec(cAddRow,
		m.GetTimestamp().Time().Format(time.RFC3339Nano),
		m.Version,
		m.Type,
		m.SessionId,
		m.HstoreAttributes(),
		m.HstoreTags(),
	)
	if err != nil {
		return err
	}

	return nil
}

package goose

import (
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/lib/pq"
)

// DBDriver encapsulates the info needed to work with
// a specific database driver
type DBDriver struct {
	Name    string
	OpenStr string
	Import  string
	Dialect SqlDialect
}

type DBConf struct {
	MigrationsDir string
	Env           string
	Driver        DBDriver
	PgSchema      string
}

// extract configuration details from the given file
func NewDBConf(p, env string, pgschema string, drv, open string) (*DBConf, error) {

	// Automatically parse postgres urls
	if drv == "postgres" {

		// Assumption: If we can parse the URL, we should
		if parsedURL, err := pq.ParseURL(open); err == nil && parsedURL != "" {
			open = parsedURL
		}
	}

	d := newDBDriver(drv, open)

	if !d.IsValid() {
		return nil, errors.New(fmt.Sprintf("Invalid DBConf: %v", d))
	}

	return &DBConf{
		MigrationsDir: filepath.Join(p, "cases"),
		Env:           env,
		Driver:        d,
		PgSchema:      pgschema,
	}, nil
}

// Create a new DBDriver and populate driver specific
// fields for drivers that we know about.
// Further customization may be done in NewDBConf
func newDBDriver(name, open string) DBDriver {

	d := DBDriver{
		Name:    name,
		OpenStr: open,
	}

	switch name {
	case "postgres":
		d.Import = "github.com/lib/pq"
		d.Dialect = &PostgresDialect{}

	case "mymysql":
		d.Import = "github.com/ziutek/mymysql/godrv"
		d.Dialect = &MySqlDialect{}

	case "mysql":
		d.Import = "github.com/go-sql-driver/mysql"
		d.Dialect = &MySqlDialect{}

	case "sqlite3":
		d.Import = "github.com/mattn/go-sqlite3"
		d.Dialect = &Sqlite3Dialect{}
	}

	return d
}

// ensure we have enough info about this driver
func (drv *DBDriver) IsValid() bool {
	return len(drv.Import) > 0 && drv.Dialect != nil
}

// OpenDBFromDBConf wraps database/sql.DB.Open() and configures
// the newly opened DB based on the given DBConf.
//
// Callers must Close() the returned DB.
func OpenDBFromDBConf(conf *DBConf) (*sql.DB, error) {

	db, err := sql.Open(conf.Driver.Name, conf.Driver.OpenStr)
	if err != nil {
		return nil, err
	}

	// if a postgres schema has been specified, apply it
	if conf.Driver.Name == "postgres" && conf.PgSchema != "" {
		if _, err := db.Exec("SET search_path TO " + conf.PgSchema); err != nil {
			return nil, err
		}
	}

	return db, nil
}

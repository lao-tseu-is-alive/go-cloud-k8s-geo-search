package database

import (
	"errors"
	"fmt"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/golog"
)

var (
	ErrNoRecordFound     = errors.New("record not found")
	ErrCouldNotBeCreated = errors.New("could not be created in DB")
)

// DB is the interface for a simple table store.
type DB interface {
	ExecActionQuery(sql string, arguments ...interface{}) (rowsAffected int, err error)
	GetQueryInt(sql string, arguments ...interface{}) (result int, err error)
	GetQueryBool(sql string, arguments ...interface{}) (result bool, err error)
	GetQueryString(sql string, arguments ...interface{}) (result string, err error)
	GetVersion() (result string, err error)
	GetSpatialVersion() (result string, err error)
	DoesTableExist(schema, table string) (exist bool)
	Close()
	IsItSpatial() bool
	GetQueryStringArr(sql string, arguments ...interface{}) (result []string, err error)
}

func GetErrorF(errMsg string, err error) error {
	return errors.New(fmt.Sprintf("%s [%v]", errMsg, err))
}

// GetInstance with appropriate driver
func GetInstance(dbDriver, dbConnectionString string, maxConnectionCount int, log golog.MyLogger) (DB, error) {
	var err error
	var db DB

	if dbDriver == "pgx" {
		db, err = newPgxConn(dbConnectionString, maxConnectionCount, log)
		if err != nil {
			return nil, fmt.Errorf("error opening postgresql database with pgx driver: %s", err)
		}
	} else if dbDriver == "sqlite3" {
		db, err = NewSqlite3DB(dbConnectionString, log)
		if err != nil {
			return nil, fmt.Errorf("error opening sqlite3 database with sqlite3 driver: %s", err)
		}
	} else {
		return nil, errors.New("unsupported DB driver type")
	}

	return db, nil
}

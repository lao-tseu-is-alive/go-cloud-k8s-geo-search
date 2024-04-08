package database

import (
	"database/sql"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/golog"
	"github.com/mattn/go-sqlite3"
	"sync"
)

const getSqliteVersion = "SELECT sqlite_version();"
const getSpatialiteVersion = "SELECT spatialite_version();"
const getSqliteTableExists = "SELECT name FROM sqlite_master WHERE type='table';"

// SQLITE3 is a struct to hold the connection to a sqlite3 database
type SQLITE3 struct {
	Conn *sql.DB
	lck  sync.RWMutex // https://godoc.org/github.com/mxk/go-sqlite/sqlite3#hdr-Concurrency
	log  golog.MyLogger
}

func NewSqlite3DB(geopackageFilePath string, log golog.MyLogger) (DB, error) {
	var successOrFailure = "OK"
	sql.Register("sqlite3_with_spatialite", &sqlite3.SQLiteDriver{
		Extensions: []string{"mod_spatialite"},
	})

	// Load the geopackage

	log.Info("--------------->>NewSqlite3DB--------------")
	db, err := sql.Open("sqlite3_with_spatialite", geopackageFilePath)
	if err != nil {
		successOrFailure = "FAILED"
		log.Info("Connecting to Sqlite3 database '%s' : %s ", geopackageFilePath, successOrFailure)
		log.Fatal("💥 ERROR TRYING DB CONNECTION : %v ", err)
	} else {
		log.Info("Connecting to Sqlite3 database '%s' : %s", geopackageFilePath, successOrFailure)
		log.Info("Fetching one record to test if db connection is valid...")
		var version string
		getSqlite3Version := "SELECT sqlite_version();"
		if errPing := db.QueryRow(getSqlite3Version).Scan(&version); errPing != nil {
			log.Error("Connection is invalid ! ")
			log.Fatal("DB ERROR scanning row: %s", errPing)
		}
		log.Info("SUCCESS Connecting to Sqlite3 version : [%s]", version)
	}
	log.Info("---------------NewSqlite3DB>>--------------")

	return &SQLITE3{
		Conn: db,
		lck:  sync.RWMutex{},
	}, err
}

func (db *SQLITE3) Close() {
	err := db.Conn.Close()
	if err != nil {
		db.log.Error("problem doing db.Conn.Close(): %v", err)
	}
	return
}

func (db *SQLITE3) ExecActionQuery(sql string, arguments ...interface{}) (rowsAffected int, err error) {
	db.lck.Lock()
	defer db.lck.Unlock()
	res, err := db.Conn.Exec(sql, arguments...)
	if err != nil {
		db.log.Error("Exec unexpectedly failed with %v: %v", sql, err)
		return 0, err
	}
	rowsAff, err := res.RowsAffected()
	if err != nil {
		db.log.Error("RowsAffected unexpectedly failed with %v: %v", sql, err)
		return 0, err
	}
	// golog.Info("Rows Affected : %v ", rowsAff)
	return int(rowsAff), err
}

func (db *SQLITE3) GetQueryInt(sql string, arguments ...interface{}) (result int, err error) {
	db.lck.RLock()
	defer db.lck.RUnlock()
	err = db.Conn.QueryRow(sql, arguments...).Scan(&result)
	if err != nil {
		db.log.Error("GetQueryInt(%s) queryRow unexpectedly failed. args : (%v), error : %v\n", sql, arguments, err)
		return 0, err
	}
	return result, err
}

func (db *SQLITE3) GetQueryBool(sql string, arguments ...interface{}) (result bool, err error) {
	db.lck.RLock()
	defer db.lck.RUnlock()
	err = db.Conn.QueryRow(sql, arguments...).Scan(&result)
	if err != nil {
		db.log.Error("GetQueryBool(%s) queryRow unexpectedly failed. args : (%v), error : %v\n", sql, arguments, err)
		return false, err
	}
	return result, err
}

func (db *SQLITE3) GetQueryString(sql string, arguments ...interface{}) (result string, err error) {
	db.lck.RLock()
	defer db.lck.RUnlock()
	err = db.Conn.QueryRow(sql, arguments...).Scan(&result)
	if err != nil {
		db.log.Error("GetQueryString(%s) queryRow unexpectedly failed. args : (%v), error : %v\n", sql, arguments, err)
		return "", err
	}
	return result, err
}

func (db *SQLITE3) GetQueryStringArr(sql string, arguments ...interface{}) (result []string, err error) {
	db.lck.RLock()
	defer db.lck.RUnlock()
	rows, err := db.Conn.Query(sql, arguments...)
	if err != nil {
		db.log.Error(" GetQueryString(%s) query unexpectedly failed. args : (%v), error : %v\n", sql, arguments, err)
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var val string
		err = rows.Scan(&val)
		if err != nil {
			db.log.Error(" GetQueryString(%s) rows.Scan unexpectedly failed. args : (%v), error : %v\n", sql, arguments, err)
			return nil, err
		}
		result = append(result, val)

	}
	return result, nil
}

func (db *SQLITE3) IsItSpatial() bool {
	db.lck.RLock()
	defer db.lck.RUnlock()
	// check if the geopackage is valid
	number, err := db.GetQueryInt("SELECT count(*) as number FROM sqlite_master WHERE type='table' AND name = 'gpkg_geometry_columns';")
	if err != nil {
		db.log.Error("IsItSpatial() query unexpectedly failed. error : %v\n", err)
		return false
	}
	return number > 0
}

func (db *SQLITE3) GetSpatialVersion() (result string, err error) {
	spatialiteVersion, err := db.GetQueryString(getSpatialiteVersion)
	if err != nil {
		db.log.Fatal("💥 ERROR: 'calling GetQueryString(SELECT spatialite_version())': %v", err)
		return "", err
	}
	return spatialiteVersion, err
}

func (db *SQLITE3) GetVersion() (result string, err error) {
	sqliteVersion, err := db.GetQueryString(getSqliteVersion)
	if err != nil {
		db.log.Fatal("💥 ERROR: 'calling GetQueryString(%s)': %v", getSqliteVersion, err)
		return "", err
	}
	return sqliteVersion, err
}

func (db *SQLITE3) DoesTableExist(schema, table string) (exist bool) {
	db.lck.RLock()
	defer db.lck.RUnlock()
	// check if the geopackage is valid
	number, err := db.GetQueryInt("SELECT count(*) as number FROM sqlite_master WHERE type='table' AND name = '%s';", table)
	if err != nil {
		db.log.Error("DoesTableExist() query unexpectedly failed. error : %v\n", err)
		return false
	}
	return number > 0
}

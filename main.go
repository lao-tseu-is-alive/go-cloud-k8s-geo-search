package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-geo-search/golog"
	"github.com/mattn/go-sqlite3"
	"html/template"
	"log"
	"os"
	"strconv"
	"sync"
)

// a simple http server using go Mux from net/http package

const (
	APP                 = "go-cloud-geo-search"
	VERSION             = "0.0.1"
	defaultPort         = 8099
	charsetUTF8         = "charset=UTF-8"
	MIMEHtml            = "text/html"
	MIMEHtmlCharsetUTF8 = MIMEHtml + "; " + charsetUTF8
	geopackageFilePath  = "geodata/swissBOUNDARIES3D_1_5_LV95_LN02.gpkg"
)

// GetPortFromEnv returns a valid TCP/IP listening ':PORT' string based on the values of environment variable :
//
//		PORT : int value between 1 and 65535 (the parameter defaultPort will be used if env is not defined)
//	 in case the ENV variable PORT exists and contains an invalid integer the functions returns an empty string and an error
func GetPortFromEnv(defaultPort int) (string, error) {
	port := defaultPort
	if portEnv := os.Getenv("PORT"); portEnv != "" {
		var err error
		if port, err = strconv.Atoi(portEnv); err != nil {
			return "", fmt.Errorf("invalid value for PORT env variable : %s", err)
		}
	}
	if port < 1 || port > 65535 {
		return "", fmt.Errorf("invalid value for PORT env variable : %d", port)
	}
	return fmt.Sprintf(":%d", port), nil
}

// getHelloMsg returns a string with a personalised greeting
func getHelloMsg(username string) (string, error) {
	const helloMsg = `
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="utf-8">
	<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/skeleton/2.0.4/skeleton.min.css" integrity="sha512-EZLkOqwILORob+p0BXZc+Vm3RgJBOe1Iq/0fiI7r/wJgzOFZMlsqTa29UEl6v6U6gsV4uIpsNZoV32YZqrCRCQ==" crossorigin="anonymous" referrerpolicy="no-referrer" />
    <title>Hello {{.UserName}}</title>
  </head>
  <body>
	<h3>Hello, {{.UserName}}!</h3>
  </body>
</html>
`
	data := struct {
		UserName string
	}{UserName: username}
	tpl := new(bytes.Buffer)
	t, err := template.New("hello-page").Parse(helloMsg)
	if err != nil {
		return "", err
	}
	if err := t.Execute(tpl, data); err != nil {
		return "", err
	}
	return tpl.String(), nil
}

// SQLITE3 is a struct to hold the connection to a sqlite3 database
type SQLITE3 struct {
	Conn *sql.DB
	lck  sync.RWMutex // https://godoc.org/github.com/mxk/go-sqlite/sqlite3#hdr-Concurrency
	log  golog.MyLogger
}

func NewSqlite3DB(geopackageFilePath string, log golog.MyLogger) (SQLITE3, error) {
	var successOrFailure = "OK"
	sql.Register("sqlite3_with_spatialite", &sqlite3.SQLiteDriver{
		Extensions: []string{"mod_spatialite"},
	})

	// Load the geopackage

	fmt.Println("--------------------------------------------------------------------------------------------")
	db, err := sql.Open("sqlite3_with_spatialite", geopackageFilePath)
	if err != nil {
		successOrFailure = "FAILED"
		log.Info("Connecting to Sqlite3 database %s :%s \n", geopackageFilePath, successOrFailure)
		log.Fatal("💥 ERROR TRYING DB CONNECTION : %v ", err)
	} else {
		log.Info("Connecting to database %s : %s \n", geopackageFilePath, successOrFailure)
		log.Info("Fetching one record to test if db connection is valid...\n")
		var version string
		getSqlite3Version := "SELECT sqlite_version();"
		if errPing := db.QueryRow(getSqlite3Version).Scan(&version); errPing != nil {
			log.Error("Connection is invalid ! ")
			log.Fatal("DB ERROR scanning row: %s", errPing)
		}
		log.Info("SUCCESS Connecting to Sqlite3 version : [%s]", version)
	}

	fmt.Println("--------------------------------------------------------------------------------------------")

	return SQLITE3{
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

func (db *SQLITE3) GetQueryString(sql string, arguments ...interface{}) (result string, err error) {
	db.lck.RLock()
	defer db.lck.RUnlock()
	err = db.Conn.QueryRow(sql, arguments...).Scan(&result)
	if err != nil {
		db.log.Error(" GetQueryString(%s) queryRow unexpectedly failed. args : (%v), error : %v\n", sql, arguments, err)
		return "", err
	}
	return result, err
}

func main() {
	prefix := fmt.Sprintf("%s ", APP)
	l, err := golog.NewLogger("zap", golog.DebugLevel, prefix)
	if err != nil {
		log.Fatalf("💥 ERROR: 'calling NewLogger()': %v", err)
	}
	l.Info("Starting %s version :%s\n", APP, VERSION)
	db, err := NewSqlite3DB(geopackageFilePath, l)
	defer db.Close()
	log.Printf("INFO: 'calling sql.Open()': geopackage %s loaded", geopackageFilePath)
	//q := "SELECT InitSpatialMetaData();"
	//db.ExecActionQuery(q)
	spatialiteVersion, err := db.GetQueryString("SELECT spatialite_version();")
	if err != nil {
		l.Fatal("💥 ERROR: 'calling GetQueryString()': %v", err)
	}
	l.Info("Spatialite version : %s", spatialiteVersion)
	//db.GetQueryString("SELECT * FROM sqlite_master WHERE type='table';")
	os.Exit(0)
	/*

		listenAddr, err := GetPortFromEnv(defaultPort)
			if err != nil {
				log.Fatalf("💥 ERROR: 'calling GetPortFromEnv()': %v", err)
			}

			mux := http.NewServeMux()
		mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
			defaultResponse := fmt.Sprintf("Hello from %s v%s", APP, VERSION)

			w.Header().Set("Content-Type", MIMEHtmlCharsetUTF8)
			n, err := fmt.Fprintf(w, defaultResponse)
			if err != nil {
				log.Printf("ERROR: 'calling fmt.Fprintf()': %v", err)
			}
			log.Printf("INFO: 'calling fmt.Fprintf()': %d bytes written", n)
		})
		mux.HandleFunc("GET /hello/{name}", func(w http.ResponseWriter, r *http.Request) {
			name := r.PathValue("name")
			helloResponse, err := getHelloMsg(name)
			if err != nil {
				log.Printf("ERROR: 'calling getHelloMsg()': %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", MIMEHtmlCharsetUTF8)
			n, err := fmt.Fprintf(w, helloResponse)
			if err != nil {
				log.Printf("ERROR: 'calling fmt.Fprintf()': %v", err)
			}
			log.Printf("INFO: 'calling fmt.Fprintf()': %d bytes written", n)

		})

		log.Printf("Starting %s v%s on %s", APP, VERSION, listenAddr)
		log.Fatal(http.ListenAndServe(listenAddr, mux))
	*/
}

package main

import (
	"bytes"
	"fmt"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/golog"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-geo-search/pkg/database"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-geo-search/pkg/version"
	"html/template"
	"log"
	"os"
	"runtime"
)

// a simple http server using go Mux from net/http package

const (
	defaultPort           = 8099
	charsetUTF8           = "charset=UTF-8"
	MIMEHtml              = "text/html"
	MIMEHtmlCharsetUTF8   = MIMEHtml + "; " + charsetUTF8
	geopackageFilePath    = "geodata/swissBOUNDARIES3D_1_5_LV95_LN02.gpkg"
	sqlListTables         = "SELECT name FROM sqlite_master WHERE type='table' AND name = 'gpkg_geometry_columns';"
	sqlListGeometryTables = "SELECT table_name FROM gpkg_geometry_columns;"
)

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

func main() {
	prefix := fmt.Sprintf("%s ", version.APP)
	l, err := golog.NewLogger("zap", golog.DebugLevel, prefix)
	if err != nil {
		log.Fatalf("💥 ERROR: 'calling NewLogger()': %v", err)
	}
	l.Info("Starting %s version :%s", version.APP, version.VERSION)

	db, err := database.GetInstance("sqlite3", geopackageFilePath, runtime.NumCPU(), l)
	if err != nil {
		l.Fatal("💥💥 error doing database.GetInstance(sqlite ...) error: %v", err)
	}
	defer db.Close()

	l.Info("SUCCESS calling GetInstance(\"sqlite3\",%s)", geopackageFilePath)
	//db.ExecActionQuery("SELECT InitSpatialMetaData();")
	isItGeoPackage := db.IsItSpatial()
	if !isItGeoPackage {
		l.Fatal("💥 ERROR: 'IsItSpatial()': %v", err)
	}
	spatialiteVersion, err := db.GetSpatialVersion()
	if err != nil {
		l.Fatal("💥 ERROR: 'calling GetSpatialVersion()': %v", err)
	}
	l.Info("Spatialite version : %s", spatialiteVersion)

	geosVersion, err := db.GetQueryString("SELECT geos_version();")
	if err != nil {
		l.Fatal("💥 ERROR: 'calling GetQueryString(SELECT geos_version)': %v", err)
	}
	l.Info("geosVersion version : %s", geosVersion)

	hasGeoPackageExtension, err := db.GetQueryBool("SELECT HasGeoPackage();")
	if err != nil {
		l.Fatal("💥 ERROR: 'calling GetQueryBool(SELECT HasGeoPackage())': %v", err)
	}
	l.Info("Is GeoPackage extension present : %v", hasGeoPackageExtension)

	tables, err := db.GetQueryStringArr(sqlListGeometryTables)
	l.Info("Listing Geometry tables :")
	if err != nil {
		l.Fatal("💥 ERROR: 'calling GetQueryStringArr(%s)': %v", sqlListGeometryTables, err)
	}
	for _, table := range tables {
		l.Info("Table : %s", table)
	}

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

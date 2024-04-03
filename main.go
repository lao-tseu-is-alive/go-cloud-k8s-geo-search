package main

import (
	"bytes"
	"fmt"
	"github.com/lukeroth/gdal"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
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

// readGeopackage reads a geopackage file and returns a string with the content
func readGeopackage(geopackageFilePath string) ([]string, error) {
	ds, err := gdal.Open(geopackageFilePath, gdal.ReadOnly)
	if err != nil {
		return make([]string, 0), fmt.Errorf("could not open geopackage file: %s", geopackageFilePath)
	}
	defer ds.Close()

	for dsFile := range ds.FileList() {
		fmt.Println(dsFile)
	}

	return ds.FileList(), nil
}

func main() {

	listenAddr, err := GetPortFromEnv(defaultPort)
	if err != nil {
		log.Fatalf("💥 ERROR: 'calling GetPortFromEnv()': %v", err)
	}
	numOGRDriver := gdal.OGRDriverCount()
	log.Printf("INFO: 'calling gdal.OGRDriverCount()': %d", numOGRDriver)
	log.Printf("INFO: 'calling gdal/ogr version ': %d.%d", gdal.VERSION_MAJOR, gdal.VERSION_MINOR)
	ogrDriver := gdal.OGRDriverByName("GPKG")
	if ogrDriver == nil {
		log.Fatalf("💥 ERROR: 'calling gdal.OGRDriverByName()': %v", err)
	}
	ds, ok := ogrDriver.Open(geopackageFilePath, 0)
	if !ok {
		log.Fatalf("💥 ERROR: 'calling ogrDriver.Open()': %v", err)
	}
	fmt.Printf("INFO: 'succes calling ogrDriver.Open(), found %d Layers'", ds.LayerCount())
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

}

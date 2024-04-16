package main

import (
	"embed"
	"fmt"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/config"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/golog"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-geo-search/pkg/go-http-server"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-geo-search/pkg/version"
	"log"
)

const (
	defaultPort = 8099
)

// content holds our static web server content.
//
//go:embed goCloudGeoSearchFront/dist/*
var content embed.FS

func main() {

	prefix := fmt.Sprintf("%s ", version.APP)
	l, err := golog.NewLogger("zap", golog.DebugLevel, prefix)
	if err != nil {
		log.Fatalf("💥💥 error log.NewLogger error: %v'\n", err)
	}

	listenAddr, err := config.GetPortFromEnv(defaultPort)
	if err != nil {
		l.Fatal("💥💥 error doing config.GetPortFromEnv got error: %v'\n", err)
	}
	l.Info("'Will start HTTP server listening on port %s'", listenAddr)
	server := go_http_server.NewHttpServer(listenAddr, l)
	err = server.StartServer()
	if err != nil {
		l.Fatal("💥💥 error doing server.StartServer() got error: %v'\n", err)
	}
}

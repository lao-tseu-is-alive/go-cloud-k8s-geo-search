package go_http_server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/golog"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	defaultProtocol        = "http"
	secondsShutDownTimeout = 5 * time.Second  // maximum number of second to wait before closing server
	defaultReadTimeout     = 10 * time.Second // max time to read request from the client
	defaultWriteTimeout    = 10 * time.Second // max time to write response to the client
	defaultIdleTimeout     = 2 * time.Minute  // max time for connections using TCP Keep-Alive
	initCallMsg            = "INITIAL CALL TO %s()"
	formatTraceRequest     = "TRACE: [%s] %s  path:'%s', RemoteAddrIP: [%s], msg: %s, val: %v"
	formatErrRequest       = "ERROR: Http method not allowed [%s] %s  path:'%s', RemoteAddrIP: [%s]\n"
	charsetUTF8            = "charset=UTF-8"
	MIMEAppJSON            = "application/json"
	MIMEAppJSONCharsetUTF8 = MIMEAppJSON + "; " + charsetUTF8
	HeaderContentType      = "Content-Type"
)

// HttpServer is a simple http server struct to store the server configuration
type HttpServer struct {
	listenAddr string
	logger     golog.MyLogger
	srvMux     *http.ServeMux
	startTime  time.Time
	httpServer *http.Server
}

// NewHttpServer creates a new HttpServer instance
func NewHttpServer(listenAddr string, l golog.MyLogger) *HttpServer {
	var defaultHttpLogger *log.Logger
	defaultHttpLogger, err := l.GetDefaultLogger()
	if err != nil {
		// in case we cannot get a valid logger.Logger for http let's create a reasonable one
		defaultHttpLogger = log.New(os.Stderr, "NewHttpServer::defaultHttpLogger", log.Ldate|log.Ltime|log.Lshortfile)
	}
	srvMux := http.NewServeMux()
	return &HttpServer{
		listenAddr: listenAddr,
		logger:     l,
		srvMux:     srvMux,
		startTime:  time.Date(1964, time.December, 21, 0, 0, 0, 0, time.UTC),
		httpServer: &http.Server{
			Addr:         listenAddr,
			Handler:      srvMux,
			ErrorLog:     defaultHttpLogger,
			ReadTimeout:  defaultReadTimeout,  // max time to read request from the client
			WriteTimeout: defaultWriteTimeout, // max time to write response to the client
			IdleTimeout:  defaultIdleTimeout,  // max time for connections using TCP Keep-Alive
		},
	}
}

// waitForShutdownToExit will wait for interrupt signal SIGINT or SIGTERM and gracefully shutdown the server after secondsToWait seconds.
func waitForShutdownToExit(srv *http.Server, secondsToWait time.Duration) {
	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// Block until a signal is received.
	// wait for SIGINT (interrupt) 	: ctrl + C keypress, or in a shell : kill -SIGINT processId
	sig := <-interruptChan
	srv.ErrorLog.Printf("INFO: 'SIGINT %d interrupt signal received, about to shut down server after max %v seconds...'\n", sig, secondsToWait.Seconds())

	// create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), secondsToWait)
	defer cancel()
	// gracefully shuts down the server without interrupting any active connections
	// as long as the actives connections last less than shutDownTimeout
	// https://pkg.go.dev/net/http#Server.Shutdown
	if err := srv.Shutdown(ctx); err != nil {
		srv.ErrorLog.Printf("💥💥 ERROR: 'Problem doing Shutdown %v'\n", err)
	}
	<-ctx.Done()
	srv.ErrorLog.Println("INFO: 'Server gracefully stopped, will exit'")
	os.Exit(0)
}

func (s *HttpServer) jsonResponse(w http.ResponseWriter, result interface{}) {
	body, err := json.Marshal(result)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		s.logger.Error("'JSON marshal failed. Error: %v'", err)
		return
	}
	var prettyOutput bytes.Buffer
	err = json.Indent(&prettyOutput, body, "", "  ")
	if err != nil {
		s.logger.Error("'JSON Indent failed. Error: %v'", err)
		return
	}
	w.Header().Set(HeaderContentType, MIMEAppJSONCharsetUTF8)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(prettyOutput.Bytes())
	if err != nil {
		s.logger.Error("'w.Write failed. Error: %v'", err)
		return
	}
}

// StartServer initializes all the handlers paths of this web server, it is called inside the NewGoHttpServer constructor
func (s *HttpServer) StartServer() error {

	// Starting the web server in his own goroutine
	go func() {
		s.logger.Info("starting http server listening at %s://localhost%s/", defaultProtocol, s.listenAddr)
		s.startTime = time.Now()
		err := s.httpServer.ListenAndServe()
		if err != nil && !errors.Is(http.ErrServerClosed, err) {
			s.logger.Fatal("💥💥 error could not listen on tcp port %q. error: %s", s.listenAddr, err)
		}
	}()
	s.logger.Debug("Server listening on : %s PID:[%d]", s.httpServer.Addr, os.Getpid())

	// Graceful Shutdown on SIGINT (interrupt)
	waitForShutdownToExit(s.httpServer, secondsShutDownTimeout)
	return nil
}

package go_http_server

import (
	"errors"
	"fmt"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-geo-search/pkg/info"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-geo-search/pkg/version"
	"github.com/rs/xid"
	"io/fs"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	httpErrMethodNotAllow = "ERROR: Http method not allowed"
	htmlHeaderStart       = `<!DOCTYPE html><html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/skeleton/2.0.4/skeleton.min.css"/>`
	defaultNotFound       = "🤔 ℍ𝕞𝕞... 𝕤𝕠𝕣𝕣𝕪 :【𝟜𝟘𝟜 : ℙ𝕒𝕘𝕖 ℕ𝕠𝕥 𝔽𝕠𝕦𝕟𝕕】🕳️ 🔥"
)

func getHtmlHeader(title string) string {
	return fmt.Sprintf("%s<title>%s</title></head>", htmlHeaderStart, title)
}

func getHtmlMsg(title, msg string) string {
	return getHtmlHeader(title) +
		fmt.Sprintf("\n<body><div class=\"container\"><h3>%s</h3></div></body></html>", msg)
}

func (s *HttpServer) getReadinessHandler() http.HandlerFunc {
	handlerName := "getReadinessHandler"
	s.logger.Info(initCallMsg, handlerName)
	return func(w http.ResponseWriter, r *http.Request) {
		s.logger.Info(formatTraceRequest, handlerName, r.Method, r.URL.Path, r.RemoteAddr)
		if r.Method == http.MethodGet {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
}
func (s *HttpServer) getHealthHandler() http.HandlerFunc {
	handlerName := "getHealthHandler"
	s.logger.Info(initCallMsg, handlerName)
	return func(w http.ResponseWriter, r *http.Request) {
		s.logger.Info(formatTraceRequest, handlerName, r.Method, r.URL.Path, r.RemoteAddr)
		if r.Method == http.MethodGet {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
}
func (s *HttpServer) getInfoHandler() http.HandlerFunc {
	handlerName := "getInfoHandler"

	s.logger.Info(initCallMsg, handlerName)
	hostName, err := os.Hostname()
	if err != nil {
		s.logger.Error("💥💥 'os.Hostname() returned an error : %v'", err)
		hostName = "#unknown#"
	}

	osReleaseInfo, errConf := info.GetOsInfo()

	if errConf.Err != nil {
		var pathError *fs.PathError
		switch {
		case errors.As(errConf.Err, &pathError):
			s.logger.Warn("NOTICE: 'GetOsInfo() dif not find os-release : %v'", errConf.Err)
		default:
			s.logger.Error("💥💥 'GetOsInfo() returned an error : %+#v'", errConf.Err)
		}
	}
	// fmt.Printf("%+v\n", osReleaseInfo)

	uptimeOS, err := info.GetOsUptime()
	if err != nil {
		s.logger.Error("💥💥 'GetOsUptime() returned an error : %+#v'", err)
	}
	k8sVersion := ""
	k8sCurrentNameSpace := ""
	k8sUrl, err := info.GetKubernetesApiUrlFromEnv()
	if err != nil {
		s.logger.Error("💥💥 'GetKubernetesApiUrlFromEnv() returned an error : %+#v'", err)
	} else {
		// here we can assume that we are inside a k8s container...
		info, errConnInfo := info.GetKubernetesConnInfo(s.logger, defaultReadTimeout)
		if errConnInfo.Err != nil {
			s.logger.Error("💥💥 'GetKubernetesConnInfo() returned an error : %s : %+#v'", errConnInfo.Msg, errConnInfo.Err)
		}
		k8sVersion = info.Version
		k8sCurrentNameSpace = info.CurrentNamespace
	}

	latestK8sVersion, err := info.GetKubernetesLatestVersion(s.logger)
	if err != nil {
		s.logger.Error("💥💥 'GetKubernetesLatestVersion() returned an error : %+#v'", err)
	}

	data := info.RuntimeInfo{
		Hostname:            hostName,
		Pid:                 os.Getpid(),
		PPid:                os.Getppid(),
		Uid:                 os.Getuid(),
		AppName:             version.APP,
		Version:             version.VERSION,
		ParamName:           "_NO_PARAMETER_NAME_",
		RemoteAddr:          "",
		RequestId:           "",
		GOOS:                runtime.GOOS,
		GOARCH:              runtime.GOARCH,
		Runtime:             runtime.Version(),
		NumGoroutine:        strconv.FormatInt(int64(runtime.NumGoroutine()), 10),
		OsReleaseName:       osReleaseInfo.Name,
		OsReleaseVersion:    osReleaseInfo.Version,
		OsReleaseVersionId:  osReleaseInfo.VersionId,
		NumCPU:              strconv.FormatInt(int64(runtime.NumCPU()), 10),
		Uptime:              fmt.Sprintf("%s", time.Since(s.startTime)),
		UptimeOs:            uptimeOS,
		K8sApiUrl:           k8sUrl,
		K8sVersion:          k8sVersion,
		K8sLatestVersion:    latestK8sVersion,
		K8sCurrentNamespace: k8sCurrentNameSpace,
		EnvVars:             os.Environ(),
		Headers:             map[string][]string{},
	}
	return func(w http.ResponseWriter, r *http.Request) {
		remoteIp := r.RemoteAddr // ip address of the original request or the last proxy
		requestedUrlPath := r.URL.Path
		guid := xid.New()
		s.logger.Info("INFO: 'Request ID: %s'\n", guid.String())
		s.logger.Info(formatTraceRequest, handlerName, r.Method, requestedUrlPath, remoteIp)
		switch r.Method {
		case http.MethodGet:
			if len(strings.TrimSpace(requestedUrlPath)) == 0 || requestedUrlPath == "/info" {
				query := r.URL.Query()
				nameValue := query.Get("name")
				if nameValue != "" {
					data.ParamName = nameValue
				}
				data.Hostname, _ = os.Hostname()
				data.RemoteAddr = remoteIp
				data.Headers = r.Header
				data.Uptime = fmt.Sprintf("%s", time.Since(s.startTime))
				uptimeOS, err := info.GetOsUptime()
				if err != nil {
					s.logger.Error("💥💥 'GetOsUptime() returned an error : %+#v'", err)
				}
				data.UptimeOs = uptimeOS
				data.RequestId = guid.String()
				s.jsonResponse(w, data)
				/*n, err := fmt.Fprintf(w, getHtmlPage(defaultMessage))
				if err != nil {
					s.logger.Printf("💥💥 ERROR: [%s] was unable to Fprintf. path:'%s', from IP: [%s], send_bytes:%d'\n", handlerName, requestedUrlPath, remoteIp, n)
					http.Error(w, "Internal server error. myDefaultHandler was unable to Fprintf", http.StatusInternalServerError)
					return
				}*/
				s.logger.Info("SUCCESS: [%s] path:'%s', from IP: [%s]\n", handlerName, requestedUrlPath, remoteIp)
			} else {
				w.WriteHeader(http.StatusNotFound)
				n, err := fmt.Fprintf(w, getHtmlMsg(version.APP, defaultNotFound))
				if err != nil {
					s.logger.Error("💥💥 [%s] Not Found was unable to Fprintf. path:'%s', from IP: [%s], send_bytes:%d\n", handlerName, requestedUrlPath, remoteIp, n)
					http.Error(w, "Internal server error. myDefaultHandler was unable to Fprintf", http.StatusInternalServerError)
					return
				}
			}
		default:
			s.logger.Info(formatErrRequest, handlerName, r.Method, r.URL.Path, r.RemoteAddr)
			http.Error(w, httpErrMethodNotAllow, http.StatusMethodNotAllowed)
		}
	}
}

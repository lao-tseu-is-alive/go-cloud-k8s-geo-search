package info

import (
	"os"
	"regexp"
	"time"
)

const (
	defaultUnknown     = "_UNKNOWN_"
	defaultReadTimeout = 10 * time.Second // max time to read request from the client
	caCertPath         = "certificates/isrg-root-x1-cross-signed.pem"
)

type RuntimeInfo struct {
	Hostname            string              `json:"hostname"`              // host name reported by the kernel.
	Pid                 int                 `json:"pid"`                   // process id of the caller.
	PPid                int                 `json:"ppid"`                  // process id of the caller's parent.
	Uid                 int                 `json:"uid"`                   // numeric user id of the caller.
	AppName             string              `json:"app_name"`              // name of this application
	Version             string              `json:"version"`               // version of this application
	ParamName           string              `json:"param_name"`            // value of the name parameter (_NO_PARAMETER_NAME_ if name was not set)
	RemoteAddr          string              `json:"remote_addr"`           // remote client ip address
	RequestId           string              `json:"request_id"`            // globally unique request id
	GOOS                string              `json:"goos"`                  // operating system
	GOARCH              string              `json:"goarch"`                // architecture
	Runtime             string              `json:"runtime"`               // go runtime at compilation time
	NumGoroutine        string              `json:"num_goroutine"`         // number of go routines
	OsReleaseName       string              `json:"os_release_name"`       // Linux release Name or _UNKNOWN_
	OsReleaseVersion    string              `json:"os_release_version"`    // Linux release Version or _UNKNOWN_
	OsReleaseVersionId  string              `json:"os_release_version_id"` // Linux release VersionId or _UNKNOWN_
	NumCPU              string              `json:"num_cpu"`               // number of cpu
	Uptime              string              `json:"uptime"`                // tells how long this service was started based on an internal variable
	UptimeOs            string              `json:"uptime_os"`             // tells how long system was started based on /proc/uptime
	K8sApiUrl           string              `json:"k8s_api_url"`           // url for k8s api based KUBERNETES_SERVICE_HOST
	K8sVersion          string              `json:"k8s_version"`           // version of k8s cluster
	K8sLatestVersion    string              `json:"k8s_latest_version"`    // latest version announced in https://kubernetes.io/
	K8sCurrentNamespace string              `json:"k8s_current_namespace"` // k8s namespace of this container
	EnvVars             []string            `json:"env_vars"`              // environment variables
	Headers             map[string][]string `json:"headers"`               // received headers
}

type OsInfo struct {
	Name      string `json:"name"`
	Version   string `json:"version"`
	VersionId string `json:"versionId"`
}

func GetOsUptime() (string, error) {
	uptimeResult := defaultUnknown
	content, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return uptimeResult, err
	}
	uptimeResult = string(content)
	return uptimeResult, nil
}

func GetOsInfo() (*OsInfo, ErrorConfig) {
	const (
		OsReleasePath          = "/etc/os-release"
		regexFindOsNameVersion = `(?m)^NAME="(?P<name>[^"]+)"\s?|^VERSION="(?P<version>[^"]+)"|^VERSION_ID="?(?P<versid>[^"]+)"?\s`
	)
	info := OsInfo{
		Name:      defaultUnknown,
		Version:   defaultUnknown,
		VersionId: defaultUnknown,
	}
	content, err := os.ReadFile(OsReleasePath)
	if err != nil {
		return &info, ErrorConfig{
			Err: err,
			Msg: "GetOsInfo: error reading " + OsReleasePath,
		}
	}
	r := regexp.MustCompile(regexFindOsNameVersion)
	// fmt.Printf("Found matches : %v\n", r.MatchString(string(content)))
	if r.MatchString(string(content)) {
		res := r.FindAllStringSubmatch(string(content), -1)
		for i, v := range res {
			// fmt.Printf("res[%d] : %+#v\n", i, v)
			for j, key := range r.SubexpNames() {
				if j > 0 && i <= len(res) && len(v[j]) > 0 {
					// fmt.Printf("name :'%s' : %+#v\n", key, v[j])
					if key == "name" {
						info.Name = v[j]
					}
					if key == "version" {
						info.Version = v[j]
					}
					if key == "versid" {
						info.VersionId = v[j]
					}
				}
			}
		}
	}
	return &info, ErrorConfig{
		Err: nil,
		Msg: "",
	}
}

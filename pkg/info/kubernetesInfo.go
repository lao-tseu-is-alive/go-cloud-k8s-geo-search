package info

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/gohttpclient"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/golog"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type K8sInfo struct {
	CurrentNamespace string `json:"current_namespace"`
	Version          string `json:"version"`
	Token            string `json:"token"`
	CaCert           string `json:"ca_cert"`
}

type ErrorConfig struct {
	Err error
	Msg string
}

// Error returns a string with an error and a specifics message
func (e *ErrorConfig) Error() string {
	return fmt.Sprintf("%s : %v", e.Msg, e.Err)
}

// GetKubernetesApiUrlFromEnv returns the k8s api url based on the content of standard env var :
//
//	KUBERNETES_SERVICE_HOST
//	KUBERNETES_SERVICE_PORT
//	in case the above ENV variables doesn't  exist the function returns an empty string and an error
func GetKubernetesApiUrlFromEnv() (string, error) {
	srvPort := 443
	k8sApiUrl := "https://"

	var err error
	val, exist := os.LookupEnv("KUBERNETES_SERVICE_HOST")
	if !exist {
		return "", &ErrorConfig{
			Err: err,
			Msg: "ERROR: KUBERNETES_SERVICE_HOST ENV variable does not exist (not inside K8s ?).",
		}
	}
	k8sApiUrl = fmt.Sprintf("%s%s", k8sApiUrl, val)
	val, exist = os.LookupEnv("KUBERNETES_SERVICE_PORT")
	if exist {
		srvPort, err = strconv.Atoi(val)
		if err != nil {
			return "", &ErrorConfig{
				Err: err,
				Msg: "ERROR: CONFIG ENV PORT should contain a valid integer.",
			}
		}
		if srvPort < 1 || srvPort > 65535 {
			return "", &ErrorConfig{
				Err: err,
				Msg: "ERROR: CONFIG ENV PORT should contain an integer between 1 and 65535",
			}
		}
	}
	return fmt.Sprintf("%s:%d", k8sApiUrl, srvPort), nil
}

// GetKubernetesConnInfo returns a K8sInfo with various information retrieved from the current k8s api url
// K8sInfo.CurrentNamespace contains the current namespace of the running pod
// K8sInfo.Token contains the bearer authentication token for the running k8s instance in which this pods lives
// K8sInfo.CaCert contains the certificate of the running k8s instance in which this pods lives
func GetKubernetesConnInfo(l golog.MyLogger, defaultReadTimeout time.Duration) (*K8sInfo, ErrorConfig) {
	const K8sServiceAccountPath = "/var/run/secrets/kubernetes.io/serviceaccount"
	K8sNamespacePath := fmt.Sprintf("%s/namespace", K8sServiceAccountPath)
	K8sTokenPath := fmt.Sprintf("%s/token", K8sServiceAccountPath)
	K8sCaCertPath := fmt.Sprintf("%s/ca.crt", K8sServiceAccountPath)

	info := K8sInfo{
		CurrentNamespace: "",
		Version:          "",
		Token:            "",
		CaCert:           "",
	}

	K8sNamespace, err := os.ReadFile(K8sNamespacePath)
	if err != nil {
		return &info, ErrorConfig{
			Err: err,
			Msg: "GetKubernetesConnInfo: error reading namespace in " + K8sNamespacePath,
		}
	}
	info.CurrentNamespace = string(K8sNamespace)

	K8sToken, err := os.ReadFile(K8sTokenPath)
	if err != nil {
		return &info, ErrorConfig{
			Err: err,
			Msg: "GetKubernetesConnInfo: error reading token in " + K8sTokenPath,
		}
	}
	info.Token = string(K8sToken)

	K8sCaCert, err := os.ReadFile(K8sCaCertPath)
	if err != nil {
		return &info, ErrorConfig{
			Err: err,
			Msg: "GetKubernetesConnInfo: error reading Ca Cert in " + K8sCaCertPath,
		}
	}
	info.CaCert = string(K8sCaCert)

	k8sUrl, err := GetKubernetesApiUrlFromEnv()
	if err != nil {
		return &info, ErrorConfig{
			Err: err,
			Msg: "GetKubernetesConnInfo: error reading GetKubernetesApiUrlFromEnv ",
		}
	}
	urlVersion := fmt.Sprintf("%s/openapi/v2", k8sUrl)
	res, err := gohttpclient.GetJsonFromUrlWithBearerAuth(urlVersion, info.Token, K8sCaCert, l, defaultReadTimeout)
	if err != nil {

		l.Info("GetKubernetesConnInfo: error in GetJsonFromUrl(url:%s) err:%v", urlVersion, err)
		//return &info, ErrorConfig{
		//	Err: Err,
		//	Msg: fmt.Sprintf("GetKubernetesConnInfo: error doing GetJsonFromUrl(url:%s)", urlVersion),
		//}
	} else {
		l.Info("GetKubernetesConnInfo: successfully returned from GetJsonFromUrl(url:%s)", urlVersion)
		var myVersionRegex = regexp.MustCompile("{\"title\":\"(?P<title>.+)\",\"version\":\"(?P<version>.+)\"}")
		match := myVersionRegex.FindStringSubmatch(strings.TrimSpace(res[:150]))
		k8sVersionFields := make(map[string]string)
		for i, name := range myVersionRegex.SubexpNames() {
			if i != 0 && name != "" {
				k8sVersionFields[name] = match[i]
			}
		}
		info.Version = fmt.Sprintf("%s, %s", k8sVersionFields["title"], k8sVersionFields["version"])
	}

	return &info, ErrorConfig{
		Err: nil,
		Msg: "",
	}
}

func GetKubernetesLatestVersion(l golog.MyLogger) (string, error) {
	k8sUrl := "https://kubernetes.io/"
	// Make an HTTP GET request to the Kubernetes releases page
	// Create a new request using http
	req, err := http.NewRequest("GET", k8sUrl, nil)
	if err != nil {
		l.Error("GetKubernetesLatestVersion got error on http.NewRequest [ERROR: %v]\n", err)
		return "", err
	}
	caCert, err := os.ReadFile(caCertPath)
	if err != nil {
		l.Error("GetKubernetesLatestVersion got error on ReadFile(caCertPath) [ERROR: %v]\n", err)
		return "", err
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs: caCertPool,
		},
	}

	//tr := &http.Transport{ TLSClientConfig: &tls.Config{InsecureSkipVerify: true} }

	// add authorization header to the req
	// req.Header.Add("Authorization", bearer)
	// Send req using http Client
	client := &http.Client{
		Timeout:   defaultReadTimeout,
		Transport: tr,
	}

	resp, err := client.Do(req)

	if err != nil {
		l.Error("GetKubernetesLatestVersion got error on response.\n[ERROR] -", err)
		return fmt.Sprintf("GetKubernetesLatestVersion was unable to get content from %s, Error: %v", k8sUrl, err), err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		l.Error("GetKubernetesLatestVersion got error while reading the response bytes:", err)
		return fmt.Sprintf("GetKubernetesLatestVersion got a problem reading the response from %s, Error: %v", k8sUrl, err), err
	}
	// Use a regular expression to extract the latest release number from the page
	re := regexp.MustCompile(`(?m)href=.+?>v(\d+\.\d+)`)
	matches := re.FindAllStringSubmatch(string(body), -1)
	if matches == nil {
		return fmt.Sprintf("GetKubernetesLatestVersion was unable to find latest release number from %s", k8sUrl), nil
	}
	// Print only the release numbers
	maxVersion := 0.0
	for _, match := range matches {
		// fmt.Println(match[1])
		if val, err := strconv.ParseFloat(match[1], 32); err == nil {
			if val > maxVersion {
				maxVersion = val
			}
		}
	}
	// latestRelease := matches[0]
	// fmt.Printf("\nThe latest major release of Kubernetes is %T : %v+", latestRelease, latestRelease)
	return fmt.Sprintf("%2.2f", maxVersion), nil
}

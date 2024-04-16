package gohttpclient

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/golog"
	"io"
	"log"
	"net/http"
	"time"
)

// WaitForHttpServer attempts to establish a http TCP connection to listenAddress
// in a given amount of time. It returns upon a successful connection;
// otherwise exits with an error.
func WaitForHttpServer(url string, waitDuration time.Duration, numRetries int) {
	httpClient := http.Client{
		Timeout: 5 * time.Second,
	}
	for i := 0; i < numRetries; i++ {
		resp, err := httpClient.Get(url)

		if err != nil {
			if i > 0 {
				fmt.Printf("\nWaitForHttpServer: httpClient.Get(%s) retry:[%d], %v\n", url, i, err)
			}
			time.Sleep(waitDuration)
			continue
		}
		// All seems is good
		fmt.Printf("OK: Server responded after %d retries, with status code %d ", i, resp.StatusCode)
		return
	}
	log.Fatalf("Server %s not ready up after %d attempts", url, numRetries)
}

func GetJsonFromUrlWithBearerAuth(url string, token string, caCert []byte, l golog.MyLogger, defaultReadTimeout time.Duration) (string, error) {
	// Create a Bearer string by appending string access token
	var bearer = "Bearer " + token

	// Create a new request using http
	req, err := http.NewRequest("GET", url, nil)

	// add authorization header to the req
	req.Header.Add("Authorization", bearer)
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs: caCertPool,
		},
	}
	// Send req using http Client
	client := &http.Client{
		Transport: tr,
		Timeout:   defaultReadTimeout,
	}
	resp, err := client.Do(req)

	if err != nil {
		l.Error("GetJsonFromUrlWithBearerAuth: Error on response.\n[ERROR] -", err)
		return "", err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			l.Error("GetJsonFromUrlWithBearerAuth: Error on Body.Close().\n[ERROR] -", err)
		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		l.Error("GetJsonFromUrlWithBearerAuth: Error while reading the response bytes:", err)
		return "", err
	}
	return string(body), nil
}

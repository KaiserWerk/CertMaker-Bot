package restclient

import (
	"crypto/tls"
	"fmt"
	"github.com/KaiserWerk/CertMaker-Bot/internal/configuration"
	"net/http"
	"net/url"
	"time"
)

var (
	client     = http.Client{Timeout: 3 * time.Second}
	caHost     = "http://localhost:8880"
	skipVerify = false

)

func GetClient() *http.Client {
	return &client
}

func SetCaHost(h string, skip bool) {
	caHost = h
	skipVerify = skip
}

func GetTlsConfig() *tls.Config {
	config := &tls.Config{
		// CipherSuites not needed with TLS 1.3
		//CipherSuites: []uint16{
		//	tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		//	tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
		//},
		MinVersion:               tls.VersionTLS13,
		PreferServerCipherSuites: true,
		InsecureSkipVerify:       false,
	}
	if skipVerify {
		config.InsecureSkipVerify = false
	}

	return config
}

func ExecuteRequest(r *http.Request) (*http.Response, error) {
	cl := GetClient()

	//fmt.Sprintf("executeRequest: using CA host %s..\n", caHost)
	hostUrl, err := url.ParseRequestURI(caHost)
	if err != nil {
		return nil, fmt.Errorf("could not parse CA host '%s': %s", caHost, err.Error())
	}

	if hostUrl.Scheme == "https" {
		cl.Transport = &http.Transport{TLSClientConfig: GetTlsConfig()}
	}

	//logger.Printf("Using CA Host %s...\n", caHost)

	r.URL.Host = hostUrl.Host
	r.URL.Scheme = hostUrl.Scheme

	config, err := configuration.GetConfiguration()
	if err != nil {
		return nil, err
	}
	r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.CertMaker.ApiKey))

	resp, err := cl.Do(r)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

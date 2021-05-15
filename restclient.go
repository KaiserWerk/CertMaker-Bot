package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

var (
	client     = http.Client{Timeout: 3 * time.Second}
	caHost     = "http://localhost:8880"
	skipVerify = false

)

func getClient() *http.Client {
	return &client
}

func setCaHost(h string, skip bool) {
	caHost = h
	skipVerify = skip
}

func getTlsConfig() *tls.Config {
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

func executeRequest(r *http.Request) (*http.Response, error) {
	cl := getClient()

	//fmt.Sprintf("executeRequest: using CA host %s..\n", caHost)
	hostUrl, err := url.ParseRequestURI(caHost)
	if err != nil {
		return nil, fmt.Errorf("could not parse CA host '%s': %s", caHost, err.Error())
	}

	if hostUrl.Scheme == "https" {
		cl.Transport = &http.Transport{TLSClientConfig: getTlsConfig()}
	}

	//log.Printf("Using CA Host %s...\n", caHost)

	r.URL.Host = hostUrl.Host
	r.URL.Scheme = hostUrl.Scheme

	config, err := getConfiguration()
	if err != nil {
		return nil, err
	}
	r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", config.SimpleCA.ApiKey))

	resp, err := cl.Do(r)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

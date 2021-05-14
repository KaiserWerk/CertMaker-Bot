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
	caHost     = "http://localhost:8880/"
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
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
		},
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

	hostUrl, err := url.ParseRequestURI(caHost)
	if err != nil {
		return nil, err
	}

	if hostUrl.Scheme == "https" {
		cl.Transport = &http.Transport{TLSClientConfig: getTlsConfig()}
	}

	path := r.URL.String()

	fullUrl, err := url.Parse(fmt.Sprintf("%s/%s", caHost, path))
	if err != nil {
		return nil, err
	}

	r.URL = fullUrl

	resp, err := cl.Do(r)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

package restclient

import (
	"crypto/tls"
	"net/http"
	"net/url"
	"time"
)

type RestClient struct {
	client *http.Client
	caHost string
	apiKey string
}

func New(caHost, apiKey string, skipVerify bool) *RestClient {
	return &RestClient{
		client: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{TLSClientConfig: &tls.Config{
				MinVersion:               tls.VersionTLS12,
				PreferServerCipherSuites: true,
				InsecureSkipVerify:       skipVerify,
			}},
		},
		caHost: caHost,
		apiKey: apiKey,
	}
}

func (rc *RestClient) Do(r *http.Request) (*http.Response, error) {
	hostUrl, err := url.ParseRequestURI(rc.caHost)
	if err != nil {
		return nil, err
	}

	r.URL.Host = hostUrl.Host
	r.URL.Scheme = hostUrl.Scheme

	r.Header.Set("X-Auth-Token", rc.apiKey)

	return rc.client.Do(r)
}

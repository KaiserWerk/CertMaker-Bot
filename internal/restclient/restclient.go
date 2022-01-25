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
	config *tls.Config
}

func New(caHost, apiKey string, skipVerify bool) *RestClient {
	return &RestClient{
		client: &http.Client{
			Timeout: 3 * time.Second,
			Transport: &http.Transport{TLSClientConfig: &tls.Config{
				MinVersion:               tls.VersionTLS13,
				PreferServerCipherSuites: true,
				InsecureSkipVerify:       skipVerify,
			}},
		},
		caHost: caHost,
		apiKey: apiKey,
	}
}

func (rc *RestClient) ExecuteRequest(r *http.Request) (*http.Response, error) {
	hostUrl, err := url.ParseRequestURI(rc.caHost)
	if err != nil {
		return nil, err
	}

	r.URL.Host = hostUrl.Host
	r.URL.Scheme = hostUrl.Scheme

	r.Header.Set("X-Auth-Token", rc.apiKey)

	return rc.client.Do(r)
}

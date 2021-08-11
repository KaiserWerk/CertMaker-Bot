package certmaker

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"github.com/KaiserWerk/CertMaker-Bot/internal/entity"
	"github.com/KaiserWerk/CertMaker-Bot/internal/helper"
	"github.com/KaiserWerk/CertMaker-Bot/internal/restclient"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const (
	apiPrefix = "/api/v1"
	routeRequest = apiPrefix + "/certificate/request"

	locationHeaderCertificate = "X-Certificate-Location"
	locationHeaderPrivateKey = "X-Privatekey-Location"

	minCertValidity = 3 // in days
)

func GetRequirementsFromFile(file string) (*entity.CertificateRequirement, error) {
	fileCont, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	var cr entity.CertificateRequirement
	err = yaml.Unmarshal(fileCont, &cr)
	if err != nil {
		return nil, err
	}

	return &cr, nil
}

func CheckIfDueForRenewal(cr *entity.CertificateRequirement, strict bool) error {
	if !helper.FileExists(cr.KeyFile) || !helper.FileExists(cr.CertFile) {
		return nil //fmt.Errorf("certificate or key file does not exist")
	}

	pairFiles, err := tls.LoadX509KeyPair(cr.CertFile, cr.KeyFile)
	if err != nil {
		return err
	}

	cert, err := x509.ParseCertificate(pairFiles.Certificate[0])
	if err != nil {
		return err
	}

	diff := cert.NotAfter.Sub(time.Now())

	if diff.Hours() < 24 * minCertValidity {
		return fmt.Errorf("certificate is invalid; remaining valididy of %f hours is below threshold of %d hours", diff.Hours(), 24 * minCertValidity)
	}

	// TODO check OCSP responder

	return nil
}

func RequestNewKeyAndCert(cr *entity.CertificateRequirement) error {
	jsonCont, err := json.Marshal(cr)
	if err != nil {
		return err
	}
	b := bytes.NewBuffer(jsonCont)

	req, err := http.NewRequest(http.MethodPost, routeRequest, b)
	if err != nil {
		return err
	}

	resp, err := restclient.ExecuteRequest(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("expected response status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	if resp.Header.Get(locationHeaderCertificate) == "" || resp.Header.Get(locationHeaderPrivateKey) == "" {
		return fmt.Errorf("missing headers %s and %s", locationHeaderCertificate, locationHeaderPrivateKey)
	}

	req, err = http.NewRequest(http.MethodGet, resp.Header.Get(locationHeaderCertificate), nil)
	if err != nil {
		return err
	}
	certReq, err := restclient.ExecuteRequest(req)
	if err != nil {
		return err
	}
	err = os.MkdirAll(filepath.Dir(cr.CertFile), 0755)
	if err != nil {
		return err
	}
	dstWriter, err := os.Create(cr.CertFile)
	if err != nil {
		return err
	}
	_, err = io.Copy(dstWriter, certReq.Body)
	if err != nil {
		return err
	}
	_ = certReq.Body.Close()
	_ = dstWriter.Close()

	req, err = http.NewRequest(http.MethodGet, resp.Header.Get(locationHeaderPrivateKey), nil)
	if err != nil {
		return err
	}
	keyReq, err := restclient.ExecuteRequest(req)
	if err != nil {
		return err
	}
	err = os.MkdirAll(filepath.Dir(cr.KeyFile), 0755)
	if err != nil {
		return err
	}
	dstWriter, err = os.Create(cr.KeyFile)
	if err != nil {
		return err
	}
	_, err = io.Copy(dstWriter, keyReq.Body)
	if err != nil {
		return err
	}
	_ = keyReq.Body.Close()
	_ = dstWriter.Close()

	return nil
}


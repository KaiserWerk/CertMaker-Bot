package certmaker

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"github.com/KaiserWerk/CertMaker-Bot/internal/entity"
	helper2 "github.com/KaiserWerk/CertMaker-Bot/internal/helper"
	"github.com/KaiserWerk/CertMaker-Bot/internal/restclient"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
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

func CheckIfDueForRenewal(cr *entity.CertificateRequirement) (bool, error) {
	requestNew := false

	if helper2.FileExists(cr.KeyFile) && helper2.FileExists(cr.CertFile) {
		pairFiles, err := tls.LoadX509KeyPair(cr.CertFile, cr.KeyFile)
		if err != nil {
			return false, err
		}

		cert, err := x509.ParseCertificate(pairFiles.Certificate[0])
		if err != nil {
			return false, err
		}

		diff := cert.NotAfter.Sub(cert.NotBefore)

		if diff.Hours()/24 < 5 { // if validity < 5 days
			requestNew = true
		}
	} else {
		requestNew = true
	}

	return requestNew, nil
}

func RequestNewKeyAndCert(cr *entity.CertificateRequirement) error {
	jsonCont, err := json.Marshal(cr)
	if err != nil {
		return err
	}
	b := bytes.NewBuffer(jsonCont)

	req, err := http.NewRequest(http.MethodPost, "/api/certificate/request", b)
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

	if resp.Header.Get("X-Certificate-Location") == "" || resp.Header.Get("X-Privatekey-Location") == "" {
		return fmt.Errorf("missing headers X-Certificate-Location and X-Privatekey-Location")
	}

	req, err = http.NewRequest(http.MethodGet, resp.Header.Get("X-Certificate-Location"), nil)
	if err != nil {
		return err
	}
	certReq, err := restclient.ExecuteRequest(req)
	if err != nil {
		return err
	}
	err = os.MkdirAll(filepath.Dir(cr.CertFile), 0700)
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

	req, err = http.NewRequest(http.MethodGet, resp.Header.Get("X-Privatekey-Location"), nil)
	if err != nil {
		return err
	}
	keyReq, err := restclient.ExecuteRequest(req)
	if err != nil {
		return err
	}
	err = os.MkdirAll(filepath.Dir(cr.KeyFile), 0700)
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


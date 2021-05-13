package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

var client = http.Client{Timeout: 3 * time.Second}

func requestIfNecessary(file string) error {
	var err error

	fileCont, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	var cr certificateRequirement
	err = json.Unmarshal(fileCont, &cr)
	if err != nil {
		return err
	}

	requestNew := false

	if fileExists(cr.KeyFile) && fileExists(cr.CertFile) {
		pairFiles, err := tls.LoadX509KeyPair(cr.CertFile, cr.KeyFile)
		if err != nil {
			return err
		}

		cert, err := x509.ParseCertificate(pairFiles.Certificate[0])
		if err != nil {
			return err
		}

		diff := cert.NotAfter.Sub(cert.NotBefore)
		fmt.Println("Debug:", diff)
		if diff.Hours() / 24 < 5 { // if validity < 5 days
			requestNew = true
		}
	} else {
		requestNew = true
	}

	if requestNew {
		err = requestNewKeyAndCert(cr)
		if err != nil {
			return err
		}
	}

	return nil
}

func requestNewKeyAndCert(cr certificateRequirement) error {
	jsonCont, err := json.Marshal(cr)
	if err != nil {
		return err
	}
	b := bytes.NewBuffer(jsonCont)
	// TODO base addr from config!!
	req, err := http.NewRequest(http.MethodPost, "/api/certificate/request", b)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("expected response status %d, got %d", http.StatusOK, resp.StatusCode)
	}

	if resp.Header.Get("X-Certificate-Location") == "" || resp.Header.Get("X-Privatekey-Location") == "" {
		return fmt.Errorf("missing headers X-Certificate-Location and X-Privatekey-Location")
	}

	certReq, err := client.Get(resp.Header.Get("X-Certificate-Location"))
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

	keyReq, err := client.Get(resp.Header.Get("X-Privatekey-Location"))
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

	return nil
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
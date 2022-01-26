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
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	apiPrefix    = "/api/v1"
	routeRequest = apiPrefix + "/certificate/request"

	locationHeaderCertificate = "X-Certificate-Location"
	locationHeaderPrivateKey  = "X-Privatekey-Location"

	minCertValidity = 3 * 24 // in days
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

func IsDueForRenewal(cr *entity.CertificateRequirement, strict bool) bool {
	if !helper.FileExists(cr.KeyFile) || !helper.FileExists(cr.CertFile) {
		return true //fmt.Errorf("certificate or key file does not exist")
	}

	pairFiles, err := tls.LoadX509KeyPair(cr.CertFile, cr.KeyFile)
	if err != nil {
		return true
	}

	cert, err := x509.ParseCertificate(pairFiles.Certificate[0])
	if err != nil {
		return true
	}

	diff := cert.NotAfter.Sub(time.Now())

	if diff.Hours() < minCertValidity {
		return true //fmt.Errorf("certificate is invalid; remaining valididy of %f hours is below threshold of %d hours", diff.Hours(), 24 * minCertValidity)
	}

	if strict {
		// TODO check OCSP responder
	}

	return false
}

func RequestNewKeyAndCert(rc *restclient.RestClient, cr *entity.CertificateRequirement) error {
	jsonCont, err := json.Marshal(cr)
	if err != nil {
		return err
	}
	b := bytes.NewBuffer(jsonCont)

	req, err := http.NewRequest(http.MethodPost, routeRequest, b)
	if err != nil {
		return err
	}

	resp, err := rc.ExecuteRequest(req)
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
	certReq, err := rc.ExecuteRequest(req)
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
	keyReq, err := rc.ExecuteRequest(req)
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

func RenewCertificates(reqDir string, logger *logrus.Entry) (uint8, error) {
	// handle certificate requests
	fi, err := ioutil.ReadDir(reqDir)
	if err != nil {
		return 0, fmt.Errorf("could not read files from requirements directory: %s", err.Error())
	}
	logger.Tracef("Found %d files total", len(fi))
	var certsToRenew uint8 = 0
	for _, reqFile := range fi {
		if !strings.HasSuffix(reqFile.Name(), ".yaml") {
			logger.Infof("Ignoring file '%s'; not a yaml file", reqFile.Name())
			continue
		}
		logger.Debugf("Found requirements file '%s'", reqFile.Name())
		fileWithPath := filepath.Join(reqDir, reqFile.Name())
		cr, err := GetRequirementsFromFile(fileWithPath)
		if err != nil {
			logger.Warningf("could not get requirements from file '%s': %s", fileWithPath, err.Error())
			continue
		}

		due := IsDueForRenewal(cr, true)
		if !due {
			logger.Errorf("no need to renew '%s'", reqFile.Name())
			continue
		}

		certsToRenew++
		logger.Debugf("Cert '%s' is due for renewal, requesting...", cr.CertFile)
		err = certmaker.RequestNewKeyAndCert(rc, cr)
		if err != nil {
			certsToRenew--
			logger.Errorf("could not request new key/cert: %s", err.Error())
			continue
		}

		logger.Printf("Cert '%s' successfully renewed!", cr.CertFile)
		// execute optional commands after fetching new cert
		if cr.PostCommands != nil && len(cr.PostCommands) > 0 {
			logger.Debugf("Found %d post operation commands", len(cr.PostCommands))
			for _, commandContent := range cr.PostCommands {
				var cmd *exec.Cmd
				if runtime.GOOS == "linux" {
					cmd = exec.Command("bash", "-c", commandContent)
				} else if runtime.GOOS == "windows" {
					cmd = exec.Command("cmd", "/c", "start", commandContent)
				} else if runtime.GOOS == "darwin" {
					// TODO ?
				}
				logger.Debugf("Command to be executed: %s", cmd.String())
				output, err := cmd.Output()
				if err != nil {
					logger.Warningf("could not execute command '%s': %s", cmd.String(), err.Error())
					continue
				}

				if output != nil && len(output) > 0 {
					logger.Debugf("command '%s' created output: %s", cmd.String(), string(output))
				}
			}
		}
	}

	if certsToRenew > 1 {
		logger.Infof("Renewed %d certificates", certsToRenew)
	} else if certsToRenew == 1 {
		logger.Info("Renewed 1 certificate")
	} else {
		logger.Info("No certificate renewed")
	}
}

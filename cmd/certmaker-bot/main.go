package main

import (
	"flag"
	"fmt"
	"github.com/KaiserWerk/CertMaker-Bot/internal/certmaker"
	"github.com/KaiserWerk/CertMaker-Bot/internal/configuration"
	"github.com/KaiserWerk/CertMaker-Bot/internal/logging"
	"github.com/KaiserWerk/CertMaker-Bot/internal/restclient"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

var (
	Version     = "0.0.0"
	VersionDate = "0000-00-00 00:00:00"

	configFile = flag.String("config", "config.yaml", "The configuration file to use")
	logFile    = flag.String("logfile", "certmaker-bot.log", "The log file to write to in normal mode")
	reqDir     = flag.String("req", "./req", "The folder which contains the certificate requirements")
)

func main() {
	flag.Parse()
	fmt.Println("CertMaker Bot")
	fmt.Printf("\tVersion %s\n", Version)
	fmt.Printf("\tVersion Date %s\n\n", VersionDate)

	if err := os.MkdirAll(*reqDir, 0700); err != nil {
		fmt.Println("could not create requirements directory:", err.Error())
		return
	}

	logger, cleanup, err := logging.New(*logFile, "main")
	if err != nil {
		fmt.Println("could not set up logger:", err.Error())
		return
	}
	defer func() {
		if err := cleanup(); err != nil {
			fmt.Println("could not executed cleanup func:", err.Error())
		}
	}()

	if err := os.MkdirAll(*reqDir, 0755); err != nil {
		logger.WithField("error", err.Error()).Error("could not create requirements directory")
		return
	}

	cfg, created, err := configuration.Setup(*configFile)
	if err != nil {
		logger.WithField("error", err.Error()).Error("could not set up configuration")
		return
	}
	if created {
		logger.Info("configuration file was created")
	}

	rc := restclient.New(cfg.CertMaker.Host, cfg.CertMaker.ApiKey, cfg.CertMaker.SkipVerify)

	logger.Trace("Starting up...")

	t := time.NewTicker(cfg.App.Interval)

	for {
		select {
		case <-t.C:
			logger.Trace(strings.Repeat("-", 20))

			// handle certificate requests
			fi, err := ioutil.ReadDir(*reqDir)
			if err != nil {
				logger.WithField("error", err.Error()).Warn("could not read files from requirements directory")
				continue
			}
			logger.Tracef("Found %d files total", len(fi))
			var certsToRenew uint8 = 0
			for _, reqFile := range fi {
				if !strings.HasSuffix(reqFile.Name(), ".yaml") {
					logger.Infof("Ignoring file '%s'; not a yaml file", reqFile.Name())
					continue
				}
				logger.Debugf("Found requirements file '%s'", reqFile.Name())
				fileWithPath := filepath.Join(*reqDir, reqFile.Name())
				cr, err := certmaker.GetRequirementsFromFile(fileWithPath)
				if err != nil {
					logger.Warningf("could not get requirements from file '%s': %s", fileWithPath, err.Error())
					continue
				}

				due := certmaker.IsDueForRenewal(cr, true)
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
	}

}

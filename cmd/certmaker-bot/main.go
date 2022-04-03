package main

import (
	"flag"
	"fmt"
	"github.com/KaiserWerk/CertMaker-Bot/internal/cert"
	"github.com/KaiserWerk/CertMaker-Bot/internal/configuration"
	"github.com/KaiserWerk/CertMaker-Bot/internal/logging"
	"github.com/KaiserWerk/CertMaker-Bot/internal/restclient"
	"os"
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
	cm := cert.CertMaker{
		Client: rc,
	}
	logger.Trace("Starting up...")

	renewedCerts, errs := cm.RenewCertificates(*reqDir)
	logger.Debugf("renewed %d certificates", renewedCerts)
	if len(errs) > 0 {
		logger.Error("got the following errors on reneweal")
		for _, e := range errs {
			logger.Error(e.Error()) // assuming no error is nil
		}
	}

	t := time.NewTicker(cfg.App.Interval)

	for {
		select {
		case <-t.C:
			logger.Trace(strings.Repeat("-", 20))

			renewedCerts, errs = cm.RenewCertificates(*reqDir)
			logger.Debugf("renewed %d certificates", renewedCerts)
			if len(errs) > 0 {
				logger.Error("got the following errors on reneweal")
				for _, e := range errs {
					logger.Error(e.Error()) // assuming no error is nil
				}
			}
		}
	}

}

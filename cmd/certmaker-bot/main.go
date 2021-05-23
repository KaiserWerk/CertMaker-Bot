package main

import (
	"flag"
	"github.com/KaiserWerk/CertMaker-Bot/internal/certmaker"
	"github.com/KaiserWerk/CertMaker-Bot/internal/configuration"
	"github.com/KaiserWerk/CertMaker-Bot/internal/logging"
	"github.com/KaiserWerk/CertMaker-Bot/internal/restclient"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	Version = "0.0.0"
	VersionDate = "0000-00-00 00:00:00.000 +00:00"
)

var (
	reqDir = "./req"
)

func main() {
	// command line flags
	configFilePtr := flag.String("config", "", "The configuration file to use")
	reqDirPtr := flag.String("req", "", "The folder which contains the certificate requirements")
	asServicePtr := flag.Bool("as-service", false, "Whether to start as a service or not")
	flag.Parse()

	// open the log file
	logHandle, err := os.Create("certmaker-bot.log")
	if err != nil {
		log.Fatal("cannot create log file!")
	}
	defer logHandle.Close()

	// set up logger stuff
	var (
		baseLogger = log.New()
		duration time.Duration
	)
	baseLogger.SetFormatter(&log.JSONFormatter{})
	baseLogger.SetOutput(io.MultiWriter(os.Stdout, logHandle))
	//baseLogger.AddHook()
	if *asServicePtr {
		baseLogger.SetLevel(log.DebugLevel)
		duration = 1 * time.Hour
	} else {
		baseLogger.SetLevel(log.InfoLevel)
		duration = 15 * time.Second
	}
	logger := baseLogger.WithFields(log.Fields{"application": "certmaker-bot", "server": "appsrv.lan", "version": Version})

	logging.SetLogger(logger)

	// configuration stuff
	if *configFilePtr != "" {
		configuration.SetConfigurationFile(*configFilePtr)
	}
	if *reqDirPtr != "" {
		reqDir = *reqDirPtr
	}
	_ = os.MkdirAll(reqDir, 0755)
	conf, err := configuration.SetupConfiguration()
	if err != nil {
		logger.Fatal(err.Error())
	}

	restclient.SetCaHost(conf.CertMaker.Host, conf.CertMaker.SkipVerify)

	logger.Debug("Starting up...")

	for {
		logger.Debug(strings.Repeat("-", 20))

		// handle certificate requests
		fi, err := ioutil.ReadDir(reqDir)
		if err != nil {
			logger.Warning("could not read requirements files: " + err.Error())
			time.Sleep(duration)
			continue
		}
		logger.Printf("Found %d files total", len(fi))
		var certsToRenew uint8 = 0
		for _, reqFile := range fi {
			if !strings.HasSuffix(reqFile.Name(), ".yaml") {
				logger.Infof("Ignoring file '%s'; not a yaml file", reqFile.Name())
				continue
			}
			logger.Debugf("Found requirements file '%s'", reqFile.Name())
			fileWithPath := filepath.Join(reqDir, reqFile.Name())
			cr, err := certmaker.GetRequirementsFromFile(fileWithPath)
			if err != nil {
				logger.Warningf("could not get requirements from file '%s': %s", fileWithPath, err.Error())
				continue
			}

			necessary, err := certmaker.CheckIfDueForRenewal(cr)
			if err != nil {
				logger.Errorf("Could not determine renewal necessity for file '%s': %s", reqFile.Name(), err.Error())
				continue
			}
			if !necessary {
				//too much debug output
				//logger.Printf("Cert '%s' is NOT due for renewal, skipping", cr.CertFile)
			} else {
				certsToRenew++
				logger.Debugf("Cert '%s' is due for renewal, requesting...", cr.CertFile)
				err = certmaker.RequestNewKeyAndCert(cr)
				if err != nil {
					logger.Errorf("could not request new key/cert: %s", err.Error())
					continue
				}
				logger.Printf("Cert '%s' successfully renewed!", cr.CertFile)
				// execute optional commands after fetching new cert
				if cr.PostCommands != nil && len(cr.PostCommands) > 0 {
					logger.Debugf("Found %d post operation commands", len(cr.PostCommands))
					for _, commandContent := range cr.PostCommands {
						cmd := exec.Command("bash", "-c", commandContent)
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
		}

		if certsToRenew > 1 {
			logger.Infof("Renewed %d certificates", certsToRenew)
		} else if certsToRenew == 1 {
			logger.Info("Renewed 1 certificate")
		} else {
			logger.Info("No certificate renewed")
		}

		time.Sleep(duration)
	}
}

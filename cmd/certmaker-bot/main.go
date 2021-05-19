package main

import (
	"flag"
	"fmt"
	"github.com/KaiserWerk/CertMaker-Bot/internal/certmaker"
	"github.com/KaiserWerk/CertMaker-Bot/internal/configuration"
	"github.com/KaiserWerk/CertMaker-Bot/internal/logging"
	"github.com/KaiserWerk/CertMaker-Bot/internal/restclient"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
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
		logger *log.Logger
		duration time.Duration
	)
	if *asServicePtr {
		// log to file as well
		logger = log.New(io.MultiWriter(os.Stdout, logHandle), "", log.LstdFlags)
		duration = 1 * time.Hour
	} else {
		logger = log.New(os.Stdout, "", log.LstdFlags | log.Lmicroseconds | log.Lshortfile)
		duration = 15 * time.Second
	}
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

	logger.Println("Starting up...")

	for {
		logger.Println(strings.Repeat("-", 20))
		// handle certificate requests
		fi, err := ioutil.ReadDir(reqDir)
		if err != nil {
			logger.Fatal("could not read requirements files: " + err.Error())
		}
		logger.Printf("Found %d files total\n", len(fi))
		var certsToRenew uint8 = 0
		for _, reqFile := range fi {
			if !strings.HasSuffix(reqFile.Name(), ".yaml") {
				logger.Printf("Ignoring file '%s'; not a yaml file\n", reqFile.Name())
				continue
			}
			logger.Printf("Found req file '%s'\n", reqFile.Name())
			fileWithPath := filepath.Join(reqDir, reqFile.Name())
			cr, err := certmaker.GetRequirementsFromFile(fileWithPath)
			if err != nil {
				logger.Printf("could not get requirements from file '%s': %s\n", fileWithPath, err.Error())
				continue
			}

			necessary, err := certmaker.CheckIfDueForRenewal(cr)
			if err != nil {
				logger.Printf("Could not determine renewal necessity for file '%s': %s\n", reqFile.Name(), err.Error())
				return
			}
			if !necessary {
				//too much debug output
				//logger.Printf("Cert '%s' is NOT due for renewal, skipping\n", cr.CertFile)
			} else {
				certsToRenew++
				logger.Printf("Cert '%s' is due for renewal, requesting...\n", cr.CertFile)
				err = certmaker.RequestNewKeyAndCert(cr)
				if err != nil {
					logger.Printf("could not request new key/cert: %s\n", err.Error())
					return
				}
				logger.Printf("Cert '%s' successfully renewed!\n", cr.CertFile)
				// execute optional commands after fetching new cert
				if cr.PostCommands != nil && len(cr.PostCommands) > 0 {
					logger.Printf("Found %d post operation commands\n", len(cr.PostCommands))
					for _, commandContent := range cr.PostCommands {
						cmd := exec.Command("bash", "-c", commandContent)
						logger.Printf("Command to be executed: %s\n", cmd.String())
						output, err := cmd.Output()
						if err != nil {
							fmt.Printf("could not execute command '%s': %s\n", cmd.String(), err.Error())
							continue
						}

						if output != nil && len(output) > 0 {
							logger.Printf("command '%s' created output: %s\n", cmd.String(), string(output))
						}
					}
				}
			}
		}

		if certsToRenew > 1 {
			logger.Printf("Renewed %d certificates\n", certsToRenew)
		} else if certsToRenew == 1 {
			logger.Println("Renewed 1 certificate")
		} else {
			logger.Println("No certificate renewed")
		}

		<- time.After(duration)
	}
}

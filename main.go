package main

import (
	"flag"
	"fmt"
	"github.com/robfig/cron/v3"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
)

var (
	certDir = "./req"
)

func main() {
	// command line flags
	configFilePtr := flag.String("c", "", "The configuration file to use")
	certDirPtr := flag.String("d", "", "The folder which contains the certificate requirements (relative or absolute)")
	flag.Parse()

	// configuration stuff
	if *configFilePtr != "" {
		setConfigurationFile(*configFilePtr)
	}
	if *certDirPtr != "" {
		certDir = *certDirPtr
	}
	_ = os.MkdirAll(certDir, 0700)
	conf, err := setupConfiguration()
	if err != nil {
		panic(err.Error())
	}

	setCaHost(conf.SimpleCA.Host, conf.SimpleCA.SkipVerify)

	log.Println("Starting up...")

	// handle certificate requests
	fi, err := ioutil.ReadDir(certDir)
	if err != nil {
		panic("could not read requirements files: " + err.Error())
	}

	c := cron.New()
	for _, reqFile := range fi {
		if !strings.HasSuffix(reqFile.Name(), ".yaml") {
			log.Printf("Ignoring file '%s'; not a yaml file\n", reqFile.Name())
			continue
		}
		fileWithPath := filepath.Join(certDir, reqFile.Name())
		cr, err := getRequirementsFromFile(fileWithPath)
		if err != nil {
			log.Printf("could not get requirements from file '%s': %s\n", fileWithPath, err.Error())
			continue
		}

		_, err = c.AddFunc("@every 20s", func() {

			necessary, err := checkIfDueForRenewal(cr)
			if err != nil {
				log.Printf("Could not determine renewal necessity for file '%s': %s\n", reqFile.Name(), err.Error())
				return
			}
			if !necessary {
				log.Printf("Cert '%s' is NOT due for renewal, skipping\n", cr.CertFile)
			} else {
				log.Printf("Cert '%s' is due for renewal, requesting...\n", cr.CertFile)
				err = requestNewKeyAndCert(cr)
				if err != nil {
					log.Printf("could not request new key/cert: %s\n", err.Error())
					return
				}
				log.Printf("Cert '%s' successfully renewed!\n", cr.CertFile)
				// execute optional commands after fetching new cert
				if cr.PostCommands != nil && len(cr.PostCommands) > 0 {
					log.Printf("Found %d post operation commands\n", len(cr.PostCommands))
					for _, commandContent := range cr.PostCommands {
						cmd := exec.Command("bash", "-c", commandContent)
						log.Printf("Command to be executed: %s\n", cmd.String())
						output, err := cmd.Output()
						if err != nil {
							fmt.Printf("could not execute command '%s': %s\n", cmd.String(), err.Error())
							continue
						}

						if output != nil && len(output) > 0 {
							log.Printf("command '%s' created output: %s\n", cmd.String(), string(output))
						}
					}
				}
			}
		})
		if err != nil {
			log.Printf("Could not enqueue file '%s': %s\n", reqFile.Name(), err.Error())
			continue
		}
	}
	c.Start()

	// block until further notice
	notify := make(chan os.Signal)
	signal.Notify(notify, os.Interrupt, os.Kill, syscall.SIGTERM)
	<-notify
	log.Println("Shutting down...")
}

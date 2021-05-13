package main

import (
	"flag"
	"github.com/robfig/cron/v3"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path/filepath"
)

var (
	certDir = "./req"
)

func main() {
	configFilePtr := flag.String("c", "", "The configuration file to use")
	certDirPtr := flag.String("d", "", "The folder which contains the certificate requirements (relative or absolute)")
	flag.Parse()

	if *configFilePtr != "" {
		setConfigurationFile(*configFilePtr)
	}

	if *certDirPtr != "" {
		certDir = *certDirPtr
	}

	// TODO create config file if it doesnt exist

	log.Println("Starting up...")

	fi, err := ioutil.ReadDir(certDir)
	if err != nil {
		panic("could not read requirements files: " + err.Error())
	}

	c := cron.New()
	for _, file := range fi {
		_, err := c.AddFunc("@daily", func() {
			err := requestIfNecessary(filepath.Join(certDir, file.Name()))
			if err != nil {
				log.Printf("Could not enqueue file '%s' as job, skipping: %s\n", file.Name(), err.Error())
			}
		})
		if err != nil {
			log.Printf("Could not enqueue file '%s': %s\n", file.Name(), err.Error())
			continue
		}
	}
	c.Start()

	notify := make(chan os.Signal)
	signal.Notify(notify, os.Interrupt)
	<-notify
	log.Println("Shutting down...")


}

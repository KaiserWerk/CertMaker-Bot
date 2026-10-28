package main

import (
	"os"

	"github.com/KaiserWerk/CertMaker-Bot/internal/entity"
	"gopkg.in/yaml.v2"
)

func main() {
	req := entity.CertificateRequirement{
		Domains: []string{"example.com"},
		IPs:     []string{"127.0.0.1", "192.168.178.15"},
		Subject: entity.Subject{
			Organization:  "My Org",
			Country:       "Germany",
			Province:      "NRW",
			Locality:      "My City",
			StreetAddress: "Example street 33",
			PostalCode:    "12345",
		},
		Days:         35,
		CertFile:     "/target/dir/cert.pem",
		KeyFile:      "/target/dir/key.pem",
		PostCommands: []string{},
	}

	y, _ := yaml.Marshal(req)

	os.WriteFile("req.yaml", y, 0o600)
}

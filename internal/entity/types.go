package entity

type CertificateRequirement struct {
	Domains      []string `yaml:"domains"`
	IPs          []string `yaml:"ips"`
	Subject      Subject  `yaml:"subject,omitempty"`
	Days         int      `yaml:"days"`
	CertFile     string   `yaml:"cert_file"`
	KeyFile      string   `yaml:"key_file"`
	PostCommands []string `yaml:"post_commands,omitempty"`
}

type Subject struct {
	Organization  string `yaml:"organization"`
	Country       string `yaml:"country"`
	Province      string `yaml:"province"`
	Locality      string `yaml:"locality"`
	StreetAddress string `yaml:"street_address"`
	PostalCode    string `yaml:"postal_code"`
}

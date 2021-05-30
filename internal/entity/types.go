package entity

type Configuration struct {
	CertMaker struct {
		Host       string `yaml:"host"`
		SkipVerify bool   `yaml:"skip_verify"`
		ApiKey     string `yaml:"apikey"`
	} `yaml:"certmaker"`
}

type CertificateRequirement struct {
	Domains []string `yaml:"domains"`
	IPs     []string `yaml:"ips"`
	Subject struct {
		Organization  string `yaml:"organization"`
		Country       string `yaml:"country"`
		Province      string `yaml:"province"`
		Locality      string `yaml:"locality"`
		StreetAddress string `yaml:"street_address"`
		PostalCode    string `yaml:"postal_code"`
	} `yaml:"subject,omitempty"`
	Days     int    `yaml:"days"`
	CertFile string `yaml:"cert_file"`
	KeyFile  string `yaml:"key_file"`
	PostCommands []string `yaml:"post_commands,omitempty"`
}
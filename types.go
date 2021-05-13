package main

type configuration struct {
	SimpleCA struct{
		Host string `json:"host"`
		ApiKey string `json:"apikey"`
	} `json:"simpleca"`
}

type certificateRequirement struct {
	Domains []string `json:"domains"`
	IPs     []string `json:"ips"`
	Subject struct {
		Organization  string `json:"organization"`
		Country       string `json:"country"`
		Province      string `json:"province"`
		Locality      string `json:"locality"`
		StreetAddress string `json:"street_address"`
		PostalCode    string `json:"postal_code"`
	} `json:"subject"`
	Days int `json:"days"`
	CertFile string `json:"cert_file"`
	KeyFile string `json:"key_file"`
}

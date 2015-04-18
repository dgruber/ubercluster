package main

import (
	"fmt"
	"github.com/spf13/viper"
)

// ProxyConfig contains the current configuration for the proxy
// which can come from command line or from configuration file.
type ProxyConfig struct {
	Verbose        bool
	AddressPort    string   // like http://localhost:8888
	CertFile       string   // Path to certification file in secure mode
	KeyFile        string   // Path to key file in secure mode
	OTP            string   // one time password or "yubikey"
	YubiID         string   // For yubikey support -> you get this from https://upgrade.yubico.com/getapikey/
	YubiSecret     string   // For yubikey support -> register your service above
	YubiAllowedIds []string // For yubikey support -> list of IDs of yubikeys which are allowed
	// section for enhanced multi-clustering (exchange of jobs between clusters)
	DistributeTo         []string // Active job distribution: List of clusters to distribute
	DistributeAcceptFrom []string // Active job distributeion: List of clusters which can accept jobs
	FetchFrom            []string // Active fetching for jobs: List of clusters which are queried
	FetchAcceptFrom      []string // Active fetching for jobs: List of clusters which are allowed to serve when they are actively requesing jobs
}

func (c ProxyConfig) String() string {
	return fmt.Sprintf("Verbose: %t\nAdress: %s\nCertFile: %s\nKeyFile: %s\nYubiID: %s\nAllowedIDs: %s\n",
		c.Verbose, c.AddressPort, c.CertFile, c.KeyFile, c.YubiID, c.YubiAllowedIds)
}

func initializeD2Proxy() (*ProxyConfig, error) {
	// configuration for proxy startup
	var config ProxyConfig
	// simplify configuration through creating a d2proxyConfig.json
	// in the directory where you start the proxy
	viper.SetConfigName("d2proxyConfig")
	viper.AddConfigPath("./")
	// use config only when there is a config file in local directory
	if err := viper.ReadInConfig(); err == nil {
		err = viper.Marshal(&config)
		if err != nil {
			return nil, err
		}
		return &config, nil
	}
	return nil, nil
}

package main

import (
	"fmt"
	"github.com/spf13/viper"
	"os"
)

// configuration for proxy startup
var config ProxyConfig

type ProxyConfig struct {
	Verbose     bool
	AddressPort string // like http://localhost:8888
	CertFile    string // Path to certification file in secure mode
	KeyFile     string // Path to key file in secure mode
}

func (c ProxyConfig) String() string {
	return fmt.Sprintf("Verbose: %b\nAdress: %s\nCertFile: %sKeyFile: %s\n",
		c.Verbose, c.AddressPort, c.CertFile, c.KeyFile)
}

func initializeD2Proxy() bool {
	// simplify configuration through creating a d2proxyConfig.json
	// in the directory where you start the proxy
	viper.SetConfigName("d2proxyConfig")
	viper.AddConfigPath("./")
	// use config only when there is a config file in local directory
	if err := viper.ReadInConfig(); err == nil {
		if err := viper.Marshal(&config); err != nil {
			fmt.Println("Error when decoding config file. ", err)
			os.Exit(1)
		} else {
			return true
		}
	}
	return false
}

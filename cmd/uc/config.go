/*
   Copyright 2014 Daniel Gruber, Univa

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package main

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/viper"
	"os"
)

// configuration for proxies of compute clusters which can be queried
var config Config

type ClusterConfig struct {
	// name to reference the cluster in this tool ("default" is
	// the address used when no cluster is explicitly referenced
	Name            string
	Address         string // like http://localhost:8888
	ProtocolVersion string // the protocol the proxy speaks
}

func (c ClusterConfig) String() string {
	return fmt.Sprintf("Name: %s\nAddress: %s\nProtocolVersion: %s\n", c.Name, c.Address, c.ProtocolVersion)
}

// Complete configuration
type Config struct {
	// Multiple endpoints of proxies can be defined
	Cluster []ClusterConfig
}

// saveDummyConfig creates a file dummyconfig.json in order
// to help the user to create a configuration file (config.json)
// in the right JSON format. "default" and "cluster1" are
// user defined names for the cluster, while the address is
// the endpoint of the proxy.
func saveDummyConfig() {
	if file, err := os.Create("dummyconfig.json"); err == nil {
		encoder := json.NewEncoder(file)
		var config Config
		config.Cluster = make([]ClusterConfig, 0)
		var def, cluster ClusterConfig
		def.Name = "default"
		def.Address = "http://localhost:8888/"
		def.ProtocolVersion = "v1"
		cluster.Name = "cluster1"
		cluster.Address = "http://localhost:8282/"
		cluster.ProtocolVersion = "v1"
		config.Cluster = append(config.Cluster, def)
		config.Cluster = append(config.Cluster, cluster)
		encoder.Encode(config)
		file.Close()
	}
}

func readConfig() {
	viper.SetConfigName("config")
	// check local directory first
	viper.AddConfigPath("./")
	// then home directory
	viper.AddConfigPath("$HOME/.ubercluster/")
	// finally /etc
	viper.AddConfigPath("/etc/ubercluster/")

	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("Error reading in config file. ", err)
		os.Exit(1)
	}
	if err := viper.Marshal(&config); err != nil {
		fmt.Println("Error when decoding config file. ", err)
		os.Exit(1)
	}
}

func listConfig(clusteraddress string) {
	for _, cc := range config.Cluster {
		fmt.Println(cc)
	}
}

// setClusterAddress searches the address of the cluster to contact to
// in the configuration ("default" point to default cluster)
func getClusterAddress(cluster string) string {
	var clusteraddress string
	for i, _ := range config.Cluster {
		if cluster == config.Cluster[i].Name {
			clusteraddress = config.Cluster[i].Address
			clusteraddress = fmt.Sprintf("%s%s", clusteraddress, config.Cluster[i].ProtocolVersion)
			break
		}
	}
	if clusteraddress == "" {
		fmt.Println("Cluster name %s not found in configuration.", cluster)
		os.Exit(1)
	}
	return clusteraddress
}

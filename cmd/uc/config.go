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
	"errors"
	"fmt"
	"github.com/spf13/viper"
	"log"
	"os"
	"strconv"
)

// Config contains configuration for proxies of compute clusters which can be queried.
var config Config

// GlobalConfig contains global configuration parameters.
var globalConfig GlobalConfig

// ClusterConfig contains all neccessary information ot access
// one cluster which is represented by a proxy.
type ClusterConfig struct {
	// name to reference the cluster in this tool ("default" is
	// the address used when no cluster is explicitly referenced
	Name            string
	Address         string // like http://localhost:8888
	ProtocolVersion string // the protocol the proxy speaks "v1"
}

func (c ClusterConfig) String() string {
	return fmt.Sprintf("Name: %s\nAddress: %s\nProtocolVersion: %s\n", c.Name, c.Address, c.ProtocolVersion)
}

// Config contains the complete configuration for all clusters. The
// configuration is intended to be read out from a config file.
type Config struct {
	// Multiple endpoints of proxies can be defined
	Cluster []ClusterConfig
}

// GlobalConfig is the merged configuration containing the
// configuration items needed in later function calls.
type GlobalConfig struct {
	OTP string
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

func ReadConfig() Config {
	viper.SetConfigName("config")
	// check local directory first
	viper.AddConfigPath("./")
	// then home directory
	viper.AddConfigPath("$HOME/.ubercluster/")
	// finally /etc
	viper.AddConfigPath("/etc/ubercluster/")

	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("Error reading in config file: %s\n", err)
		os.Exit(1)
	}

	if err := viper.Unmarshal(&config); err != nil {
		fmt.Printf("Internal error parsing config file: %s\n", err)
		os.Exit(1)
	}
	return config
}

func listConfig(clusteraddress string) {
	for _, cc := range config.Cluster {
		fmt.Println(cc)
	}
}

// GetClusterAddress searches the address of the cluster to contact to
// in the configuration ("default" point to default cluster)
func GetClusterAddress(cluster string) (string, string, error) {
	var clusteraddress string
	for i := range config.Cluster {
		if cluster == config.Cluster[i].Name {
			clusteraddress = config.Cluster[i].Address
			clusteraddress = fmt.Sprintf("%s%s", clusteraddress, config.Cluster[i].ProtocolVersion)
			break
		}
	}
	if clusteraddress == "" {
		text := fmt.Sprintf("Cluster %s not found in configuration", cluster)
		fmt.Printf("%s\n", text)
		return "", "", errors.New(text)
	}
	log.Println("Chosen cluster: ", cluster, clusteraddress)
	return clusteraddress, cluster, nil
}

// makeTestConfig creates a configuration for testing
func makeTestConfig(amount int) Config {
	var conf Config
	conf.Cluster = make([]ClusterConfig, amount, amount)
	for i := 0; i < amount; i++ {
		conf.Cluster[i].Name = "cluster" + strconv.Itoa(i)
		conf.Cluster[i].Address = "10.0.0." + strconv.Itoa(i%255)
		conf.Cluster[i].ProtocolVersion = "v1"
	}
	return conf
}

func selectClusterAddress(cluster, alg string) (string, string, error) {
	// a cluster selection algorithm chooses the right cluster
	switch alg {
	case "rand": // random scheduling
		return GetClusterAddress(MakeNewScheduler(RandomSchedulerType, config).Impl.SelectCluster())
	case "prob": // probabilistic scheduling
		return GetClusterAddress(MakeNewScheduler(ProbabilisticSchedulerType, config).Impl.SelectCluster())
	case "load": // load based scheduling
		return GetClusterAddress(MakeNewScheduler(LoadBasedSchedulerType, config).Impl.SelectCluster())
	}
	if alg != "" {
		fmt.Println("Unkown scheduler selection algorithm: ", alg)
		os.Exit(2)
	}
	return GetClusterAddress(cluster)
}

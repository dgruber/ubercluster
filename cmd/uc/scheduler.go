/*
   Copyright 2015 Daniel Gruber, info@gridengine.eu

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
	"log"
	"math"
	"math/rand"
	"sync"
	"time"
)

var seeded bool = false

// Implements the cluster selection algorithms

// probabilisticScheduler returns the name of the selected
// cluster from the configuration. The selection is based
// on the cluster load and selects a valid cluster (one
// with a lower load than 1). A cluster with a load of 0
// has a higher probability to be chosen than one with 0.9.
// If all clusters have the same load all of them have the
// same probability to be chosen.
func probabilisticScheduler(conf Config) string {
	// TODO
	return conf.Cluster[0].Name
}

type loadValues struct {
	sync.Mutex
	sync.WaitGroup
	load []float64
}

func getClusterLoad(lv *loadValues, index int, request string) {
	if resp, err := uberGet(request); err == nil {
		defer resp.Body.Close()
		decoder := json.NewDecoder(resp.Body)
		var load float64
		if err := decoder.Decode(&load); err != nil {
			lv.load[index] = load
		} else {
			log.Println("Error during decoding cluster load from ", request, err)
		}
	}
	lv.Done()
}

func getAllLoadValues(conf Config) []float64 {
	var lv loadValues
	lv.load = make([]float64, len(conf.Cluster), len(conf.Cluster))
	lv.Add(len(conf.Cluster))
	for i, _ := range conf.Cluster {
		addr := conf.Cluster[i].Address
		ver := conf.Cluster[i].ProtocolVersion
		go getClusterLoad(&lv, i, fmt.Sprintf("%s/%s/drmsload", addr, ver))
	}
	lv.Wait()
	return lv.load
}

func minLoad(load []float64) int {
	min := math.MaxFloat64
	index := 0
	for k, v := range load {
		if v < min {
			min = v
			index = k
		}
	}
	return index
}

// loadBasedScheduler is a simple scheduler that selects
// the cluster with the lowest load.
func loadBasedScheduler(conf Config) string {
	// get all load values (time consuming)
	load := getAllLoadValues(conf)
	return conf.Cluster[minLoad(load)].Name
}

// randomScheduler selects a cluster purely radomly
func randomScheduler(conf Config) string {
	if !seeded {
		rand.Seed(time.Now().UTC().UnixNano())
	}
	return conf.Cluster[rand.Intn(len(conf.Cluster))].Name
}

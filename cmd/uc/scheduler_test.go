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
	"fmt"
	"testing"
)

func TestProbabilisticSelection(t *testing.T) {
	amount := 1000000
	// the load of the clusters
	var p = []float64{
		0.9, 0.8, 1.0, 0.3,
	}
	distribution := make([]int, 4, 4)
	selection := make([]int, amount, amount)
	for i := 0; i < amount; i++ {
		selection[i] = probabilisticSelection(p)
		distribution[selection[i]]++
	}
	// expecting to have 10%, 20%, 0% and 70%
	for i := 0; i < 4; i++ {
		if p[i] == 1.0 {
			if distribution[i] > 10 {
				t.Errorf("Amount of selections %d exceeded 10 but should be 0.",
					distribution[i])
			}
		} else {
			if float64(distribution[i]) > ((1.0-p[i])*float64(amount))*1.01 ||
				float64(distribution[i]) < ((1.0-p[i])*float64(amount))*0.09 {
				t.Errorf("Expected amount of selections of %d differs more than 1 percent (%d but got %d)",
					i, int((1.0-p[i])*float64(amount)), distribution[i])
			}
		}
		fmt.Printf("%d has %d\n", i, distribution[i])
	}
}

func BenchmarkProbabilisticSelection(b *testing.B) {
	var loads = []float64{
		0.9, 0.8, 1.0, 0.3, 0.1, 0.2, 0.4,
	}
	for i := 0; i < b.N; i++ {
		probabilisticSelection(loads)
	}
}

func TestRandomScheduling(t *testing.T) {
	for amountOfCluster := 1; amountOfCluster < 10; amountOfCluster++ {
		conf := makeTestConfig(amountOfCluster)
		sched := MakeNewScheduler(RandomSchedulerType, conf)
		names := make([]string, 10000, 10000)
		for i := 0; i < 10000; i++ {
			names[i] = sched.Impl.SelectCluster()
		}
		// check how many different names are selected
		m := make(map[string]int)
		for i := 0; i < 10000; i++ {
			if v, exists := m[names[i]]; exists {
				m[names[i]] = v + 1
			} else {
				m[names[i]] = 1
			}
		}
		// each cluster schoould be selected not more than 10% more or less
		// than each other
		expectedOccurence := int(10000 / amountOfCluster)
		for k, v := range m {
			if v > int(float64(expectedOccurence)*1.2) {
				t.Errorf("Cluster %s was choosen too often: %d\n", k, v)
			}
			if v < int(float64(expectedOccurence)*0.9) {
				t.Errorf("Cluster %s was choosen too less: %d\n", k, v)
			}
			//fmt.Printf("Good! Cluster %s was chosen %d times\n", k, v)
		}
	}
}

func BenchmarkRandomScheduling(b *testing.B) {
	conf := makeTestConfig(10)
	sched := MakeNewScheduler(RandomSchedulerType, conf)
	for i := 0; i < b.N; i++ {
		sched.Impl.SelectCluster()
	}
}

func BenchmarkLoadBasedScheduling(b *testing.B) {
	// doesn't make much sense since it tries to get the load
	// from the clusters (which does not exist of course)
	conf := makeTestConfig(10)
	sched := MakeNewScheduler(LoadBasedSchedulerType, conf)
	for i := 0; i < b.N; i++ {
		sched.Impl.SelectCluster()
	}
}

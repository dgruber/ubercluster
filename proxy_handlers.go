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

package ubercluster

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
)

func getDRMAA2JobState(state string) JobState {
	switch state {
	case "r":
		return Running
	case "q":
		return Queued
	case "h":
		return QueuedHeld
	case "s":
		return Suspended
	case "R":
		return Requeued
	case "Rh":
		return RequeuedHeld
	case "d":
		return Done
	case "f":
		return Failed
	case "u":
		return Undetermined
	}
	return Undetermined
}

func MakeMSessionJobInfosHandler(impl ProxyImplementer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		filterSet := false
		var filter JobInfo
		if state := r.FormValue("state"); state != "all" && state != "" {
			filter.State = getDRMAA2JobState(state)
			log.Println("filter for state: ", filter.State)
			filterSet = true
		}
		if user := r.FormValue("user"); user != "" {
			filter.JobOwner = user
			log.Println("filter for user: ", filter.JobOwner)
			filterSet = true
		}
		if jobinfos := impl.GetJobInfosByFilter(filterSet, filter); jobinfos != nil {
			encoder := json.NewEncoder(w)
			if err := encoder.Encode(jobinfos); err != nil {
				fmt.Println("Encoding error: ", err)
			} else {
				log.Println("Encoded: ", jobinfos)
			}
		}
	}
}

func MakeMSessionJobInfoHandler(impl ProxyImplementer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		if jobid := vars["jobid"]; jobid != "" {
			if jobinfo := impl.GetJobInfo(jobid); jobinfo != nil {
				json.NewEncoder(w).Encode(*jobinfo)
			}
		}
	}
}

func MakeMachinesHandler(impl ProxyImplementer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if machines, err := impl.GetAllMachines(nil); err == nil {
			json.NewEncoder(w).Encode(machines)
		} else {
			log.Println("Error in GetAllMachines: ", err)
		}
	}
}

func MakeMachineHandler(impl ProxyImplementer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		name := vars["name"]
		if machines, err := impl.GetAllMachines([]string{name}); err == nil {
			json.NewEncoder(w).Encode(machines)
		} else {
			log.Println("Error in GetAllMachines: ", err)
		}
	}
}

func MakeQueuesHandler(impl ProxyImplementer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if queues, err := impl.GetAllQueues(nil); err == nil {
			json.NewEncoder(w).Encode(queues)
		} else {
			log.Println("Error in GetAllQueues: ", err)
		}
	}
}

func MakeQueueHandler(impl ProxyImplementer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		name := vars["name"]
		if queues, err := impl.GetAllQueues([]string{name}); err == nil {
			json.NewEncoder(w).Encode(queues)
		} else {
			log.Println("Error in GetAllQueues: ", err)
		}
	}
}

func MakeJSessionCategoriesHandler(impl ProxyImplementer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if categories, err := impl.GetAllCategories(); err == nil {
			json.NewEncoder(w).Encode(categories)
		} else {
			log.Println("Error in GetAllCategoires: ", err)
		}
	}
}

func MakeJSessionCategoryHandler(impl ProxyImplementer) http.HandlerFunc {
	// at the moment all job sessions have the same categories
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		name := vars["category"]
		if categories, err := impl.GetAllCategories(); err == nil {
			for _, c := range categories {
				if c == name {
					json.NewEncoder(w).Encode(c)
					return
				}
			}
		} else {
			log.Println("Error in GetJobCategories: ", err)
		}
	}
}

func MakeMSessionDRMSNameHandler(impl ProxyImplementer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(impl.DRMSName())
	}
}

func MakeMSessionDRMSVersionHandler(impl ProxyImplementer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(impl.DRMSVersion())
	}
}

func MakeMSessionDRMSLoadHandler(impl ProxyImplementer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(impl.DRMSLoad())
	}
}

// Reads in JSON for DRMAA2 job template struct.
func MakeJSessionSubmitHandler(impl ProxyImplementer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if body, err := ioutil.ReadAll(r.Body); err != nil {
			log.Println("(proxy)", err)
		} else {
			var jt JobTemplate
			if uerr := json.Unmarshal(body, &jt); uerr != nil {
				log.Println("(proxy) Unmarshall error")
				http.Error(w, uerr.Error(), http.StatusInternalServerError)
			} else {
				log.Println("(proxy) Submit now job")
				// Submit job in compute cluster
				if jobid, joberr := impl.RunJob(jt); joberr != nil {
					log.Println("(proxy) Error during job submission: ", joberr)
					http.Error(w, uerr.Error(), http.StatusInternalServerError)

				} else {
					log.Println("(proxy) Job successfully submitted: ", jobid)
				}
			}
		}
	}
}

func MakeJSessionJobManipulationHandler(impl ProxyImplementer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		name := vars["jsname"]
		operation := vars["operation"]
		jobid := vars["jobid"]
		log.Println("(jobManipulationHandler) called with: ", name, operation, jobid)

		// job session name must be the one created by d2proxy
		if name != "ubercluster" {
			json.NewEncoder(w).Encode("invalid job session name")
			return
		}
		if str, err := impl.JobOperation(name, operation, jobid); err == nil {
			json.NewEncoder(w).Encode(str)
		} else {
			json.NewEncoder(w).Encode(err)
		}
	}
}

func AutenticationErrorHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Authentication error")
	http.NotFound(w, r)
}

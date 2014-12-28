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
	"github.com/dgruber/drmaa2"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
)

func getDRMAA2JobState(state string) drmaa2.JobState {
	switch state {
	case "r":
		return drmaa2.Running
	case "q":
		return drmaa2.Queued
	case "h":
		return drmaa2.QueuedHeld
	case "s":
		return drmaa2.Suspended
	case "R":
		return drmaa2.Requeued
	case "Rh":
		return drmaa2.RequeuedHeld
	case "d":
		return drmaa2.Done
	case "f":
		return drmaa2.Failed
	case "u":
		return drmaa2.Undetermined
	}
	return drmaa2.Undetermined
}

func msessionJobInfosHandler(w http.ResponseWriter, r *http.Request) {
	filterSet := false
	var ji *drmaa2.JobInfo
	filter := drmaa2.CreateJobInfo()
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
	if filterSet == true {
		ji = &filter
	}
	if jobinfos := getJobInfosByFilter(ms, ji); jobinfos != nil {
		encoder := json.NewEncoder(w)
		if err := encoder.Encode(jobinfos); err != nil {
			fmt.Println("Encoding error: ", err)
		} else {
			log.Println("Encoded: ", jobinfos)
		}
	}
}

func msessionJobInfoHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	if jobid := vars["jobid"]; jobid != "" {
		if jobinfo := getJobInfo(ms, jobid); jobinfo != nil {
			json.NewEncoder(w).Encode(jobinfo)
		}
	}
}

func monitoringSessionHandler(w http.ResponseWriter, r *http.Request) {
	// get all jobs in a certain state
	if state := r.FormValue("state"); state != "" {
		if jobinfo := getJobInfoByState(ms, state); jobinfo != nil {
			encoder := json.NewEncoder(w)
			if err := encoder.Encode(jobinfo); err != nil {
				fmt.Println("Encoding error: ", err)
			}
		}
		return
	}

	// get all jobs from DRM
	if jobs := r.FormValue("jobs"); jobs != "" {
		if jobs == "all" {
			if js, err := ms.GetAllJobs(nil); err == nil {
				ji := make([]drmaa2.JobInfo, 0, 500)
				encoder := json.NewEncoder(w)
				for i, _ := range js {
					jinfo, _ := js[i].GetJobInfo()
					ji = append(ji, *jinfo)
				}
				encoder.Encode(ji)
			}
			return
		}
	}

}

func msessionMachinesHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("(msessionMachinesHandler)")
	if machines, err := ms.GetAllMachines(nil); err == nil {
		json.NewEncoder(w).Encode(machines)
	} else {
		log.Println("Error in GetAllMachines: ", err)
	}
}

func msessionMachineHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("(msessionMachineHandler)")
	vars := mux.Vars(r)
	name := vars["name"]
	if machines, err := ms.GetAllMachines([]string{name}); err == nil {
		json.NewEncoder(w).Encode(machines)
	} else {
		log.Println("Error in GetAllMachines: ", err)
	}
}

func msessionQueuesHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("(msessionQueuesHandler)")
	if queues, err := ms.GetAllQueues(nil); err == nil {
		json.NewEncoder(w).Encode(queues)
	} else {
		log.Println("Error in GetAllQueues: ", err)
	}
}

func msessionQueueHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("(msessionQueueHandler)")
	vars := mux.Vars(r)
	name := vars["name"]
	if machines, err := ms.GetAllQueues([]string{name}); err == nil {
		json.NewEncoder(w).Encode(machines)
	} else {
		log.Println("Error in GetAllQueues: ", err)
	}
}

func msessionDRMSNameHandler(w http.ResponseWriter, r *http.Request) {
	var sm drmaa2.SessionManager
	if name, err := sm.GetDrmsName(); err == nil {
		json.NewEncoder(w).Encode(name)
	} else {
		log.Println("Error in SessionManager.GetDrmsName(): ", err)
	}
}

func msessionDRMSVersionHandler(w http.ResponseWriter, r *http.Request) {
	var sm drmaa2.SessionManager
	if version, err := sm.GetDrmsVersion(); err == nil {
		json.NewEncoder(w).Encode(version.String())
	} else {
		log.Println("Error in SessionManager.GetDrmsVersion(): ", err)
	}
}

// Reads in JSON for DRMAA2 job template struct.
func jobSubmitHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("(jobSubmitHandler)")
	if body, err := ioutil.ReadAll(r.Body); err != nil {
		log.Println("(proxy)", err)
	} else {
		var jt drmaa2.JobTemplate
		if uerr := json.Unmarshal(body, &jt); uerr != nil {
			log.Println("(proxy) Unmarshall error")
			http.Error(w, uerr.Error(), http.StatusInternalServerError)
		} else {
			log.Println("(proxy) Submit now job")
			// Submit job in compute cluster
			if job, joberr := js.RunJob(jt); joberr != nil {
				log.Println("(proxy) Error duing job submission: ", joberr)
				http.Error(w, uerr.Error(), http.StatusInternalServerError)

			} else {
				log.Println("(proxy) Job successfully submitted: ", job.GetId())
			}
		}
	}
}

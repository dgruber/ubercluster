/*
   Copyright 2014 Daniel Gruber

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
	"flag"
	"fmt"
	"github.com/dgruber/drmaa2"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

func init() {
	// Disable logging by default
	log.SetOutput(ioutil.Discard)
}

// DRMAA2 Sessions required for accessing the cluster
var ms *drmaa2.MonitoringSession
var js *drmaa2.JobSession

// Returns a DRMAA2 JobInfo struct based on the given jobid.
func getJobInfo(jobid string) *drmaa2.JobInfo {
	var jobinfo *drmaa2.JobInfo
	filter := drmaa2.CreateJobInfo()
	filter.Id = jobid
	if job, err := ms.GetAllJobs(&filter); err != nil || job == nil {
		filter.Id = fmt.Sprintf("%s.1", jobid)
		if job2, err2 := ms.GetAllJobs(&filter); err2 != nil || job2 == nil {
			log.Printf("No job found")
		} else {
			jobinfo, _ = job2[0].GetJobInfo()
		}
	} else {
		log.Printf("amount of matching jobs %d\n", len(job))
		if len(job) >= 1 {
			jobinfo, _ = job[0].GetJobInfo()
		}
	}
	return jobinfo
}

// getJobInfoByState returns an array of JobInfo objects
// of jobs matching a given job state (or nil)
func getJobInfoByState(state string) []drmaa2.JobInfo {
	jobinfo := drmaa2.CreateJobInfo()
	filter := &jobinfo
	switch state {
	case "r":
		filter.State = drmaa2.Running
	case "q":
		filter.State = drmaa2.Queued
	case "h":
		filter.State = drmaa2.QueuedHeld
	case "s":
		filter.State = drmaa2.Suspended
	case "R":
		filter.State = drmaa2.Requeued
	case "Rh":
		filter.State = drmaa2.RequeuedHeld
	case "d":
		filter.State = drmaa2.Done
	case "f":
		filter.State = drmaa2.Failed
	case "u":
		filter.State = drmaa2.Undetermined
	case "all":
		// no filter, we need all jobs
		filter = nil
	default:
		filter.State = drmaa2.Done
	}

	if job, err := ms.GetAllJobs(filter); err != nil || job == nil {
		if err != nil {
			fmt.Println("Error during GetAllJobs(): %s", err)
		} else {
			log.Println("No job in that state found!")
		}
	} else {
		log.Printf("amount of matching jobs %d\n", len(job))
		if len(job) >= 1 {
			ji := make([]drmaa2.JobInfo, 0, 500)
			for i, _ := range job {
				jinfo, _ := job[i].GetJobInfo()
				ji = append(ji, *jinfo)
			}
			return ji
		}
	}
	return nil
}

func monitoringSessionHandler(w http.ResponseWriter, r *http.Request) {
	// get a job with a specific id
	if jobid := r.FormValue("jobid"); jobid != "" {
		if jobinfo := getJobInfo(jobid); jobinfo != nil {
			encoder := json.NewEncoder(w)
			if err := encoder.Encode(jobinfo); err != nil {
				log.Println("Encoding error: ", err)
			}
		}
		return
	}

	// get all jobs in a certain state
	if state := r.FormValue("state"); state != "" {
		if jobinfo := getJobInfoByState(state); jobinfo != nil {
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

	// get compute nodes of DRM
	if machines := r.FormValue("machines"); machines != "" {
		var filter []string
		if machines == "all" {
			filter = nil
		} else {
			filter = make([]string, 0, 0)
			filter = append(filter, machines)
		}
		if machines, err := ms.GetAllMachines(filter); err == nil {
			encoder := json.NewEncoder(w)
			encoder.Encode(machines)
		}
	}

	// get queues of DRM
	if queues := r.FormValue("queues"); queues != "" {
		var filter []string
		if queues == "all" {
			filter = nil
		} else {
			filter = make([]string, 0, 0)
			filter = append(filter, queues)
		}
		if qs, err := ms.GetAllQueues(filter); err == nil {
			encoder := json.NewEncoder(w)
			if err := encoder.Encode(qs); err != nil {
				fmt.Println("Queue encode error: %s", err)
			}
		}
	}
}

// Reads in JSON for DRMAA2 job template struct.
func jobSubmitHandler(w http.ResponseWriter, r *http.Request) {
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

func main() {
	var err error
	var sm drmaa2.SessionManager
	jsName := "proxy_jsession"
	// Port number where proxy listens
	prt := ":8888"
	f := flag.NewFlagSet("d2proxy", flag.ExitOnError)
	port := f.String("port", "", "Sets tcp/ip port on which proxy is listening. (default :8888)")

	if len(os.Args) > 1 {
		if err := f.Parse(os.Args[1:]); err != nil {
			fmt.Println("Error during parsing: ", err)
			os.Exit(2)
		}
		if *port != "" {
			prt = *port
		}
	}

	if ms, err = sm.OpenMonitoringSession(""); err != nil {
		log.Fatal("Couldn't open DRMAA2 MonitoringSession")
	}

	if js, err = sm.CreateJobSession(jsName, ""); err != nil {
		log.Println("(proxy): Job session proxySession exists already. Reopen it.")
		if js, err = sm.OpenJobSession(jsName); err != nil {
			log.Fatal("(proxy): Couldn't open job session: ", err)
		}
	}

	r := mux.NewRouter()
	r.HandleFunc("/monitoring", monitoringSessionHandler)
	r.HandleFunc("/session", jobSubmitHandler).Methods("POST")
	http.Handle("/", r)
	log.Fatal(http.ListenAndServe(prt, nil))
}

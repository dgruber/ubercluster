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

func jobIdHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	if jobid := vars["jobid"]; jobid != "" {
		if jobinfo := getJobInfo(ms, jobid); jobinfo != nil {
			json.NewEncoder(w).Encode(jobinfo)
		} // TODO not found
	}
}

func monitoringSessionHandler(w http.ResponseWriter, r *http.Request) {
	// get a job with a specific id
	if jobid := r.FormValue("jobid"); jobid != "" {
		if jobinfo := getJobInfo(ms, jobid); jobinfo != nil {
			encoder := json.NewEncoder(w)
			if err := encoder.Encode(jobinfo); err != nil {
				log.Println("Encoding error: ", err)
			}
		}
		return
	}

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

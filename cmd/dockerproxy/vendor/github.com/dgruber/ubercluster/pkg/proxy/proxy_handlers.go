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

package proxy

import (
	"encoding/json"
	"fmt"
	"github.com/dgruber/ubercluster/pkg/persistency"
	"github.com/dgruber/ubercluster/pkg/staging"
	"github.com/dgruber/ubercluster/pkg/types"
	"github.com/gorilla/mux"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

func getDRMAA2JobState(state string) types.JobState {
	switch state {
	case "r":
		return types.Running
	case "q":
		return types.Queued
	case "h":
		return types.QueuedHeld
	case "s":
		return types.Suspended
	case "R":
		return types.Requeued
	case "Rh":
		return types.RequeuedHeld
	case "d":
		return types.Done
	case "f":
		return types.Failed
	case "u":
		return types.Undetermined
	}
	return types.Undetermined
}

// MakeMSessionJobInfosHandler retuns an http handler function which returns
// a JSON encoded collection of DRMAA2 job info object of all jobs available.
func MakeMSessionJobInfosHandler(impl ProxyImplementer, pi persistency.PersistencyImplementer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		filterSet := false
		var filter types.JobInfo
		if state := r.FormValue("state"); state != "all" && state != "" {
			filter.State = getDRMAA2JobState(state)
			log.Printf("filter for state: %s\n", filter.State)
			filterSet = true
		}
		if user := r.FormValue("user"); user != "" {
			filter.JobOwner = user
			log.Printf("filter for user: %s\n", filter.JobOwner)
			filterSet = true
		}
		if jobinfos := impl.GetJobInfosByFilter(filterSet, filter); jobinfos != nil {
			encoder := json.NewEncoder(w)
			if err := encoder.Encode(jobinfos); err != nil {
				fmt.Printf("Encoding error: %s\n", err)
			} else {
				log.Printf("Encoded: %s\n", jobinfos)
			}
		}
	}
}

// MakeMSessionJobInfoHandler returns an http handler function which returns
// a JSON encoded DRMAA2 Job Info object.
func MakeMSessionJobInfoHandler(impl ProxyImplementer, pi persistency.PersistencyImplementer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		if jobid := vars["jobid"]; jobid != "" {
			if jobinfo := impl.GetJobInfo(jobid); jobinfo != nil {
				json.NewEncoder(w).Encode(*jobinfo)
			} else {
				log.Printf("JobInfo not found for job %s\n", jobinfo)
			}
		}
	}
}

// MakeMachinesHandler returns an http handler function which returns
// a JSON encoded collection of all machines availale in the DRM.
func MakeMachinesHandler(impl ProxyImplementer, pi persistency.PersistencyImplementer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if machines, err := impl.GetAllMachines(nil); err == nil {
			json.NewEncoder(w).Encode(machines)
		} else {
			log.Printf("Error in GetAllMachines: %s\n", err)
		}
	}
}

// MakeMachineHandler retuns an http handler function which returns
// a JSON encoded DRMAA2 machine object if the machine is part of the
// DRM system.
func MakeMachineHandler(impl ProxyImplementer, pi persistency.PersistencyImplementer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		name := vars["name"]
		if machines, err := impl.GetAllMachines([]string{name}); err == nil {
			json.NewEncoder(w).Encode(machines)
		} else {
			log.Printf("Error in GetAllMachines: %s\n", err)
		}
	}
}

// MakeQueuesHandler retuns an http handler function which returns
// all available queues in the DRM system JSON encoded.
func MakeQueuesHandler(impl ProxyImplementer, pi persistency.PersistencyImplementer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if queues, err := impl.GetAllQueues(nil); err == nil {
			json.NewEncoder(w).Encode(queues)
		} else {
			log.Printf("Error in GetAllQueues: %s\n", err)
		}
	}
}

// MakeQueueHandler returns an http handler function which returns
// the requested queue if it is available on the system JSON encoded.
func MakeQueueHandler(impl ProxyImplementer, pi persistency.PersistencyImplementer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		name := vars["name"]
		if queues, err := impl.GetAllQueues([]string{name}); err == nil {
			json.NewEncoder(w).Encode(queues)
		} else {
			log.Printf("Error in GetAllQueues: %s\n", err)
		}
	}
}

// MakeMSessionDRMSVersionHandler returns an http handler function which
// returns all available DRMAA2 job categories as JSON encoded string.
func MakeJSessionCategoriesHandler(impl ProxyImplementer, pi persistency.PersistencyImplementer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if categories, err := impl.GetAllCategories(); err == nil {
			json.NewEncoder(w).Encode(categories)
		} else {
			log.Printf("Error in GetAllCategories: %s\n", err)
		}
	}
}

// MakeJSessionCategroyHandler returns an http handler function which
// returns a requested job category when it is available.
func MakeJSessionCategoryHandler(impl ProxyImplementer, pi persistency.PersistencyImplementer) http.HandlerFunc {
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
			log.Printf("Error in GetJobCategories: %s\n", err)
		}
	}
}

// MakeMSessionDRMSNameHandler returns an http handler function which
// returns the DRMS name encoded by the ProxyImplementer as JSON string.
func MakeMSessionDRMSNameHandler(impl ProxyImplementer, pi persistency.PersistencyImplementer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(impl.DRMSName())
	}
}

// MakeMSessionDRMSVersionHandler returns an http handler function which
// returns the DRMS name encoded by the ProxyImplementer as JSON string.
func MakeMSessionDRMSVersionHandler(impl ProxyImplementer, pi persistency.PersistencyImplementer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(impl.DRMSVersion())
	}
}

// MakeMSessionDRMSLoadHandler returns an http handler function which
// returns the DRMS encoded load by the ProxyImplementer as JSON string.
func MakeMSessionDRMSLoadHandler(impl ProxyImplementer, pi persistency.PersistencyImplementer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(impl.DRMSLoad())
	}
}

// RunJobResult is the JSON answer when a job could successully
// started in the cluster.
type RunJobResult struct {
	JobId string `json:"jobid"`
}

// MakeJSessionSubmitHandler returns an http handler function which
// reads in a DRMAA2 job template struct (in JSON) in the body of the
// http request. In case of success the job is submitted in the cluster
// using the RunJob function implemented by the proxy.
// TODO In case a ProxyImplementer is given as a parameter the job template
// is made persistent.
func MakeJSessionSubmitHandler(impl ProxyImplementer, pi persistency.PersistencyImplementer) http.HandlerFunc {
	var workingDir string
	if wd, wdErr := os.Getwd(); wdErr == nil {
		log.Println("(proxy) adapt cwd to ", wd, "uploads")
		workingDir = wd + "/uploads"
	} else {
		fmt.Println("Can't set working directory for the jobs.")
		os.Exit(2)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if body, err := ioutil.ReadAll(r.Body); err != nil {
			log.Printf("(proxy) %s\n", err)
		} else {
			var jt types.JobTemplate
			if uerr := json.Unmarshal(body, &jt); uerr != nil {
				log.Println("(proxy) Unmarshall error")
				http.Error(w, uerr.Error(), http.StatusInternalServerError)
			} else {
				log.Printf("(proxy) Got JobTemplate: %v", jt)
				log.Printf("(proxy) Set working dir for job %s\n", workingDir)
				jt.WorkingDirectory = workingDir
				// required when file is in staging area but not for general path
				// jt.RemoteCommand = workingDir + "/" + jt.RemoteCommand
				log.Println("(proxy) Submit now job")
				// Submit job in compute cluster
				if jobid, joberr := impl.RunJob(jt); joberr != nil {
					log.Printf("(proxy) Error during job submission: %s\n", joberr)
					http.Error(w, joberr.Error(), http.StatusInternalServerError)
				} else {
					log.Printf("(proxy) Job successfully submitted: %s\n", jobid)

					// make job submission persistent on proxy
					if pi != nil {
						if err := pi.SaveJobTemplate(jobid, jt); err != nil {
							log.Printf("(proxy) Error during making Job Template persistent: %s\n", err)
						} else {
							log.Printf("(proxy) Job template for job %s successfully made persistent.\n", jobid)
						}
					}

					var result RunJobResult
					result.JobId = jobid
					json.NewEncoder(w).Encode(result)
				}
			}
		}
	}
}

// MakeRunLocalHandler spawns a process on the same host as proxy.
func MakeRunLocalHandler(impl ProxyImplementer, pi persistency.PersistencyImplementer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if body, err := ioutil.ReadAll(r.Body); err != nil {
			log.Printf("(proxy) %s\n", err)
		} else {
			var rlr types.RunLocalRequest
			if uerr := json.Unmarshal(body, &rlr); uerr != nil {
				log.Println("(proxy) Unmarshall error")
				http.Error(w, uerr.Error(), http.StatusInternalServerError)
			}
			cli := []string{"-c", rlr.Command + " " + rlr.Arg}

			cmd := exec.Command("/bin/sh", cli...)
			cmd.Stdout = os.Stdout
			cmd.Stdin = os.Stdin
			cmd.Stderr = os.Stderr

			log.Printf("Start command: %s %v\n", cmd.Path, cmd.Args)
			if errStart := cmd.Start(); errStart != nil {
				log.Printf("(proxy) Error during starting command %s %s: %s\n", rlr.Command, rlr.Arg, errStart.Error())
				json.NewEncoder(w).Encode(fmt.Sprintf("Failed starting command: %s", errStart.Error()))
			} else {
				json.NewEncoder(w).Encode(fmt.Sprintf("Started command with PID %d", cmd.Process.Pid))
			}
		}
	}
}

func MakeUCFileUploadHandler(impl ProxyImplementer, pi persistency.PersistencyImplementer) http.HandlerFunc {
	stagingDir := "uploads"

	if err := staging.CheckUploadFilesystem(stagingDir); err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
	return func(w http.ResponseWriter, r *http.Request) {
		// currently limited to 1GB until tested
		const maxSize = 1024 * 1024 * 1024
		if r.ContentLength > maxSize {
			log.Println("File content too large", r.ContentLength)
			http.Error(w, "File too large", http.StatusExpectationFailed)
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, maxSize)
		err := r.ParseMultipartForm(1024 * 1024 * 128)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusExpectationFailed)
			return
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			log.Println("Error: ", err)
			panic(err)
		}
		if strings.ContainsAny(header.Filename, "/\\!") || strings.Contains(header.Filename, "..") {
			log.Println("File name contains invalid characters..", header.Filename)
			http.Error(w, "File name contains invalid chars", http.StatusExpectationFailed)
			return
		}
		dst, err := os.Create(stagingDir + "/" + header.Filename)
		defer dst.Close()
		if err != nil {
			panic(err)
		}

		if written, err := io.Copy(dst, io.LimitReader(file, maxSize)); err != nil {
			log.Println("Error: ", err)
			panic(err)
		} else {
			if written == maxSize {
				log.Println("File upload too large.")
				http.Error(w, "File too large", http.StatusExpectationFailed)
				return
			}
			log.Println("File saved successfully")
		}
		log.Println(r.FormValue("permission"))
		if r.FormValue("permission") == "exec" {
			// make the file an executable
			if err := dst.Chmod(0700); err != nil {
				log.Println(err)
			} else {
				log.Println("Made file executable.")
			}
		}

		json.NewEncoder(w).Encode("File upload successful")
	}
}

// MakeJSessionJobManipulationHandler returns an http handler function which
// calls the JobOperation function defined by an ProxyImplementer.
func MakeJSessionJobManipulationHandler(impl ProxyImplementer, pi persistency.PersistencyImplementer) http.HandlerFunc {
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

// MakeListFilesHandler creates an http handler function which returns
// a list of all files in the staging area over http.
func MakeListFilesHandler(impl ProxyImplementer, pi persistency.PersistencyImplementer) http.HandlerFunc {
	// TODO disallow based on config / startup params ...
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("(ListFilesHandler) called")
		// job session name must be the one created by d2proxy
		// json.NewEncoder(w).Encode("invalid job session name")
		if dir, err := os.Open("uploads"); err != nil {
			fmt.Println("Can't open staging directory. ", err)
			os.Exit(1)
		} else {
			if fi, err := dir.Stat(); err != nil {
				log.Println("Can't stat file staging directory: ", err)
				http.Error(w, "Error in staging area", http.StatusForbidden)
				return
			} else {
				if fi.IsDir() == false {
					log.Println("File staging directory not found: ", err)
					http.Error(w, "Error in staging area", http.StatusForbidden)
					return
				} else {
					if fis, err := dir.Readdir(-1); err == nil {
						log.Println("Files in staging directory found ")
						fileinfos := make([]types.FileInfo, 0, len(fis))
						for _, fi := range fis {
							if fi.IsDir() == false {
								var info types.FileInfo
								info.Filename = fi.Name()
								info.Bytes = fi.Size()
								if fi.Mode() == 0700 {
									info.Executable = true
								} else {
									info.Executable = false
								}
								fileinfos = append(fileinfos, info)
								log.Println("added: ", info.Filename)
							}
						}
						fmt.Println(fileinfos)
						json.NewEncoder(w).Encode(fileinfos)
					} else {
						log.Println("Error during dir.Readdir: ", err)
						http.Error(w, "Error in staging area", http.StatusForbidden)
					}
				}
			}
		}
	}
}

// MakeDownloadFilesHandler returns an http handler function which
// serves a file requested with the *name* http request.
func MakeDownloadFilesHandler(impl ProxyImplementer, pi persistency.PersistencyImplementer) http.HandlerFunc {
	// TODO uploads directory should be defined by the proxy implementer
	// or depend from the job session.
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		if filename := vars["name"]; filename != "" {
			log.Println("Serving file: ./uploads/", filename)
			http.ServeFile(w, r, "./uploads/"+filename)
		} else {
			http.Error(w, "No filename given.", http.StatusForbidden)
		}
	}
}

// MakeSessionListHandler implements an http handler which serves
// a list of (DRMAA2) job sessions available on this proxy.
func MakeSessionListHandler(impl ProxyImplementer, pi persistency.PersistencyImplementer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if sessions, err := impl.GetAllSessions(nil); err == nil {
			json.NewEncoder(w).Encode(sessions)
		} else {
			log.Println("Error in GetAllSessions: ", err)
		}
	}
}

func AutenticationErrorHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Authentication error")
	http.NotFound(w, r)
}

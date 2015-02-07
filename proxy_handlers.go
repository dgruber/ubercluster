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
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
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
			log.Println("(proxy)", err)
		} else {
			var jt JobTemplate
			if uerr := json.Unmarshal(body, &jt); uerr != nil {
				log.Println("(proxy) Unmarshall error")
				http.Error(w, uerr.Error(), http.StatusInternalServerError)
			} else {
				log.Println("(proxy) Set working dir for job ", workingDir)
				jt.WorkingDirectory = workingDir
				jt.RemoteCommand = workingDir + "/" + jt.RemoteCommand
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

func MakeUCFileUploadHandler(impl ProxyImplementer) http.HandlerFunc {
	if err := checkUploadFilesystem("uploads"); err != nil {
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
			panic(err)
		}
		if strings.ContainsAny(header.Filename, "/\\!") || strings.Contains(header.Filename, "..") {
			log.Println("File name contains invalid characters..", header.Filename)
			http.Error(w, "File name contains invalid chars", http.StatusExpectationFailed)
			return
		}
		dst, err := os.Create("uploads/" + header.Filename)
		defer dst.Close()
		if err != nil {
			panic(err)
		}

		if written, err := io.Copy(dst, io.LimitReader(file, maxSize)); err != nil {
			panic(err)
		} else {
			if written == maxSize {
				log.Println("File upload too large.")
				http.Error(w, "File too large", http.StatusExpectationFailed)
				return
			}
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

// MakeUCListFilesHandler creates an http handler which serves
// a list of files in the staging area of the proxy
func MakeListFilesHandler(impl ProxyImplementer) http.HandlerFunc {
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
						fileinfos := make([]FileInfo, 0, len(fis))
						for _, fi := range fis {
							if fi.IsDir() == false {
								var info FileInfo
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

func MakeDownloadFilesHandler(impl ProxyImplementer) http.HandlerFunc {
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

func AutenticationErrorHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Authentication error")
	http.NotFound(w, r)
}

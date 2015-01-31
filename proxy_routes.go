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
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
)

// A ProxyImplementer implements functions required to interface
// a ubercluster proxy. Those functions are called in the standard
// http request handlers.
type ProxyImplementer interface {
	GetJobInfosByFilter(filtered bool, filter JobInfo) []JobInfo
	GetJobInfo(jobid string) *JobInfo
	GetAllMachines(machines []string) ([]Machine, error)
	GetAllQueues(queues []string) ([]Queue, error)
	GetAllCategories() ([]string, error)
	DRMSVersion() string
	DRMSName() string
	RunJob(template JobTemplate) (string, error)
	JobOperation(jobsessionname, operation, jobid string) (string, error)
	DRMSLoad() float64
}

type Routes []Route

// A Route maps a name, method, and a http request pattern
// to a http handler function which is executed.
type Route struct {
	Name            string
	Method          string
	Pattern         string
	MakeHandlerFunc func(ProxyImplementer) http.HandlerFunc
}

var routes = Routes{
	Route{
		"JobSubmit", "POST", "/v1/jsession/default/run", MakeJSessionSubmitHandler,
	},
	// Operations are: suspend resume delete (hold / release)
	Route{
		"JobManipulation", "POST", "/v1/jsession/{jsname}/{operation}/{jobid}", MakeJSessionJobManipulationHandler,
	},
	Route{
		"JobCategories", "GET", "/v1/jsession/{jsname}/jobcategories", MakeJSessionCategoriesHandler,
	},
	Route{
		"JobCategory", "GET", "/v1/jsession/{jsname}/jobcategory/{category}", MakeJSessionCategoryHandler,
	},
	Route{
		"msessionJobInfos", "GET", "/v1/msession/jobinfos", MakeMSessionJobInfosHandler,
	},
	Route{
		"jobid", "GET", "/v1/msession/jobinfo/{jobid}", MakeMSessionJobInfoHandler,
	},
	Route{
		"msessionMachines", "GET", "/v1/msession/machines", MakeMachinesHandler,
	},
	Route{
		"msessionMachine", "GET", "/v1/msession/machine/{name}", MakeMachineHandler,
	},
	Route{
		"msessionQueues", "GET", "/v1/msession/queues", MakeQueuesHandler,
	},
	Route{
		"msessionQueue", "GET", "/v1/msession/queue/{name}", MakeQueueHandler,
	},
	Route{
		"msessionDRMSName", "GET", "/v1/msession/drmsname", MakeMSessionDRMSNameHandler,
	},
	Route{
		"msessionDRMSVersion", "GET", "/v1/msession/drmsversion", MakeMSessionDRMSVersionHandler,
	},
	Route{
		"msessionDRMSload", "GET", "/v1/msession/drmsload", MakeMSessionDRMSLoadHandler,
	},
	Route{
		"uberclusterFileUpload", "POST", "/v1/ubercluster/fileupload", MakeUCFileUploadHandler,
	},
	Route{
		"uberclusterFileList", "GET", "/v1/jsession/staging/files", MakeListFilesHandler,
	},
}

// Simple security through a shared secret
func MakeFixedSecretHandler(secret string, f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if secret != "" {
			otpFromClient := r.FormValue("otp")
			if otpFromClient == "" {
				otpFromClient = r.PostFormValue("otp")
			}
			// log.Println(*r)
			// log.Printf("OTP is set to %s and request is %s\n", secret, otpFromClient)
			// check otp
			if otpFromClient == secret {
				f(w, r)
			} else {
				log.Println("Unauthorized access by ", r.RemoteAddr)
				// slow down
				http.Error(w, "authorization failed", http.StatusUnauthorized)
				return
			}
		} else {
			// don't check one time password
			f(w, r)
		}
	}
}

// TODO Make Yubikey handler

// NewProxyRouter creates a mux router for matching
// http requests to handlers.
func NewProxyRouter(impl ProxyImplementer, otp string) *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	if otp == "" {
		for _, route := range routes {
			router.
				Methods(route.Method).
				Path(route.Pattern).
				Name(route.Name).
				Handler(route.MakeHandlerFunc(impl))
		}
	} else {
		for _, route := range routes {
			router.
				Methods(route.Method).
				Path(route.Pattern).
				Name(route.Name).
				Handler(MakeFixedSecretHandler(otp, route.MakeHandlerFunc(impl)))
		}
	}
	return router
}

func ProxyListenAndServe(addr, certFile, keyFile, otp string, impl ProxyImplementer) {
	if certFile != "" && keyFile != "" {
		if err := http.ListenAndServeTLS(addr, certFile, keyFile, NewProxyRouter(impl, otp)); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	} else {
		if err := http.ListenAndServe(addr, NewProxyRouter(impl, otp)); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
}

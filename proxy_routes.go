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
	"github.com/GeertJohan/yubigo"
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
	GetAllSessions(session []string) ([]string, error)
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
		"JobSubmit", "POST", "/v1/jsession/{jsname}/run", MakeJSessionSubmitHandler,
	},
	// Operations are: suspend resume delete (hold / release)
	Route{
		"JobManipulation", "POST", "/v1/jsession/{jsname}/{operation:suspend|resume|terminate}/{jobid}", MakeJSessionJobManipulationHandler,
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
		"uberclusterFileUpload", "POST", "/v1/jsession/{jsname}/staging/upload", MakeUCFileUploadHandler,
	},
	Route{
		"jsessionSessions", "GET", "/v1/jsessions", MakeSessionListHandler,
	},
	Route{
		"jsessionFiles", "GET", "/v1/jsession/{jsname}/staging/files", MakeListFilesHandler,
	},
	Route{
		"jsessionFileDownload", "GET", "/v1/jsession/{jsname}/staging/file/{name}", MakeDownloadFilesHandler,
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

// global authenticfication instance which is used by all
// http handlers - a bit hacky
var yubiAuth *yubigo.YubiAuth

// MaeYubikeyHandler creates an http handler which is protected by an
// yubkikey one-time-password verification. The OTP needs to be given
// by either a form value ("otp") or a POST form value ("otp").
func MakeYubikeyHandler(id, key string, f http.HandlerFunc) http.HandlerFunc {
	var errAuth error
	if yubiAuth == nil {
		if yubiAuth, errAuth = yubigo.NewYubiAuth(id, key); errAuth != nil {
			fmt.Println("Error during yubiAuth instance creation: ", errAuth)
			os.Exit(1)
		} else {
			log.Println("Succesfully created yubiAuth instance.")
		}
	}
	return func(w http.ResponseWriter, r *http.Request) {
		otpFromClient := r.FormValue("otp")
		if otpFromClient == "" {
			otpFromClient = r.PostFormValue("otp")
		}
		// verify OTP
		if result, ok, err := yubiAuth.Verify(otpFromClient); ok {
			// successfully verified the one time password
			f(w, r)
		} else {
			if err != nil {
				// something really bad! probably best to abort
				fmt.Println("Verification of yubikey failed with error: ", err)
				os.Exit(1)
			}
			log.Println("Verification of yubikey OTP failed: ", result)
			log.Println("Unauthorized access by ", r.RemoteAddr)
			http.Error(w, "authorization failed", http.StatusUnauthorized)
		}
	}
}

// NewProxyRouter creates a mux router for matching
// http requests to handlers.
func NewProxyRouter(impl ProxyImplementer, sc SecConfig) *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	if sc.OTP == "" {
		for _, route := range routes {
			router.
				Methods(route.Method).
				Path(route.Pattern).
				Name(route.Name).
				Handler(route.MakeHandlerFunc(impl))
		}
	} else if sc.OTP == "yubikey" {
		// add yubikey one-time-password verifcation for each call
		if sc.YubiID == "" || sc.YubiSecret == "" {
			fmt.Println("yubikey is configured but ID or Secret not set!")
			os.Exit(1)
		}
		for _, route := range routes {
			router.
				Methods(route.Method).
				Path(route.Pattern).
				Name(route.Name).
				Handler(MakeYubikeyHandler(sc.YubiID, sc.YubiSecret, route.MakeHandlerFunc(impl)))
		}
	} else {
		// fixed key
		for _, route := range routes {
			router.
				Methods(route.Method).
				Path(route.Pattern).
				Name(route.Name).
				Handler(MakeFixedSecretHandler(sc.OTP, route.MakeHandlerFunc(impl)))
		}
	}
	return router
}

// Security related configuration settings for the ubercluster Proxy
type SecConfig struct {
	OTP        string // secret key or "yubikey"
	YubiID     string // ID of yubiservice in case of yubikey https://upgrade.yubico.com/getapikey/
	YubiSecret string // Secret of yubiservice in case of yubikey https://upgrade.yubico.com/getapikey/
}

func ProxyListenAndServe(addr, certFile, keyFile string, sc SecConfig, impl ProxyImplementer) {
	if certFile != "" && keyFile != "" {
		if err := http.ListenAndServeTLS(addr, certFile, keyFile, NewProxyRouter(impl, sc)); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	} else {
		if err := http.ListenAndServe(addr, NewProxyRouter(impl, sc)); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
}

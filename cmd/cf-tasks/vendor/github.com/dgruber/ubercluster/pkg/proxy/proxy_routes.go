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
	"fmt"
	"github.com/GeertJohan/yubigo"
	"github.com/dgruber/ubercluster/pkg/persistency"
	"github.com/dgruber/ubercluster/pkg/types"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
)

// ProxyImplementer interface specified functions required to interface
// a ubercluster proxy. Those functions are called in the standard
// http request handlers.
type ProxyImplementer interface {
	GetJobInfosByFilter(filtered bool, filter types.JobInfo) []types.JobInfo
	GetJobInfo(jobid string) *types.JobInfo
	GetAllMachines(machines []string) ([]types.Machine, error)
	GetAllQueues(queues []string) ([]types.Queue, error)
	GetAllCategories() ([]string, error)
	GetAllSessions(session []string) ([]string, error)
	DRMSVersion() string
	DRMSName() string
	RunJob(template types.JobTemplate) (string, error)
	JobOperation(jobsessionname, operation, jobid string) (string, error)
	DRMSLoad() float64
}

type Routes []Route

// Route is a structure which maps a name, method, and a http request pattern
// to a http handler function which is executed.
type Route struct {
	Name            string
	Method          string
	Pattern         string
	MakeHandlerFunc func(ProxyImplementer, persistency.PersistencyImplementer) http.HandlerFunc
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

// MakeFixedSecretHandler protects an http handler by a simple shared secret
// given a request or a post form value. Note that without TLS it is not
// encrypted through the network and it can be sniffed.
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

// MakeYubikeyHandler creates an http handler which is protected by an
// yubkikey one-time-password verification. The OTP needs to be given
// by either a form value ("otp") or a POST form value ("otp").
func MakeYubikeyHandler(id, key string, allowedIDs []string, f http.HandlerFunc) http.HandlerFunc {
	var errAuth error
	if yubiAuth == nil {
		if yubiAuth, errAuth = yubigo.NewYubiAuth(id, key); errAuth != nil {
			fmt.Println("Error during yubiAuth instance creation: ", errAuth)
			os.Exit(1)
		} else {
			log.Println("Succesfully created yubiAuth instance.")
		}
	}
	if allowedIDs == nil {
		fmt.Println("No allowed yubikey IDs given. Aborting.")
		os.Exit(1)
	}
	return func(w http.ResponseWriter, r *http.Request) {
		otpFromClient := r.FormValue("otp")
		if otpFromClient == "" {
			otpFromClient = r.PostFormValue("otp")
		}
		// check if ID of OTP is allowed (first 12 chars)
		if len(otpFromClient) != 44 {
			log.Println("Unauthorized access by ", r.RemoteAddr)
			log.Printf("Length of OTP does not match 44: %d", len(otpFromClient))
			http.Error(w, "authorization failed", http.StatusUnauthorized)
		}

		id := otpFromClient[0:12]
		found := false
		for _, v := range allowedIDs {
			log.Printf("Compare %s with %s\n", v, id)
			if v == id {
				found = true
				break
			}
		}
		if found == false {
			log.Println("Unauthorized access by ", r.RemoteAddr)
			log.Printf("ID %s not in list of allowed IDs", id)
			http.Error(w, "authorization failed", http.StatusUnauthorized)
			return
		}

		// verify OTP
		if result, ok, err := yubiAuth.Verify(otpFromClient); ok {
			// successfully verified the one time password
			f(w, r)
		} else {
			if err != nil {
				// something really bad! probably best to abort
				fmt.Println("Verification of yubikey failed with error: ", err)
				log.Println("Unauthorized access by ", r.RemoteAddr)
				http.Error(w, "authorization failed", http.StatusUnauthorized)
			} else {
				log.Println("Verification of yubikey OTP failed: ", result)
				log.Println("Unauthorized access by ", r.RemoteAddr)
				http.Error(w, "authorization failed", http.StatusUnauthorized)
			}
		}
	}
}

// NewProxyRouter creates a mux router for matching http requests to handlers.
// When security is configured it adds neccessary closures around the functions.
func NewProxyRouter(impl ProxyImplementer, sc SecConfig, pi persistency.PersistencyImplementer) *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	if sc.OTP == "" {
		for _, route := range routes {
			router.
				Methods(route.Method).
				Path(route.Pattern).
				Name(route.Name).
				Handler(route.MakeHandlerFunc(impl, pi))
		}
	} else if sc.OTP == "yubikey" {
		// add yubikey one-time-password verifcation for each call
		if sc.YubiID == "" || sc.YubiSecret == "" {
			fmt.Println("yubikey is configured but ID or Secret not set!")
			os.Exit(1)
		}
		if sc.YubiAllowedIDs == nil {
			fmt.Println("yubikey is configured but no allowed keys set (first 12 chars of your OTP)!")
			os.Exit(1)
		}
		for _, route := range routes {
			router.
				Methods(route.Method).
				Path(route.Pattern).
				Name(route.Name).
				Handler(MakeYubikeyHandler(sc.YubiID, sc.YubiSecret, sc.YubiAllowedIDs, route.MakeHandlerFunc(impl, pi)))
		}
	} else {
		// fixed key
		for _, route := range routes {
			router.
				Methods(route.Method).
				Path(route.Pattern).
				Name(route.Name).
				Handler(MakeFixedSecretHandler(sc.OTP, route.MakeHandlerFunc(impl, pi)))
		}
	}
	return router
}

// SecConfig stores security related configuration settings for the ubercluster Proxy
type SecConfig struct {
	OTP            string   // secret key or "yubikey"
	YubiID         string   // ID of yubiservice in case of yubikey https://upgrade.yubico.com/getapikey/
	YubiSecret     string   // Secret of yubiservice in case of yubikey https://upgrade.yubico.com/getapikey/
	YubiAllowedIDs []string // IDs of yubkeys which are allowed
}

// ProxyListenAndServe starts an http proxy for a cluster which is accessed by functions
// specified in the ProxyImplementer interface. If a certification and key file is given
// as parameter then it starts an TLS secured http proxy. The port is specified by addr
// in the form which is used by http.ListenAndServe.
func ProxyListenAndServe(addr, certFile, keyFile string, sc SecConfig, pi persistency.PersistencyImplementer, impl ProxyImplementer) {
	if certFile != "" && keyFile != "" {
		if err := http.ListenAndServeTLS(addr, certFile, keyFile, NewProxyRouter(impl, sc, pi)); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	} else {
		if err := http.ListenAndServe(addr, NewProxyRouter(impl, sc, pi)); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
}

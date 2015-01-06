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
	"net/http"
	"os"
)

// A ProxyImplementerer implements functions required to interface
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
}

// NewProxyRouter creates a mux router for matching
// http requests to handlers.
func NewProxyRouter(impl ProxyImplementer) *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	for _, route := range routes {
		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(route.MakeHandlerFunc(impl))
	}
	return router
}

func ProxyListenAndServe(addr, certFile, keyFile string, impl ProxyImplementer) {
	if certFile != "" && keyFile != "" {
		if err := http.ListenAndServeTLS(addr, certFile, keyFile, NewProxyRouter(impl)); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	} else {
		if err := http.ListenAndServe(addr, NewProxyRouter(impl)); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
}

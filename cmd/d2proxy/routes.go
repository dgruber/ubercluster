package main

import (
	"github.com/gorilla/mux"
	"net/http"
)

type Routes []Route

// A Route maps a name, method, and a http request pattern
// to a http handler function which is executed.
type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

var routes = Routes{
	Route{
		"MonitoringSession", "GET", "/v1/monitoring", monitoringSessionHandler,
	},
	Route{
		"JobSubmit", "POST", "/v1/jsession/default/run", jobSubmitHandler,
	},
	Route{
		"jobid", "GET", "/v1/msession/jobinfo/{jobid}", msessionJobInfoHandler,
	},
	Route{
		"msessionMachines", "GET", "/v1/msession/machines", msessionMachinesHandler,
	},
	Route{
		"msessionMachine", "GET", "/v1/msession/machine/{name}", msessionMachineHandler,
	},
	Route{
		"msessionQueues", "GET", "/v1/msession/queues", msessionQueuesHandler,
	},
	Route{
		"msessionMachine", "GET", "/v1/msession/queue/{name}", msessionQueueHandler,
	},
	Route{
		"msessionDRMSName", "GET", "/v1/msession/drmsname", msessionDRMSNameHandler,
	},
	Route{
		"msessionDRMSVersion", "GET", "/v1/msession/drmsversion", msessionDRMSVersionHandler,
	},
}

// NewProxyRouter creates a mux router for matching
// http requests to handlers.
func NewProxyRouter() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	for _, route := range routes {
		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(route.HandlerFunc)
	}
	return router
}

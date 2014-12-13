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
		"MonitoringSession",
		"GET",
		"/monitoring",
		monitoringSessionHandler,
	},
	Route{
		"JobInfo",
		"POST",
		"/session",
		jobSubmitHandler,
	},
	Route{
		"jonid",
		"GET",
		"/jobid/{jobid}",
		jobIdHandler,
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

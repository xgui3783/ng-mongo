package main

import (
	"net/http"
	"regexp"
)

type route struct {
	pattern *regexp.Regexp
	handler http.Handler
}

type regexRouter struct {
	routes []*route
}

func (handler *regexRouter) AddRegExpRoute(s string, h http.Handler) {
	pattern, err := regexp.Compile(s)
	if err != nil {
		panic(err)
	}
	handler.routes = append(handler.routes, &route{pattern, h})
}

func (handler *regexRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, route := range handler.routes {
		if route.pattern.MatchString(r.URL.Path) {
			route.handler.ServeHTTP(w, r)
			return
		}
	}
	http.NotFound(w, r)
}

package api

import (
	"net/http"

	"github.com/jkoelndorfer/wedding-website/rsvp/log"
)

var defaultHeaders = map[string]string{
	"Content-Type": "application/json",
}

var logger = log.Logger()

func RequestHandler(w http.ResponseWriter, r *http.Request) {
	for k, v := range defaultHeaders {
		w.Header().Add(k, v)
	}

	var handler HandlerFunction
	var statusCode int
	var body string
	if r.RequestURI == "/lookup" {
    handler = 
		// look up an invitation and return information about it
	} else if r.RequestURI == "/respond" {
		// allow a user to reply to their invitation
	} else {
		statusCode = http.StatusNotFound
	}
	logger.Printf("got request for URI '%s'", r.RequestURI)

	w.WriteHeader(statusCode)
	w.Write([]byte(body))
}

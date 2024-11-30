package api

import (
	"encoding/json"
	"net/http"

	"github.com/jkoelndorfer/wedding-website/rsvp/config"
	"github.com/jkoelndorfer/wedding-website/rsvp/db"
	"github.com/jkoelndorfer/wedding-website/rsvp/log"
)

var defaultHeaders = map[string]string{
	"Content-Type": "application/json",
}

var logger = log.Logger()

func RequestHandler(cfg config.RSVPConfig, invitationRepository db.InvitationRepository) func(http.ResponseWriter, *http.Request) {
	authenticationService := NewAuthenticationService(cfg)

	return func(w http.ResponseWriter, r *http.Request) {
		for k, v := range defaultHeaders {
			w.Header().Add(k, v)
		}

		var passedDB db.InvitationRepository
		var handler HandlerFunction
		var statusCode int
		var response APIResponse
		var privilegedEndpoint bool

		passedDB = invitationRepository

		if r.RequestURI == "/lookup" {
			// Look up an invitation and return information about it.
			handler = Lookup
			privilegedEndpoint = false
		} else if r.RequestURI == "/respond" {
			// Allow a user to reply to their invitation.
			handler = Respond
			privilegedEndpoint = false
		} else if r.RequestURI == "/load" {
			// Allow invitation information to be uploaded.
			//
			// This requires special authentication.
			handler = Load
			privilegedEndpoint = true
		} else {
			handler = Default
			privilegedEndpoint = false
			passedDB = nil
		}
		logger.Printf("got request for URI '%s'", r.RequestURI)

		requestAPIClientToken := r.Header.Get("X-API-Client-Token")
		requestSecretKey := r.Header.Get("X-API-Secret-Key")
		if !authenticationService.APIClientTokenValid(requestAPIClientToken) {
			handler = Unauthorized
			passedDB = nil
		}

		if privilegedEndpoint && !authenticationService.APISecretKeyValid(requestSecretKey) {
			handler = Unauthorized
			passedDB = nil
		}

		statusCode, response = handler(passedDB, r)
		response_json, err := json.Marshal(response)
		if err != nil {
			logger.Printf("error marshalling JSON: %s\n", err)
		}

		w.WriteHeader(statusCode)
		w.Write(response_json)
	}
}

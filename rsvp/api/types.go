package api

import (
	"net/http"

	"github.com/jkoelndorfer/wedding-website/rsvp/db"
)

// Signature for a standard RSVP handler function.
type HandlerFunction func(db.InvitationRepository, *http.Request) (statusCode int, response APIResponse)

// Data structure representing an error from the API.
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Data structure representing a standard API response.
type APIResponse struct {
	Error    *APIError   `json:"error"`
	Response interface{} `json:"response"`
}

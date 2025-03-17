package api

import (
	"net/http"

	"github.com/jkoelndorfer/wedding-website/rsvp/db"
)

func Default(db db.InviteRepository, r *http.Request) (int, APIResponse) {
	return http.StatusNotFound, APIResponse{
		Error:    &APIError{Code: "not_found", Message: "the requested endpoint was not found"},
		Response: nil,
	}
}

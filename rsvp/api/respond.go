package api

import (
	"net/http"

	"github.com/jkoelndorfer/wedding-website/rsvp/db"
)

func Respond(db db.InvitationRepository, r *http.Request) (int, APIResponse) {
	return http.StatusInternalServerError, APIResponse{
		Error: &APIError{Code: "not_implemented", Message: "this endpoint is not implemented"},
	}
}

package api

import (
	"net/http"

	"github.com/jkoelndorfer/wedding-website/rsvp/db"
)

func Unauthorized(db db.InviteRepository, r *http.Request) (int, APIResponse) {
	return http.StatusUnauthorized, APIResponse{
		Error: &APIError{Code: "unauthorized", Message: "invalid authorization for this endpoint"},
	}
}

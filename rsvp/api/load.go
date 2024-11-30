package api

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/jkoelndorfer/wedding-website/rsvp/db"
	"github.com/jkoelndorfer/wedding-website/rsvp/model"
)

func Load(db db.InvitationRepository, r *http.Request) (int, APIResponse) {
	requestBody, err := io.ReadAll(r.Body)
	if err != nil {
		return http.StatusInternalServerError, APIResponse{
			Error: APIError{Code: "internal_server_error", Message: "error processing request"},
		}
	}

	loadRequest := &LoadRequest{Invitations: make([]model.Invitation, 128)}
	err = json.Unmarshal(requestBody, &loadRequest)
	if err != nil {
		return http.StatusBadRequest, APIResponse{
			Error: APIError{Code: "bad_request", Message: "bad load request"},
		}
	}

	db.Load(loadRequest.Invitations)

	return http.StatusInternalServerError, APIResponse{
		Error: APIError{Code: "not_implemented", Message: "this endpoint is not implemented"},
	}
}

package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/jkoelndorfer/wedding-website/rsvp/db"
	"github.com/jkoelndorfer/wedding-website/rsvp/model"
)

// Data structure representing a request to load invite data.
type LoadRequest struct {
	// Invites to be loaded.
	Invites []model.Invite `json:"invites"`
}

// Data structure representing a response from the load endpoint.
type LoadResponse struct {
	Message string
}

func Load(db db.InviteRepository, r *http.Request) (int, APIResponse) {
	requestBody, err := io.ReadAll(r.Body)
	if err != nil {
		return http.StatusInternalServerError, APIResponse{
			Error: &APIError{Code: "internal_server_error", Message: "error processing request"},
		}
	}

	loadRequest := &LoadRequest{Invites: make([]model.Invite, 128)}
	err = json.Unmarshal(requestBody, &loadRequest)
	if err != nil {
		return http.StatusBadRequest, APIResponse{
			Error: &APIError{Code: "bad_request", Message: fmt.Sprintf("bad load request: %s", err.Error())},
		}
	}

	err = db.Load(loadRequest.Invites)

	if err != nil {
		return http.StatusInternalServerError, APIResponse{
			Error: &APIError{Code: "load_failed", Message: fmt.Sprintf("error loading data into database: %s", err.Error())},
		}
	}
	return http.StatusOK, APIResponse{Error: nil, Response: LoadResponse{Message: "load OK"}}
}

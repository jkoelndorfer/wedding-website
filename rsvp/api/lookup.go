package api

import (
	"fmt"
	"net/http"

	"github.com/jkoelndorfer/wedding-website/rsvp/db"
	"github.com/jkoelndorfer/wedding-website/rsvp/model"
)

type LookupResponse struct {
	Invite *model.Invite `json:"invite"`
}

func Lookup(db db.InviteRepository, r *http.Request) (int, APIResponse) {
	queryParams := r.URL.Query()
	inviteId := queryParams.Get("invite_id")

	if inviteId == "" {
		return http.StatusBadRequest, APIResponse{
			Error: &APIError{Code: "no_invite_id", Message: fmt.Sprintf("expected paramater invite_id to contain invite ID")},
		}
	}
	if len(inviteId) != 8 {
		return http.StatusBadRequest, APIResponse{
			Error: &APIError{Code: "invalid_invite_id", Message: fmt.Sprintf("expected invite ID to be 8 characters")},
		}
	}

	invite, err := db.Get(model.InviteId(inviteId))
	if err != nil {
		return http.StatusInternalServerError, APIResponse{
			Error: &APIError{Code: "invite_lookup_error", Message: fmt.Sprintf("failed looking up invite")},
		}
	}

	return http.StatusOK, APIResponse{
		Error: nil, Response: LookupResponse{Invite: invite},
	}
}

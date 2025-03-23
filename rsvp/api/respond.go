package api

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/jkoelndorfer/wedding-website/rsvp/db"
	"github.com/jkoelndorfer/wedding-website/rsvp/model"
)

type RespondRequest struct {
	Id       model.InviteId             `json:"id"`
	Response []model.IndividualResponse `json:"response"`
}

type RespondResponse struct {
	Id      model.InviteId `json:"id"`
	Message string         `json:"message"`
}

func Respond(db db.InviteRepository, r *http.Request) (int, APIResponse) {
	requestBody, err := io.ReadAll(r.Body)
	if err != nil {
		return http.StatusInternalServerError, APIResponse{
			Error: &APIError{Code: "internal_server_error", Message: "error processing request"},
		}
	}
	respondRequest := &RespondRequest{}
	json.Unmarshal(requestBody, respondRequest)

	response := model.InviteResponse{}
	response.InviteId = respondRequest.Id
	response.Response = respondRequest.Response
	response.ResponseTime = time.Now()

	err = db.PutResponse(response)

	if err != nil {
		return http.StatusInternalServerError, APIResponse{
			Error: &APIError{Code: "failed_recording_response", Message: "failed recording invite response"},
		}
	}
	return http.StatusOK, APIResponse{
		Error: nil, Response: RespondResponse{Id: respondRequest.Id, Message: "successfully recorded response"},
	}
}

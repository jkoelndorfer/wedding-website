package api

import (
	"net/http"
)

// Signature for a standard RSVP handler function.
type HandlerFunction func(*http.Request) (statusCode int, body string)

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Data structure representing a standard API response.
type APIResponse struct {
	Error    APIError    `json:"error"`
	Response interface{} `json:"response"`
}

// Data structure representing an invited person.
type InvitedPerson struct {
	// A unique identifer for the invited person.
	id string

	// The person's name.
	name string

	// The number of additional guests a person is permitted to bring.
	plus int
}

type Invitation struct {
	// The unique identifier of the invite.
	//
	// Users will use this identifier to look their invitation up.
	id string

	// The salutation text of the invitation, sans-comma.
	//
	// For example, "Dear John & Jane Doe".
	salutation string

	// The people who are invited as part of this invitation.
	invitees []InvitedPerson
}

type InvitationResponse struct {
	// The unique identifier of the invitation that this response
	// corresponds to.
	inviteId string
}

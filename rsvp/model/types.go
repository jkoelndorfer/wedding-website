package model

import (
	"time"
)

// See https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue#Marshal
// for information about dynamodbav struct tgs.

// The data type representing an invitation ID.
type InviteId string

// The data type representing an invited person's ID.
type InvitedPersonId string

// Data structure representing an invited person.
type InvitedPerson struct {
	// A unique identifer for the invited person.
	Id InvitedPersonId `json:"id" dynamodbav:",string"`

	// The person's name.
	Name string `json:"name"`

	// True if this is a "plus one" person, false otherwise.
	Plus bool `json:"plus"`
}

// Data structure representing an invitation.
type Invitation struct {
	// The unique identifier of the invite.
	//
	// Users will use this identifier to look their invitation up.
	Id InviteId `json:"id" dyanamodbav:"InviteId,string"`

	// The salutation text of the invitation, sans-comma.
	//
	// For example, "Dear John & Jane Doe".
	Salutation string `json:"salutation" dynamodbav:",string"`

	// The people who are invited as part of this invitation.
	Invitees []InvitedPerson `json:"invitees"`

	// Whether the invitation includes the ceremony.
	CeremonyInvite bool `json:"ceremony_invite" dynamodbav:",bool"`

	// Whether the invitation includes the reception.
	ReceptionInvite bool `json:"reception_invite" dynamodbav:",bool"`

	// The number of additional guests allowed with this invitation.
	Plus int `json:"plus" dynamodbav:",number"`
}

// Data structure representing an invitation response for a single
// InvitedPerson.
type IndividualResponse struct {
	// The ID of the InvitedPerson that this response corresponds with.
	PersonId string `json:"person_id"`

	// True if this person indicates they will attend the ceremony; false otherwise.
	AttendingCeremony bool `json:"attending_ceremony"`

	// True if this person indicates they will attend the reception; false otherwise.
	AttendingReception bool `json:"attending_reception"`
}

type InvitationResponse struct {
	// The invitation that this response corresponds to.
	Invite *Invitation

	// The invitation ID that this response corresponds to.
	InviteId InviteId `json:"invite_id" dynamodbav:",string"`

	// The time that the response was submitted.
	ResponseTime time.Time `json:"response_time" dynamodbav:",string"`

	// Responses for each individual on the RSVP.
	Responses []IndividualResponse `json:"responses"`
}

package model

import (
	"github.com/jkoelndorfer/wedding-website/rsvp/log"
)

// Returns true if the invite response is valid for the given invite.
func (r *InviteResponse) ValidFor(i *Invite) bool {
	l := log.Logger()

	for _, individualResponse := range r.Response {
		personIdInInvite := false

		for _, invitePerson := range i.Invitees {
			if individualResponse.PersonId == invitePerson.Id {
				personIdInInvite = true
				break
			}
		}

		if !personIdInInvite {
			l.Printf("invite response contains person ID %s not present in invite", individualResponse.PersonId)
			return false
		}
	}

	for _, invitePerson := range i.Invitees {
		personIdInResponse := false

		for _, responsePerson := range r.Response {
			if invitePerson.Id == responsePerson.PersonId {
				personIdInResponse = true
				break
			}
		}

		if !personIdInResponse {
			l.Printf("invite contains person ID %s not present in response", invitePerson.Id)
			return false
		}
	}

	return true
}

package db

import (
	"errors"

	"github.com/jkoelndorfer/wedding-website/rsvp/model"
)

type DummyInvitationRepository struct{}

func dummyErr() error {
	return errors.New("this is a dummy repository")
}

func NewDummy() *DummyInvitationRepository {
	return &DummyInvitationRepository{}
}

func (d *DummyInvitationRepository) Get(invitation model.InviteId) (model.Invitation, error) {
	return model.Invitation{}, dummyErr()
}

func (d *DummyInvitationRepository) Load(invitations []model.Invitation) error {
	return dummyErr()
}

func (d *DummyInvitationRepository) Put(invitation model.Invitation) error {
	return dummyErr()
}

func (d *DummyInvitationRepository) PutResponse(inviteResponse model.InvitationResponse) error {
	return dummyErr()
}

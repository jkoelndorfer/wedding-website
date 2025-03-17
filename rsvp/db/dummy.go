package db

import (
	"errors"

	"github.com/jkoelndorfer/wedding-website/rsvp/model"
)

type DummyInviteRepository struct{}

func dummyErr() error {
	return errors.New("this is a dummy repository")
}

func NewDummy() *DummyInviteRepository {
	return &DummyInviteRepository{}
}

func (d *DummyInviteRepository) Get(invite model.InviteId) (model.Invite, error) {
	return model.Invite{}, dummyErr()
}

func (d *DummyInviteRepository) Load(invite []model.Invite) error {
	return dummyErr()
}

func (d *DummyInviteRepository) Put(invite model.Invite) error {
	return dummyErr()
}

func (d *DummyInviteRepository) PutResponse(inviteResponse model.InviteResponse) error {
	return dummyErr()
}

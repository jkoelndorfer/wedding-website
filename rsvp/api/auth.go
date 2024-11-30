package api

import (
	"github.com/jkoelndorfer/wedding-website/rsvp/config"
)

type AuthenticationService struct {
	configuration config.RSVPConfig
}

func NewAuthenticationService(configuration config.RSVPConfig) *AuthenticationService {
	return &AuthenticationService{configuration: configuration}
}

// Returns true if the given API client token is valid; false otherwise.
func (s *AuthenticationService) APIClientTokenValid(token string) bool {
	actual_token, err := s.configuration.APIClientToken()

	if err != nil {
		return false
	}

	return token == actual_token
}

// Returns true if the given API secret key is valid; false otherwise.
func (s *AuthenticationService) APISecretKeyValid(key string) bool {
	actual_key, err := s.configuration.AuthSecretKey()

	if err != nil {
		return false
	}

	return key == actual_key
}

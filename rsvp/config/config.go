package config

import (
	"errors"
	"os"
)

type RSVPConfig interface {
	// Returns the client token used for unprivileged operations against the API.
	APIClientToken() (string, error)

	// Returns the secret key required for specially-privileged endpoints.
	APISecretKey() (string, error)

	// Returns the endpoint to be used for DynamoDB.
	//
	// If the second argument is false, the endpoint should not be changed from the
	// default.
	DynamoDBEndpoint() (string, bool)

	// Returns true if this program is currently executing in Lambda; false otherwise.
	InLambda() bool

	// Returns true if this program is currently executing in a local development context; false otherwise.
	IsLocalDev() bool

	// Returns the name of the DynamoDB table that invitations and responses are stored in.
	InvitationsDynamoTable() (string, error)

	// Returns the address that this service should bind to.
	ListenAddress() string
}

type StandardRSVPConfig struct{}

func New() RSVPConfig {
	return &StandardRSVPConfig{}
}

// This is the client token used for unprivileged access to the API.
//
// Its purpose is to weed out drive-by bots banging on the API and
// incurring charges.
func (c *StandardRSVPConfig) APIClientToken() (string, error) {
	key, present := os.LookupEnv("API_CLIENT_TOKEN")

	if !present {
		return "", errors.New("API client token not set in environment")
	}

	if len(key) < 64 {
		return "", errors.New("API client token is too short")
	}

	return key, nil
}

// This is the secret key used for privileged access to the API.
//
// It permits loading invitation data.
func (c *StandardRSVPConfig) APISecretKey() (string, error) {
	key, present := os.LookupEnv("API_SECRET_KEY")

	if !present {
		return "", errors.New("authorization key not set in environment")
	}

	if len(key) < 64 {
		return "", errors.New("authorization key is too short")
	}

	return key, nil
}

// Returns the configured DynamoDB endpoint to use.
//
// If the second argument is false, no override should be used.
func (c *StandardRSVPConfig) DynamoDBEndpoint() (string, bool) {
	return os.LookupEnv("DYNAMODB_ENDPOINT")
}

// Indicates whether the RSVP application is running in AWS Lambda.
func (c *StandardRSVPConfig) InLambda() bool {
	_, lambdaEnvSet := os.LookupEnv("LAMBDA_TASK_ROOT")

	return lambdaEnvSet
}

// Indicates whether the RSVP application is running on a development system.
func (c *StandardRSVPConfig) IsLocalDev() bool {
	return !c.InLambda()
}

func (c *StandardRSVPConfig) InvitationsDynamoTable() (string, error) {
	tableName, valueSet := os.LookupEnv("DYNAMODB_INVITATIONS_TABLE")

	if !valueSet {
		return "", errors.New("DYNAMODB_INVITATIONS_TABLE not set in environment")
	}

	return tableName, nil
}

func (c *StandardRSVPConfig) ListenAddress() string {
	address, valueSet := os.LookupEnv("RSVP_LISTEN_ADDRESS")
	if !valueSet {
		address = "127.0.0.1:9000"
	}

	return address
}

// Ensure StandardRSVPConfig satisfies the RSVPConfig interface.
var _ RSVPConfig = &StandardRSVPConfig{}

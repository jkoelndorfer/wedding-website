package config

import (
	"crypto/rand"
	"errors"
	"fmt"
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

	// Returns the name of the DynamoDB table that invites and most recent responses are stored in.
	InvitesDynamoTable() (string, error)

	// Returns the address that this service should bind to.
	ListenAddress() string

	// Returns the name of the DynamoDB table that the response log is written to.
	//
	// All responses are recorded in this table.
	ResponseLogDynamoTable() (string, error)
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
	return secretEnv("API_CLIENT_TOKEN", 64)
}

// This is the secret key used for privileged access to the API.
//
// It permits loading invite data.
func (c *StandardRSVPConfig) APISecretKey() (string, error) {
	return secretEnv("API_SECRET_KEY", 64)
}

// Returns the configured DynamoDB endpoint to use.
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

func (c *StandardRSVPConfig) InvitesDynamoTable() (string, error) {
	return envOrError("DYNAMODB_INVITES_TABLE")
}

func (c *StandardRSVPConfig) ResponseLogDynamoTable() (string, error) {
	return envOrError("DYNAMODB_RESPONSE_LOG_TABLE")
}

func (c *StandardRSVPConfig) ListenAddress() string {
	return envOrDefault("RSVP_LISTEN_ADDRESS", "127.0.0.1:9000")
}

func envOrError(variableName string) (string, error) {
	value, valueIsSet := os.LookupEnv(variableName)

	if !valueIsSet {
		return "", errors.New(fmt.Sprintf("%s required, but not set in the environment", variableName))
	}

	return value, nil
}

func envOrDefault(variableName string, defaultValue string) string {
	value, valueIsSet := os.LookupEnv(variableName)

	if !valueIsSet {
		return defaultValue
	}

	return value
}

func secretEnv(variableName string, minLength int) (string, error) {
	secret, err := envOrError(variableName)

	if err != nil {
		// Just in case the error we return is not handled properly,
		// return a very large, random value. This should prevent
		// accidentally allowing wide-open authentication.
		return randString(minLength * 4), err
	}

	if len(secret) < minLength {
		return "", errors.New(fmt.Sprintf("%s: environment value too short", variableName))
	}

	return secret, nil
}

// Generates a random string of length n.
//
// Used defensively when secret values cannot be read.
func randString(n int) string {
	const charset = "0123456789abcdef"
	byteArray := make([]byte, n)
	rand.Read(byteArray)
	for i, c := range byteArray {
		byteArray[i] = charset[c%byte(len(charset))]
	}
	return string(byteArray)
}

// Ensure StandardRSVPConfig satisfies the RSVPConfig interface.
var _ RSVPConfig = &StandardRSVPConfig{}

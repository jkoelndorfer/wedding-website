package db

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	rsvpconfig "github.com/jkoelndorfer/wedding-website/rsvp/config"
	"github.com/jkoelndorfer/wedding-website/rsvp/log"
	"github.com/jkoelndorfer/wedding-website/rsvp/model"
)

var logger = log.Logger()

type InvitationRepository interface {
	Get(model.InviteId) (model.Invitation, error)
	Load([]model.Invitation) error
	Put(model.Invitation) error
	PutResponse(model.InvitationResponse) error
}

type DynamoDBInvitationRepository struct {
	dynamoDBClient      *dynamodb.Client
	invitationTableName string
}

func New(lcfg rsvpconfig.RSVPConfig) *DynamoDBInvitationRepository {
	opts := make([]func(*config.LoadOptions) error, 0, 8)

	if lcfg.IsLocalDev() {
		opts = append(opts, config.WithRegion("local-mock"))
		opts = append(opts, config.WithBaseEndpoint("http://localhost:8000"))
	}

	cfg, err := config.LoadDefaultConfig(context.Background(), opts...)

	if err != nil {
		logger.Fatalf("failed to load AWS default configuration for InvitationRepository: %v", err)
	}

	tableName, err := lcfg.InvitationsDynamoTable()
	if err != nil {
		logger.Fatalf("unable to determine DynamoDB table for invitations: %v", err)
	}
	return &DynamoDBInvitationRepository{
		dynamoDBClient:      dynamodb.NewFromConfig(cfg),
		invitationTableName: tableName,
	}
}

func (r *DynamoDBInvitationRepository) Get(invitationId model.InviteId) (model.Invitation, error) {
	return model.Invitation{}, nil
}

func (r *DynamoDBInvitationRepository) Load(invitations []model.Invitation) error {
	for _, inv := range invitations {
		err := r.Put(inv)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *DynamoDBInvitationRepository) Put(invitation model.Invitation) error {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(2*time.Second))
	defer cancel()

	av, err := attributevalue.MarshalMap(invitation)
	if err != nil {
		return err
	}

	// See https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/dynamodb#PutItemInput
	putItemInput := dynamodb.PutItemInput{
		TableName: &r.invitationTableName,
		Item:      av,
	}
	_, err = r.dynamoDBClient.PutItem(ctx, &putItemInput)

	return err
}

func (r *DynamoDBInvitationRepository) PutResponse(invitationResponse model.InvitationResponse) error {
	return nil
}

// Ensure DynamoDBInvitiationRepository implements the InvitationRepository interface.
var _ InvitationRepository = &DynamoDBInvitationRepository{}

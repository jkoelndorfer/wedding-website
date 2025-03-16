package db

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

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
	initialized         bool
	invitationTableName string
	localDev            bool
}

func New(lcfg rsvpconfig.RSVPConfig) *DynamoDBInvitationRepository {
	opts := make([]func(*config.LoadOptions) error, 0, 8)

	if lcfg.IsLocalDev() {
		opts = append(opts, config.WithRegion("local-mock"))
		dbEndpoint, overrideEndpoint := lcfg.DynamoDBEndpoint()
		if overrideEndpoint {
			opts = append(opts, config.WithBaseEndpoint(dbEndpoint))
		}
	}

	cfg, err := config.LoadDefaultConfig(context.Background(), opts...)

	if err != nil {
		logger.Fatalf("failed to load AWS default configuration for InvitationRepository: %v", err)
	}

	tableName, err := lcfg.InvitationsDynamoTable()
	if err != nil {
		logger.Fatalf("unable to determine DynamoDB table for invitations: %v", err)
	}

	logger.Printf("using dynamoDB client with config: %v", cfg)
	repository := &DynamoDBInvitationRepository{
		dynamoDBClient:      dynamodb.NewFromConfig(cfg),
		initialized:         false,
		invitationTableName: tableName,
		localDev:            lcfg.IsLocalDev(),
	}
	err = repository.Initialize()
	if err != nil {
		logger.Fatalf("failed initializing invitation repository: %v", err)
	}
	return repository
}

func (r *DynamoDBInvitationRepository) Get(invitationId model.InviteId) (model.Invitation, error) {
	return model.Invitation{}, nil
}

func (r *DynamoDBInvitationRepository) Initialize() error {
	if r.initialized {
		return nil
	}

	// We don't need to perform any initialization outside local development.
	//
	// The table should already be created by Terraform when running in AWS.
	if !r.localDev {
		r.initialized = true
		return nil
	}
	logger.Printf("preparing to initialize invitation repository")

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(5*time.Second))
	defer cancel()

	tableActive, err := r.tableIsActive(ctx, r.invitationTableName)
	// The invitiatons table has already been created. No need to try
	// creating it again.
	if tableActive {
		logger.Printf("invitation repository already initalized; nothing to do")
		r.initialized = true
		return nil
	}

	inviteIdAttrName := "InviteId"
	var capacity int64 = 1024
	input := dynamodb.CreateTableInput{
		TableName: &r.invitationTableName,
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: &inviteIdAttrName,
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: &inviteIdAttrName,
				KeyType:       types.KeyTypeHash,
			},
		},
		BillingMode: types.BillingModeProvisioned,
		ProvisionedThroughput: &types.ProvisionedThroughput{
			ReadCapacityUnits:  &capacity,
			WriteCapacityUnits: &capacity,
		},
	}

	_, err = r.dynamoDBClient.CreateTable(ctx, &input)

	if err != nil {
		return err
	}

	for !tableActive {
		select {
		case <-ctx.Done():
			return fmt.Errorf("invitations table did not initialize in before deadline")
		default:
			tableActive, err = r.tableIsActive(ctx, r.invitationTableName)
		}
	}

	return nil
}

func (r *DynamoDBInvitationRepository) tableIsActive(ctx context.Context, tableName string) (bool, error) {
	describeTableInput := dynamodb.DescribeTableInput{TableName: &tableName}

	output, err := r.dynamoDBClient.DescribeTable(ctx, &describeTableInput)
	if err != nil {
		return false, err
	}

	return output.Table.TableStatus == "ACTIVE", nil
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
	logger.Printf("putting item with TableName %v and item %v", r.invitationTableName, putItemInput.Item)
	_, err = r.dynamoDBClient.PutItem(ctx, &putItemInput)

	return err
}

func (r *DynamoDBInvitationRepository) PutResponse(invitationResponse model.InvitationResponse) error {
	return nil
}

// Ensure DynamoDBInvitiationRepository implements the InvitationRepository interface.
var _ InvitationRepository = &DynamoDBInvitationRepository{}

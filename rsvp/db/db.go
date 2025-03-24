package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/smithy-go"

	rsvpconfig "github.com/jkoelndorfer/wedding-website/rsvp/config"
	"github.com/jkoelndorfer/wedding-website/rsvp/log"
	"github.com/jkoelndorfer/wedding-website/rsvp/model"
)

var logger = log.Logger()
var inviteIdAttrName = "InviteId"

type InviteRepository interface {
	Get(model.InviteId) (*model.Invite, error)
	Load([]model.Invite) error
	Put(model.Invite) error
	PutResponse(model.InviteResponse) error
}

type DynamoDBInviteRepository struct {
	dynamoDBClient       *dynamodb.Client
	initialized          bool
	inviteTableName      string
	responseLogTableName string
	localDev             bool
}

// Interface to describe a basic DynamoDB client.
//
// A custom interface is used to facilitate a mock client being used
// during tests.
type DynamoDBClient interface {
	CreateTable(ctx context.Context, params *dynamodb.CreateTableInput, optFns ...func(*dynamodb.Options)) (*dynamodb.CreateTableOutput, error)
	DescribeTable(ctx context.Context, params *dynamodb.DescribeTableInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DescribeTableOutput, error)
	GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
	PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
}

func New(lcfg rsvpconfig.RSVPConfig) *DynamoDBInviteRepository {
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
		logger.Fatalf("failed to load AWS default configuration for InviteRepository: %v", err)
	}

	inviteTableName, err := lcfg.InvitesDynamoTable()
	if err != nil {
		logger.Fatalf("unable to determine DynamoDB table for invites: %v", err)
	}

	responseLogTableName, err := lcfg.InviteResponseLogDynamoTable()
	if err != nil {
		logger.Fatalf("unable to determine DynamoDB table for response logging: %v", err)
	}

	logger.Printf("using dynamoDB client with config: %v", cfg)
	repository := &DynamoDBInviteRepository{
		dynamoDBClient:       dynamodb.NewFromConfig(cfg),
		initialized:          false,
		inviteTableName:      inviteTableName,
		responseLogTableName: responseLogTableName,
		localDev:             lcfg.IsLocalDev(),
	}
	err = repository.Initialize()
	if err != nil {
		logger.Fatalf("failed initializing invite repository: %v", err)
	}
	return repository
}

func (r *DynamoDBInviteRepository) Get(inviteId model.InviteId) (*model.Invite, error) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(3*time.Second))
	defer cancel()

	consistentRead := true
	getItemInput := dynamodb.GetItemInput{
		TableName: &r.inviteTableName,
		Key: map[string]types.AttributeValue{
			"InviteId": &types.AttributeValueMemberS{Value: string(inviteId)},
		},
		ConsistentRead: &consistentRead,
	}
	output, err := r.dynamoDBClient.GetItem(ctx, &getItemInput)
	if err != nil {
		return nil, err
	}
	invite := model.Invite{}
	attributevalue.UnmarshalMap(output.Item, &invite)
	if invite.InviteId == "" {
		logger.Printf("unmarshalled invite: %+v", invite)
		return nil, fmt.Errorf("no such invite: %s", inviteId)
	}
	return &invite, nil
}

func (r *DynamoDBInviteRepository) Initialize() error {
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
	logger.Printf("preparing to initialize invite repository")

	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(5*time.Second))
	defer cancel()

	invChannel := make(chan error)
	respChannel := make(chan error)

	go r.initializeInviteTable(ctx, invChannel)
	go r.initializeInviteResponseLogTable(ctx, respChannel)

	invTableActive := false
	respTableActive := false
	var e error
	for !(invTableActive && respTableActive) {
		select {
		case <-ctx.Done():
			return fmt.Errorf("tables did not initialize before deadline")
		case e = <-invChannel:
			if e != nil {
				return e
			} else {
				logger.Printf("invites table is ready")
				invTableActive = true
			}
		case e = <-respChannel:
			if e != nil {
				return e
			} else {
				logger.Printf("response log table is ready")
				respTableActive = true
			}
		}
	}

	return nil
}

func (r *DynamoDBInviteRepository) initializeTable(
	ctx context.Context,
	tableName string,
	attributeDefinitions []types.AttributeDefinition,
	keySchema []types.KeySchemaElement,
) error {
	// We don't need to perform any initialization outside local development.
	//
	// The table should already be created by Terraform when running in AWS.
	//
	// We definitely do not want to inadvertently create any tables which
	// would use a provisioned billing mode.
	if !r.localDev {
		return nil
	}

	tableStatus, err := r.tableStatus(ctx, tableName)

	if err != nil {
		return err
	}

	if tableStatus == "ACTIVE" {
		return nil
	}

	if tableStatus == "NOT_CREATED" {
		var capacity int64 = 1024
		input := dynamodb.CreateTableInput{
			TableName:            &tableName,
			AttributeDefinitions: attributeDefinitions,
			KeySchema:            keySchema,
			// Local DynamoDB requires us to create tables with the provisioned billing mode.
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
	}

	for tableStatus != "ACTIVE" {
		select {
		case <-ctx.Done():
			return fmt.Errorf("canceled waiting for table %s to be ready", tableName)
		default:
			tableStatus, err = r.tableStatus(ctx, tableName)
		}
	}

	return nil
}

func (r *DynamoDBInviteRepository) initializeInviteTable(ctx context.Context, resp chan<- error) {
	resp <- r.initializeTable(
		ctx,
		r.inviteTableName,
		[]types.AttributeDefinition{
			{
				AttributeName: &inviteIdAttrName,
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		[]types.KeySchemaElement{
			{
				AttributeName: &inviteIdAttrName,
				KeyType:       types.KeyTypeHash,
			},
		},
	)
}

func (r *DynamoDBInviteRepository) initializeInviteResponseLogTable(ctx context.Context, resp chan<- error) {
	responseTimeAttrName := "ResponseTime"
	resp <- r.initializeTable(
		ctx,
		r.responseLogTableName,
		[]types.AttributeDefinition{
			{
				AttributeName: &inviteIdAttrName,
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: &responseTimeAttrName,
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		[]types.KeySchemaElement{
			{
				AttributeName: &inviteIdAttrName,
				KeyType:       types.KeyTypeHash,
			},
			{
				AttributeName: &responseTimeAttrName,
				KeyType:       types.KeyTypeRange,
			},
		},
	)
}

func (r *DynamoDBInviteRepository) tableStatus(ctx context.Context, tableName string) (string, error) {
	describeTableInput := dynamodb.DescribeTableInput{TableName: &tableName}
	output, err := r.dynamoDBClient.DescribeTable(ctx, &describeTableInput)

	var apiErr smithy.APIError
	if err != nil {
		if errors.As(err, &apiErr) {
			if apiErr.ErrorCode() == "ResourceNotFoundException" {
				return "NOT_CREATED", nil
			}
		}

		return "UNKNOWN", err
	}

	return string(output.Table.TableStatus), nil
}

func (r *DynamoDBInviteRepository) Load(invites []model.Invite) error {
	for _, inv := range invites {
		err := r.Put(inv)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *DynamoDBInviteRepository) Put(invite model.Invite) error {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(2*time.Second))
	defer cancel()

	av, err := attributevalue.MarshalMap(invite)
	if err != nil {
		return err
	}

	// See https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/dynamodb#PutItemInput
	putItemInput := dynamodb.PutItemInput{
		TableName: &r.inviteTableName,
		Item:      av,
	}
	logger.Printf("putting item with TableName %v and item %v", r.inviteTableName, putItemInput.Item)
	_, err = r.dynamoDBClient.PutItem(ctx, &putItemInput)

	return err
}

func (r *DynamoDBInviteRepository) PutResponse(inviteResponse model.InviteResponse) error {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(5*time.Second))
	defer cancel()

	invite, err := r.Get(inviteResponse.InviteId)
	if err != nil {
		return err
	}

	if !inviteResponse.ValidFor(invite) {
		logger.Printf("invite response is not valid for invite")
		return fmt.Errorf("invite response is not valid for invite")
	}

	responseAV, err := attributevalue.MarshalMap(inviteResponse)
	if err != nil {
		logger.Printf("failed marshalling inviteResponse into DynamoDB attributevalue map: %s", err.Error())
		return err
	}

	invite.Response = &inviteResponse
	inviteAV, err := attributevalue.MarshalMap(invite)
	if err != nil {
		logger.Printf("failed marshalling invite into DynamoDB attributevalue map: %s", err.Error())
		return err
	}

	responseLogPut := dynamodb.PutItemInput{
		TableName: &r.responseLogTableName,
		Item:      responseAV,
	}
	_, err = r.dynamoDBClient.PutItem(ctx, &responseLogPut)
	if err != nil {
		logger.Printf("failed recording response in response log table")
		return err
	}

	invitePut := dynamodb.PutItemInput{
		TableName: &r.inviteTableName,
		Item:      inviteAV,
	}
	_, err = r.dynamoDBClient.PutItem(ctx, &invitePut)

	if err != nil {
		logger.Printf("failed recording response in invite table")
		return err
	}
	return nil
}

// Ensure DynamoDBInvitiationRepository implements the InviteRepository interface.
var _ InviteRepository = &DynamoDBInviteRepository{}

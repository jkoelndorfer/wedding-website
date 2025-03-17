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

type InviteRepository interface {
	Get(model.InviteId) (*model.Invite, error)
	Load([]model.Invite) error
	Put(model.Invite) error
	PutResponse(model.InviteResponse) error
}

type DynamoDBInviteRepository struct {
	dynamoDBClient  *dynamodb.Client
	initialized     bool
	inviteTableName string
	localDev        bool
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

	tableName, err := lcfg.InvitesDynamoTable()
	if err != nil {
		logger.Fatalf("unable to determine DynamoDB table for invites: %v", err)
	}

	logger.Printf("using dynamoDB client with config: %v", cfg)
	repository := &DynamoDBInviteRepository{
		dynamoDBClient:  dynamodb.NewFromConfig(cfg),
		initialized:     false,
		inviteTableName: tableName,
		localDev:        lcfg.IsLocalDev(),
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

	tableActive, err := r.tableIsActive(ctx, r.inviteTableName)
	// The invitiatons table has already been created. No need to try
	// creating it again.
	if tableActive {
		logger.Printf("invite repository already initalized; nothing to do")
		r.initialized = true
		return nil
	}

	inviteIdAttrName := "InviteId"
	var capacity int64 = 1024
	input := dynamodb.CreateTableInput{
		TableName: &r.inviteTableName,
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
			return fmt.Errorf("invites table did not initialize in before deadline")
		default:
			tableActive, err = r.tableIsActive(ctx, r.inviteTableName)
		}
	}

	return nil
}

func (r *DynamoDBInviteRepository) tableIsActive(ctx context.Context, tableName string) (bool, error) {
	describeTableInput := dynamodb.DescribeTableInput{TableName: &tableName}

	output, err := r.dynamoDBClient.DescribeTable(ctx, &describeTableInput)
	if err != nil {
		return false, err
	}

	return output.Table.TableStatus == "ACTIVE", nil
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
	return nil
}

// Ensure DynamoDBInvitiationRepository implements the InviteRepository interface.
var _ InviteRepository = &DynamoDBInviteRepository{}

package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

func createTableIfNotExist(ctx context.Context, client *dynamodb.Client, input *dynamodb.CreateTableInput) (*dynamodb.CreateTableOutput, bool, error) {
	tableName := input.TableName

	_, err := client.DescribeTable(ctx, &dynamodb.DescribeTableInput{TableName: tableName})
	// 既にテーブルが存在していれば作成を試みずに終了
	if err == nil {
		return nil, false, nil
	}

	// エラーがテーブルが存在しないものによるものであれば作成
	if _, ok := errors.AsType[*types.ResourceNotFoundException](err); ok {
		output, err := client.CreateTable(ctx, input)
		if err != nil {
			return nil, false, fmt.Errorf("failed to create table: %w", err)
		}
		return output, true, nil
	}

	// その他のエラーはエラーとして扱う
	return nil, false, fmt.Errorf("failed to describe table: %w", err)
}

func initializeTables(ctx context.Context, client *dynamodb.Client) error {
	_, _, err := createTableIfNotExist(ctx, client, &dynamodb.CreateTableInput{
		TableName: new("ToyChatApp"),
		KeySchema: []types.KeySchemaElement{
			{
				AttributeName: new("PK"),
				KeyType:       types.KeyTypeHash,
			},
			{
				AttributeName: new("SK"),
				KeyType:       types.KeyTypeRange,
			},
		},
		AttributeDefinitions: []types.AttributeDefinition{
			{
				AttributeName: new("PK"),
				AttributeType: types.ScalarAttributeTypeS,
			},
			{
				AttributeName: new("SK"),
				AttributeType: types.ScalarAttributeTypeS,
			},
		},
		BillingMode:                        types.BillingModePayPerRequest,
		GlobalTableSettingsReplicationMode: types.GlobalTableSettingsReplicationModeDisabled,
		TableClass:                         types.TableClassStandard,
		DeletionProtectionEnabled:          new(false),
		GlobalSecondaryIndexes:             nil,
		GlobalTableSourceArn:               nil,
		LocalSecondaryIndexes:              nil,
		OnDemandThroughput:                 nil,
		ProvisionedThroughput:              nil,
		ResourcePolicy:                     nil,
		SSESpecification:                   nil,
		StreamSpecification:                nil,
		Tags:                               nil,
		WarmThroughput:                     nil,
	})
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}
	return nil
}

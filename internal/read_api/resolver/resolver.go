package resolver

import "github.com/aws/aws-sdk-go-v2/service/dynamodb"

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require
// here.

type Resolver struct {
	DynamoDBClient *dynamodb.Client
}

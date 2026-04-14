package resolver

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"slices"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/sudame/chat/internal/read_api/model"
)

type JoinedRoom struct {
	PK           string `dynamodbav:"PK"` // USER#<id>
	SK           string `dynamodbav:"SK"` // ROOM#<id>
	MembershipID string `dynamodbav:"membership_id"`
	RoomID       string `dynamodbav:"room_id"`
	UserID       string `dynamodbav:"user_id"`
}

type Cursor struct {
	PK string `dynamodbav:"PK"`
	SK string `dynamodbav:"SK"`
}

type QueryParams struct {
	PK       string
	SKPrefix *string
	First    *int32
	After    *string
	Last     *int32
	Before   *string
}

type Edge[Node any] struct {
	Node   *Node
	Cursor string
}

type Connection[Node any] struct {
	Items    []*Edge[Node]
	PageInfo *model.PageInfo
}

func EncodeCursor(cursor Cursor) (*string, error) {
	j, err := json.Marshal(cursor)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal cursor to json: %w", err)
	}
	b := base64.StdEncoding.EncodeToString(j)

	return new(b), nil
}

func DecodeCursor(encodedCursor string) (map[string]types.AttributeValue, error) {
	j, err := base64.StdEncoding.DecodeString(encodedCursor)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %w", err)
	}

	var cursor Cursor
	if err := json.Unmarshal(j, &cursor); err != nil {
		return nil, fmt.Errorf("failed to unmarshal json to cursor: %w", err)
	}

	av, err := attributevalue.MarshalMap(cursor)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal cursor to attribute value: %w", err)
	}

	return av, nil
}

type QueryForwardParams struct {
	PK       string
	SKPrefix *string
	First    *int32
	After    *string
}

func (p QueryParams) ToQueryForwardParams() QueryForwardParams {
	return QueryForwardParams{
		PK:       p.PK,
		SKPrefix: p.SKPrefix,
		First:    p.First,
		After:    p.After,
	}
}

func QueryForward[Node any](ctx context.Context, client *dynamodb.Client, tableName string, params QueryForwardParams) (*Connection[Node], error) {
	first := *params.First
	var skPrefix string
	if params.SKPrefix != nil {
		skPrefix = *params.SKPrefix
	}

	input := &dynamodb.QueryInput{
		TableName:              new(tableName),
		KeyConditionExpression: new("PK = :pk AND begins_with(SK, :sk_prefix)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk":        &types.AttributeValueMemberS{Value: params.PK},
			":sk_prefix": &types.AttributeValueMemberS{Value: skPrefix},
		},
		Limit:            new(first + 1),
		ScanIndexForward: new(true),
	}

	if params.After != nil {
		cursorAv, err := DecodeCursor(*params.After)
		if err != nil {
			return nil, fmt.Errorf("failed to decode cursor: %w", err)
		}
		input.ExclusiveStartKey = cursorAv
	}

	result, err := client.Query(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to query: %w", err)
	}

	rawItems := slices.Clone(result.Items)

	if len(rawItems) == 0 {
		return &Connection[Node]{
			Items: make([]*Edge[Node], 0),
			PageInfo: &model.PageInfo{
				HasNextPage:     false,
				HasPreviousPage: false,
			},
		}, nil
	}

	hasNextPage := len(rawItems) > int(first)
	if hasNextPage {
		rawItems = rawItems[:first]
	}

	cursors := make([]Cursor, len(rawItems))
	for i, item := range rawItems {
		var cursor Cursor
		if err := attributevalue.UnmarshalMap(item, &cursor); err != nil {
			slog.ErrorContext(ctx, "failed to unmarshal item to cursor", "item", item, "err", err)
			continue
		}
		cursors[i] = cursor
	}

	startCursor, err := EncodeCursor(cursors[0])
	if err != nil {
		return nil, fmt.Errorf("failed to encode start cursor: %w", err)
	}
	endCursor, err := EncodeCursor(cursors[len(cursors)-1])
	if err != nil {
		return nil, fmt.Errorf("failed to encode end cursor: %w", err)
	}

	pageInfo := model.PageInfo{
		HasPreviousPage: params.After != nil,
		HasNextPage:     hasNextPage,
		StartCursor:     startCursor,
		EndCursor:       endCursor,
	}

	edges := make([]*Edge[Node], len(rawItems))
	for i, item := range rawItems {
		var node Node
		err := attributevalue.UnmarshalMap(item, &node)
		if err != nil {
			slog.ErrorContext(ctx, "failed to unmarshal item to node", "item", item, "err", err)
			continue
		}

		cursor, err := EncodeCursor(cursors[i])
		if err != nil {
			slog.ErrorContext(ctx, "failed to encode cursor", "item", item, "cursor", cursors[i], "err", err)
			continue
		}

		edge := Edge[Node]{
			Node:   &node,
			Cursor: *cursor,
		}

		edges[i] = &edge
	}

	connection := Connection[Node]{
		PageInfo: &pageInfo,
		Items:    edges,
	}

	return &connection, nil
}

type QueryBackwardParams struct {
	PK       string
	SKPrefix *string
	Last     *int32
	Before   *string
}

func (p QueryParams) ToQueryBackwardParams() QueryBackwardParams {
	return QueryBackwardParams{
		PK:       p.PK,
		SKPrefix: p.SKPrefix,
		Last:     p.Last,
		Before:   p.Before,
	}
}

func QueryBackward[Node any](ctx context.Context, client *dynamodb.Client, tableName string, params QueryBackwardParams) (*Connection[Node], error) {
	last := *params.Last
	var skPrefix string
	if params.SKPrefix != nil {
		skPrefix = *params.SKPrefix
	}

	input := &dynamodb.QueryInput{
		TableName:              new(tableName),
		KeyConditionExpression: new("PK = :pk AND begins_with(SK, :sk_prefix)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk":        &types.AttributeValueMemberS{Value: params.PK},
			":sk_prefix": &types.AttributeValueMemberS{Value: skPrefix},
		},
		Limit:            new(last + 1),
		ScanIndexForward: new(false),
	}

	if params.Before != nil {
		cursorAv, err := DecodeCursor(*params.Before)
		if err != nil {
			return nil, fmt.Errorf("failed to decode cursor: %w", err)
		}
		input.ExclusiveStartKey = cursorAv
	}

	result, err := client.Query(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to query: %w", err)
	}

	rawItems := slices.Clone(result.Items)
	slices.Reverse(rawItems)

	if len(rawItems) == 0 {
		return &Connection[Node]{
			Items: make([]*Edge[Node], 0),
			PageInfo: &model.PageInfo{
				HasNextPage:     false,
				HasPreviousPage: false,
			},
		}, nil
	}

	hasPreviousPage := len(rawItems) > int(last)
	if hasPreviousPage {
		rawItems = rawItems[1:]
	}

	cursors := make([]Cursor, len(rawItems))
	for i, item := range rawItems {
		var cursor Cursor
		if err := attributevalue.UnmarshalMap(item, &cursor); err != nil {
			slog.ErrorContext(ctx, "failed to unmarshal item to cursor", "item", item, "err", err)
			continue
		}
		cursors[i] = cursor
	}

	startCursor, err := EncodeCursor(cursors[0])
	if err != nil {
		return nil, fmt.Errorf("failed to encode start cursor: %w", err)
	}
	endCursor, err := EncodeCursor(cursors[len(cursors)-1])
	if err != nil {
		return nil, fmt.Errorf("failed to encode end cursor: %w", err)
	}

	pageInfo := model.PageInfo{
		HasPreviousPage: hasPreviousPage,
		HasNextPage:     params.Before != nil,
		StartCursor:     startCursor,
		EndCursor:       endCursor,
	}

	edges := make([]*Edge[Node], len(rawItems))
	for i, item := range rawItems {
		var node Node
		err := attributevalue.UnmarshalMap(item, &node)
		if err != nil {
			slog.ErrorContext(ctx, "failed to unmarshal item to node", "item", item, "err", err)
			continue
		}

		cursor, err := EncodeCursor(cursors[i])
		if err != nil {
			slog.ErrorContext(ctx, "failed to encode cursor", "item", item, "cursor", cursors[i], "err", err)
			continue
		}

		edge := Edge[Node]{
			Node:   &node,
			Cursor: *cursor,
		}

		edges[i] = &edge
	}

	connection := Connection[Node]{
		PageInfo: &pageInfo,
		Items:    edges,
	}

	return &connection, nil
}

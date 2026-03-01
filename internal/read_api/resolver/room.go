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
	PK     string
	First  *int32
	After  *string
	Last   *int32
	Before *string
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
	var b []byte
	base64.StdEncoding.Encode(b, j)

	return new(string(b)), nil
}

func DecodeCursor(encodedCursor string) (map[string]types.AttributeValue, error) {
	var j []byte
	_, err := base64.StdEncoding.Decode(j, []byte(encodedCursor))
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %w", err)
	}

	av, err := attributevalue.UnmarshalMapJSON(j)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal json to attribute value: %w", err)
	}

	return av, nil
}

type QueryForwardParams struct {
	PK    string
	First *int32
	After *string
}

func (p QueryParams) ToQueryForwardParams() QueryForwardParams {
	return QueryForwardParams{
		PK:    p.PK,
		First: p.First,
		After: p.After,
	}
}

func QueryForward[Node any](ctx context.Context, client *dynamodb.Client, tableName string, params QueryForwardParams) (*Connection[Node], error) {
	first := *params.First

	input := &dynamodb.QueryInput{
		TableName:                 new(tableName),
		KeyConditionExpression:    new("PK = :pk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{":pk": &types.AttributeValueMemberS{Value: params.PK}},
		Limit:                     new(first + 1),
		ScanIndexForward:          new(true),
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

	cursors := make([]Cursor, len(result.Items))
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
		HasNextPage:     len(rawItems) > int(first),
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
	PK     string
	Last   *int32
	Before *string
}

func (p QueryParams) ToQueryBackwardParams() QueryBackwardParams {
	return QueryBackwardParams{
		PK:     p.PK,
		Last:   p.Last,
		Before: p.Before,
	}
}

func QueryBackward[Node any](ctx context.Context, client *dynamodb.Client, tableName string, params QueryBackwardParams) (*Connection[Node], error) {
	last := *params.Last

	input := &dynamodb.QueryInput{
		TableName:                 new(tableName),
		KeyConditionExpression:    new("PK = :pk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{":pk": &types.AttributeValueMemberS{Value: params.PK}},
		Limit:                     new(last + 1),
		ScanIndexForward:          new(false),
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
		HasPreviousPage: len(rawItems) > int(last),
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

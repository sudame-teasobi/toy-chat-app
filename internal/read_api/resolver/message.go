package resolver

type Message struct {
	PK        string `dynamodbav:"PK"` // ROOM#<id>
	SK        string `dynamodbav:"SK"` // MESSAGE#<id>
	MessageID string `dynamodbav:"message_id"`
	Body      string `dynamodbav:"body"`
	RoomID    string `dynamodbav:"room_id"`
	UserID    string `dynamodbav:"user_id"`
}

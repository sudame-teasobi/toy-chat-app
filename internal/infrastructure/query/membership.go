package query

import (
	"fmt"

	"github.com/sudame/chat/internal/domain/membership"
	"github.com/sudame/chat/pkg/httpclient"
)

type MembershipQuery struct {
	client *httpclient.HTTPClient
}

func NewMembershipQuery(client *httpclient.HTTPClient) *MembershipQuery {
	return &MembershipQuery{client: client}
}

const CheckMembershipExistencePath = "/check-membership-existence"

// CheckMembershipExistence implements [membership.Query].
func (r *MembershipQuery) CheckMembershipExistence(req membership.CheckMembershipExistenceRequest) (membership.CheckMembershipExistenceResponse, error) {
	var zero membership.CheckMembershipExistenceResponse
	res, err := httpclient.Post[membership.CheckMembershipExistenceRequest, membership.CheckMembershipExistenceResponse](r.client, CheckMembershipExistencePath, req)
	if err != nil {
		return zero, fmt.Errorf("failed to post: %w", err)
	}

	return res, nil
}

var _ membership.Query = (*MembershipQuery)(nil)

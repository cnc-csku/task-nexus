package grpcserver

import (
	"context"

	taskv1 "github.com/cnc-csku/task-nexus/api-specification/gen/proto/task/v1"
	"github.com/cnc-csku/task-nexus/task-management/domain/requests"
	"github.com/cnc-csku/task-nexus/task-management/domain/services"
)

type MemberHandler struct {
	taskv1.UnimplementedMemberServiceServer
	services.MemberService
}

func (h *MemberHandler) GetMembers(ctx context.Context, in *taskv1.GetMembersRequest) (*taskv1.GetMembersResponse, error) {
	// Convert request
	req := &requests.GetMembersRequest{}

	_, err := h.MemberService.GetMembers(ctx, req)
	if err != nil {
		return nil, err
	}

	// Convert response
	protoResp := &taskv1.GetMembersResponse{}
	return protoResp, nil
}

package services

import (
	"context"
	"math"

	"github.com/cnc-csku/task-nexus/go-lib/utils/errutils"
	"github.com/cnc-csku/task-nexus/task-management/config"
	"github.com/cnc-csku/task-nexus/task-management/domain/constant"
	"github.com/cnc-csku/task-nexus/task-management/domain/exceptions"
	"github.com/cnc-csku/task-nexus/task-management/domain/models"
	"github.com/cnc-csku/task-nexus/task-management/domain/repositories"
	"github.com/cnc-csku/task-nexus/task-management/domain/requests"
	"github.com/cnc-csku/task-nexus/task-management/domain/responses"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type ProjectService interface {
	Create(ctx context.Context, req *requests.CreateProjectRequest, userId string) (*responses.CreateProjectResponse, *errutils.Error)
	ListMyProjects(ctx context.Context, req *requests.ListMyProjectsPathParams, userID string) ([]*models.Project, *errutils.Error)
	GetProjectDetail(ctx context.Context, req *requests.GetProjectsDetailPathParams, userID string) (*models.Project, *errutils.Error)
	AddPositions(ctx context.Context, req *requests.AddPositionsRequest, userID string) (*responses.AddPositionsResponse, *errutils.Error)
	ListPositions(ctx context.Context, req *requests.ListPositionsPathParams) ([]string, *errutils.Error)
	AddMembers(ctx context.Context, req *requests.AddProjectMembersRequest, userID string) (*responses.AddProjectMembersResponse, *errutils.Error)
	ListMembers(ctx context.Context, req *requests.ListProjectMembersRequest) (*responses.ListProjectMembersResponse, *errutils.Error)
	AddWorkflows(ctx context.Context, req *requests.AddWorkflowsRequest, userID string) (*responses.AddWorkflowsResponse, *errutils.Error)
	ListWorkflows(ctx context.Context, req *requests.ListWorkflowsPathParams) ([]models.Workflow, *errutils.Error)
}

type projectServiceImpl struct {
	userRepo      repositories.UserRepository
	workspaceRepo repositories.WorkspaceRepository
	projectRepo   repositories.ProjectRepository
	config        *config.Config
}

func NewProjectService(userRepo repositories.UserRepository, workspaceRepo repositories.WorkspaceRepository, projectRepo repositories.ProjectRepository, config *config.Config) ProjectService {
	return &projectServiceImpl{
		userRepo:      userRepo,
		workspaceRepo: workspaceRepo,
		projectRepo:   projectRepo,
		config:        config,
	}
}

func (p *projectServiceImpl) Create(ctx context.Context, req *requests.CreateProjectRequest, userId string) (*responses.CreateProjectResponse, *errutils.Error) {
	bsonUserId, err := bson.ObjectIDFromHex(userId)
	if err != nil {
		return nil, errutils.NewError(exceptions.ErrInternalError, errutils.BadRequest).WithDebugMessage(err.Error())
	}
	bsonWorkspaceID, err := bson.ObjectIDFromHex(req.WorkspaceID)
	if err != nil {
		return nil, errutils.NewError(exceptions.ErrInvalidWorkspaceID, errutils.BadRequest).WithDebugMessage(err.Error())
	}

	// Check if the creator is owner or moderator of the workspace
	member, err := p.workspaceRepo.FindWorkspaceMemberByWorkspaceIDAndUserID(ctx, bsonWorkspaceID, bsonUserId)
	if err != nil {
		return nil, errutils.NewError(exceptions.ErrInternalError, errutils.InternalError).WithDebugMessage(err.Error())
	} else if member == nil {
		return nil, errutils.NewError(exceptions.ErrMemberNotFoundInWorkspace, errutils.BadRequest)
	} else if member.Role != models.WorkspaceMemberRoleOwner && member.Role != models.WorkspaceMemberRoleModerator {
		return nil, errutils.NewError(exceptions.ErrPermissionDenied, errutils.BadRequest)
	}

	// Check if project's name already exists
	existsProjectByName, err := p.projectRepo.FindByWorkspaceIDAndName(ctx, bsonWorkspaceID, req.Name)
	if err != nil {
		return nil, errutils.NewError(exceptions.ErrInternalError, errutils.InternalError).WithDebugMessage(err.Error())
	}
	if existsProjectByName != nil {
		return nil, errutils.NewError(exceptions.ErrProjectNameAlreadyExists, errutils.BadRequest)
	}

	// Check if project's prefix already exists
	existsProjectByPrefix, err := p.projectRepo.FindByWorkspaceIDAndProjectPrefix(ctx, bsonWorkspaceID, req.ProjectPrefix)
	if err != nil {
		return nil, errutils.NewError(exceptions.ErrInternalError, errutils.InternalError).WithDebugMessage(err.Error())
	}
	if existsProjectByPrefix != nil {
		return nil, errutils.NewError(exceptions.ErrProjectPrefixAlreadyExists, errutils.BadRequest)
	}

	// Create owner member
	owner := &models.ProjectMember{
		UserID:      member.UserID,
		DisplayName: member.DisplayName,
		ProfileUrl:  member.ProfileUrl,
		Role:        models.ProjectMemberRoleOwner,
	}

	project := &repositories.CreateProjectRequest{
		WorkspaceID:   bsonWorkspaceID,
		Name:          req.Name,
		ProjectPrefix: req.ProjectPrefix,
		Description:   req.Description,
		Status:        models.ProjectStatusActive,
		Owner:         owner,
		Workflows:     models.GetDefaultWorkflows(),
		CreatedBy:     bsonUserId,
	}

	createdProject, err := p.projectRepo.Create(ctx, project)
	if err != nil {
		return nil, errutils.NewError(exceptions.ErrInternalError, errutils.InternalError).WithDebugMessage(err.Error())
	}

	res := &responses.CreateProjectResponse{
		ID:            createdProject.ID.Hex(),
		WorkspaceID:   createdProject.WorkspaceID.Hex(),
		Name:          createdProject.Name,
		ProjectPrefix: createdProject.ProjectPrefix,
		Description:   createdProject.Description,
	}

	return res, nil
}

func (p *projectServiceImpl) ListMyProjects(ctx context.Context, req *requests.ListMyProjectsPathParams, userID string) ([]*models.Project, *errutils.Error) {
	bsonWorkspaceID, err := bson.ObjectIDFromHex(req.WorkspaceID)
	if err != nil {
		return nil, errutils.NewError(exceptions.ErrInvalidWorkspaceID, errutils.BadRequest).WithDebugMessage(err.Error())
	}

	bsonUserID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errutils.NewError(exceptions.ErrInternalError, errutils.BadRequest).WithDebugMessage(err.Error())
	}

	projects, err := p.projectRepo.FindByWorkspaceIDAndUserID(ctx, bsonWorkspaceID, bsonUserID)
	if err != nil {
		return nil, errutils.NewError(exceptions.ErrInternalError, errutils.InternalError).WithDebugMessage(err.Error())
	}

	return projects, nil
}

func (p *projectServiceImpl) GetProjectDetail(ctx context.Context, req *requests.GetProjectsDetailPathParams, userID string) (*models.Project, *errutils.Error) {
	bsonProjectID, err := bson.ObjectIDFromHex(req.ProjectID)
	if err != nil {
		return nil, errutils.NewError(exceptions.ErrInternalError, errutils.BadRequest).WithDebugMessage(err.Error())
	}

	bsonUserID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errutils.NewError(exceptions.ErrInternalError, errutils.BadRequest).WithDebugMessage(err.Error())
	}

	// Check if the user is a member of the project
	member, err := p.projectRepo.FindMemberByProjectIDAndUserID(ctx, bsonProjectID, bsonUserID)
	if err != nil {
		return nil, errutils.NewError(exceptions.ErrInternalError, errutils.InternalError).WithDebugMessage(err.Error())
	} else if member == nil {
		return nil, errutils.NewError(exceptions.ErrPermissionDenied, errutils.BadRequest)
	}

	project, err := p.projectRepo.FindByProjectID(ctx, bsonProjectID)
	if err != nil {
		return nil, errutils.NewError(exceptions.ErrInternalError, errutils.InternalError).WithDebugMessage(err.Error())
	} else if project == nil {
		return nil, errutils.NewError(exceptions.ErrProjectNotFound, errutils.NotFound)
	}

	return project, nil
}

func (p *projectServiceImpl) AddPositions(ctx context.Context, req *requests.AddPositionsRequest, userID string) (*responses.AddPositionsResponse, *errutils.Error) {
	bsonProjectID, err := bson.ObjectIDFromHex(req.ProjectID)
	if err != nil {
		return nil, errutils.NewError(exceptions.ErrInternalError, errutils.BadRequest).WithDebugMessage(err.Error())
	}

	bsonUserID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errutils.NewError(exceptions.ErrInternalError, errutils.BadRequest).WithDebugMessage(err.Error())
	}

	// Check if the user is owner or moderator of the project
	member, err := p.projectRepo.FindMemberByProjectIDAndUserID(ctx, bsonProjectID, bsonUserID)
	if err != nil {
		return nil, errutils.NewError(exceptions.ErrInternalError, errutils.InternalError).WithDebugMessage(err.Error())
	} else if member == nil {
		return nil, errutils.NewError(exceptions.ErrUserNotFound, errutils.BadRequest)
	} else if member.Role != models.ProjectMemberRoleOwner && member.Role != models.ProjectMemberRoleModerator {
		return nil, errutils.NewError(exceptions.ErrPermissionDenied, errutils.BadRequest)
	}

	// Check if the position already exists
	existingPositions, err := p.projectRepo.FindPositionByProjectID(ctx, bsonProjectID)
	if err != nil {
		return nil, errutils.NewError(exceptions.ErrInternalError, errutils.InternalError).WithDebugMessage(err.Error())
	}

	positionMap := make(map[string]struct{})
	for _, position := range existingPositions {
		positionMap[position] = struct{}{}
	}

	var newPositions []string
	for _, position := range req.Title {
		if _, ok := positionMap[position]; !ok {
			newPositions = append(newPositions, position)
		}
	}

	if len(newPositions) == 0 {
		return &responses.AddPositionsResponse{
			Message: "No new position added",
		}, nil
	}

	err = p.projectRepo.AddPositions(ctx, bsonProjectID, newPositions)
	if err != nil {
		return nil, errutils.NewError(exceptions.ErrInternalError, errutils.InternalError).WithDebugMessage(err.Error())
	}

	return &responses.AddPositionsResponse{
		Message: "Position added successfully",
	}, nil
}

func (p *projectServiceImpl) ListPositions(ctx context.Context, req *requests.ListPositionsPathParams) ([]string, *errutils.Error) {
	bsonProjectID, err := bson.ObjectIDFromHex(req.ProjectID)
	if err != nil {
		return nil, errutils.NewError(exceptions.ErrInternalError, errutils.BadRequest).WithDebugMessage(err.Error())
	}

	positions, err := p.projectRepo.FindPositionByProjectID(ctx, bsonProjectID)
	if err != nil {
		return nil, errutils.NewError(exceptions.ErrInternalError, errutils.InternalError).WithDebugMessage(err.Error())
	}

	return positions, nil
}

func (p *projectServiceImpl) AddMembers(ctx context.Context, req *requests.AddProjectMembersRequest, userID string) (*responses.AddProjectMembersResponse, *errutils.Error) {
	bsonProjectID, err := bson.ObjectIDFromHex(req.ProjectID)
	if err != nil {
		return nil, errutils.NewError(exceptions.ErrInternalError, errutils.BadRequest).WithDebugMessage(err.Error())
	}

	bsonUserID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errutils.NewError(exceptions.ErrInternalError, errutils.BadRequest).WithDebugMessage(err.Error())
	}

	if len(req.Members) == 0 {
		return &responses.AddProjectMembersResponse{
			Message: "No member added",
		}, nil
	}

	project, err := p.projectRepo.FindByProjectID(ctx, bsonProjectID)
	if err != nil {
		return nil, errutils.NewError(exceptions.ErrInternalError, errutils.InternalError).WithDebugMessage(err.Error())
	} else if project == nil {
		return nil, errutils.NewError(exceptions.ErrProjectNotFound, errutils.NotFound)
	}

	// Check if the user is owner or moderator of the project
	member, err := p.projectRepo.FindMemberByProjectIDAndUserID(ctx, bsonProjectID, bsonUserID)
	if err != nil {
		return nil, errutils.NewError(exceptions.ErrInternalError, errutils.InternalError).WithDebugMessage(err.Error())
	} else if member == nil {
		return nil, errutils.NewError(exceptions.ErrUserNotFound, errutils.BadRequest)
	} else if member.Role != models.ProjectMemberRoleOwner && member.Role != models.ProjectMemberRoleModerator {
		return nil, errutils.NewError(exceptions.ErrPermissionDenied, errutils.BadRequest)
	}

	createProjMemberReq := make([]repositories.CreateProjectMemberRequest, 0)
	for _, member := range req.Members {
		bsonMemberID, err := bson.ObjectIDFromHex(member.UserID)
		if err != nil {
			return nil, errutils.NewError(exceptions.ErrInternalError, errutils.BadRequest).WithDebugMessage(err.Error())
		}

		// Check if the member already exists
		existingMember, err := p.projectRepo.FindMemberByProjectIDAndUserID(ctx, bsonProjectID, bsonMemberID)
		if err != nil {
			return nil, errutils.NewError(exceptions.ErrInternalError, errutils.InternalError).WithDebugMessage(err.Error())
		} else if existingMember != nil {
			continue
		}

		// Check if the member is a member of the workspace
		workspaceMember, err := p.workspaceRepo.FindWorkspaceMemberByWorkspaceIDAndUserID(ctx, project.WorkspaceID, bsonMemberID)
		if err != nil {
			return nil, errutils.NewError(exceptions.ErrInternalError, errutils.InternalError).WithDebugMessage(err.Error())
		} else if workspaceMember == nil {
			return nil, errutils.NewError(exceptions.ErrMemberNotFoundInWorkspace, errutils.BadRequest)
		}

		createProjMemberReq = append(createProjMemberReq, repositories.CreateProjectMemberRequest{
			UserID:      bsonMemberID,
			DisplayName: workspaceMember.DisplayName,
			ProfileUrl:  workspaceMember.ProfileUrl,
			Position:    member.Position,
			Role:        models.ProjectMemberRole(member.Role),
		})
	}

	err = p.projectRepo.AddMembers(ctx, bsonProjectID, createProjMemberReq)
	if err != nil {
		return nil, errutils.NewError(exceptions.ErrInternalError, errutils.InternalError).WithDebugMessage(err.Error())
	}

	return nil, nil
}

func validateListMembersPaginationRequestSortBy(sortBy string) bool {
	switch sortBy {
	case constant.ProjectMemberFieldDisplayName, constant.ProjectMemberFieldJoinedAt:
		return true
	}
	return false
}

func validateListMembersPaginationRequest(req *requests.ListProjectMembersRequest) {
	if req.PaginationRequest.Page <= 0 {
		req.PaginationRequest.Page = 1
	}
	if req.PaginationRequest.PageSize <= 0 {
		req.PaginationRequest.PageSize = 100
	}
	if req.PaginationRequest.SortBy == "" || !validateListMembersPaginationRequestSortBy(req.PaginationRequest.SortBy) {
		req.PaginationRequest.SortBy = constant.ProjectMemberFieldDisplayName
	}
	if req.PaginationRequest.Order == "" {
		req.PaginationRequest.Order = constant.ASC
	}
}

func (p *projectServiceImpl) ListMembers(ctx context.Context, req *requests.ListProjectMembersRequest) (*responses.ListProjectMembersResponse, *errutils.Error) {
	bsonProjectID, err := bson.ObjectIDFromHex(req.ProjectID)
	if err != nil {
		return nil, errutils.NewError(exceptions.ErrInternalError, errutils.BadRequest).WithDebugMessage(err.Error())
	}

	validateListMembersPaginationRequest(req)

	members, totalMember, err := p.projectRepo.SearchProjectMember(ctx, &repositories.SearchProjectMemberRequest{
		ProjectID: bsonProjectID,
		Keyword:   req.Keyword,
		PaginationRequest: repositories.PaginationRequest{
			Page:     req.PaginationRequest.Page,
			PageSize: req.PaginationRequest.PageSize,
			Order:    req.PaginationRequest.Order,
			SortBy:   req.PaginationRequest.SortBy,
		},
	})
	if err != nil {
		return nil, errutils.NewError(exceptions.ErrInternalError, errutils.InternalError).WithDebugMessage(err.Error())
	}

	return &responses.ListProjectMembersResponse{
		Members: members,
		PaginationResponse: &responses.PaginationResponse{
			Page:      req.PaginationRequest.Page,
			PageSize:  req.PaginationRequest.PageSize,
			TotalPage: int(math.Ceil(float64(totalMember) / float64(req.PaginationRequest.PageSize))),
			TotalItem: int(totalMember),
		},
	}, nil
}

func (p *projectServiceImpl) AddWorkflows(ctx context.Context, req *requests.AddWorkflowsRequest, userID string) (*responses.AddWorkflowsResponse, *errutils.Error) {
	bsonUserID, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errutils.NewError(exceptions.ErrInternalError, errutils.BadRequest).WithDebugMessage(err.Error())
	}

	bsonProjectID, err := bson.ObjectIDFromHex(req.ProjectID)
	if err != nil {
		return nil, errutils.NewError(exceptions.ErrInternalError, errutils.BadRequest).WithDebugMessage(err.Error())
	}

	project, err := p.projectRepo.FindByProjectID(ctx, bsonProjectID)
	if err != nil {
		return nil, errutils.NewError(exceptions.ErrInternalError, errutils.InternalError).WithDebugMessage(err.Error())
	} else if project == nil {
		return nil, errutils.NewError(exceptions.ErrProjectNotFound, errutils.NotFound)
	}

	// Check if the user is owner or moderator of the project
	member, err := p.projectRepo.FindMemberByProjectIDAndUserID(ctx, bsonProjectID, bsonUserID)
	if err != nil {
		return nil, errutils.NewError(exceptions.ErrInternalError, errutils.InternalError).WithDebugMessage(err.Error())
	} else if member == nil {
		return nil, errutils.NewError(exceptions.ErrPermissionDenied, errutils.BadRequest)
	} else if member.Role != models.ProjectMemberRoleOwner && member.Role != models.ProjectMemberRoleModerator {
		return nil, errutils.NewError(exceptions.ErrPermissionDenied, errutils.BadRequest)
	}

	// Check if the workflow already exists
	existingWorkflows, err := p.projectRepo.FindWorkflowByProjectID(ctx, bsonProjectID)
	if err != nil {
		return nil, errutils.NewError(exceptions.ErrInternalError, errutils.InternalError).WithDebugMessage(err.Error())
	}

	workflowMap := make(map[string]struct{})
	for _, workflow := range existingWorkflows {
		workflowMap[workflow.Status] = struct{}{}
	}

	var newWorkflows []models.Workflow
	for _, workflow := range req.Workflows {
		if _, ok := workflowMap[workflow.Status]; !ok {
			newWorkflows = append(newWorkflows, models.Workflow{
				Status:           workflow.Status,
				PreviousStatuses: workflow.PreviousStatuses,
			})
		}
	}

	if len(newWorkflows) == 0 {
		return &responses.AddWorkflowsResponse{
			Message: "No new workflow added",
		}, nil
	}

	err = p.projectRepo.AddWorkflows(ctx, bsonProjectID, newWorkflows)
	if err != nil {
		return nil, errutils.NewError(exceptions.ErrInternalError, errutils.InternalError).WithDebugMessage(err.Error())
	}

	return &responses.AddWorkflowsResponse{
		Message: "Workflow added successfully",
	}, nil
}

func (p *projectServiceImpl) ListWorkflows(ctx context.Context, req *requests.ListWorkflowsPathParams) ([]models.Workflow, *errutils.Error) {
	bsonProjectID, err := bson.ObjectIDFromHex(req.ProjectID)
	if err != nil {
		return nil, errutils.NewError(exceptions.ErrInternalError, errutils.BadRequest).WithDebugMessage(err.Error())
	}

	workflows, err := p.projectRepo.FindWorkflowByProjectID(ctx, bsonProjectID)
	if err != nil {
		return nil, errutils.NewError(exceptions.ErrInternalError, errutils.InternalError).WithDebugMessage(err.Error())
	}

	return workflows, nil
}

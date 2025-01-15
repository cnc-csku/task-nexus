package services

import (
	"context"
	"math"
	"time"

	"github.com/cnc-csku/task-nexus/go-lib/utils/errutils"
	"github.com/cnc-csku/task-nexus/task-management/config"
	"github.com/cnc-csku/task-nexus/task-management/domain/constant"
	"github.com/cnc-csku/task-nexus/task-management/domain/exceptions"
	"github.com/cnc-csku/task-nexus/task-management/domain/models"
	"github.com/cnc-csku/task-nexus/task-management/domain/repositories"
	"github.com/cnc-csku/task-nexus/task-management/domain/requests"
	"github.com/cnc-csku/task-nexus/task-management/domain/responses"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	Register(ctx context.Context, req *requests.RegisterRequest) (*responses.UserResponse, *errutils.Error)
	Login(ctx context.Context, req *requests.LoginRequest) (*responses.UserWithTokenResponse, *errutils.Error)
	FindUserByEmail(ctx context.Context, email string) (*responses.UserResponse, *errutils.Error)
	Search(ctx context.Context, req *requests.SearchUserRequest, searcherUserId string) (*responses.ListUserResponse, *errutils.Error)
}

type userServiceImpl struct {
	userRepo repositories.UserRepository
	config   *config.Config
}

func NewUserService(userRepo repositories.UserRepository, config *config.Config) UserService {
	return &userServiceImpl{
		userRepo: userRepo,
		config:   config,
	}
}

func (u *userServiceImpl) Register(ctx context.Context, req *requests.RegisterRequest) (*responses.UserResponse, *errutils.Error) {
	// Check if email already exists
	existsUser, err := u.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, errutils.NewError(exceptions.ErrInternalError, errutils.InternalError)
	}

	if existsUser != nil {
		return nil, errutils.NewError(exceptions.ErrUserAlreadyExists, errutils.BadRequest)
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errutils.NewError(err, errutils.InternalServerError)
	}

	req.Password = string(hashedPassword)

	user := &repositories.CreateUserRequest{
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		FullName:     req.FullName,
		DisplayName:  req.DisplayName,
	}

	createdUser, err := u.userRepo.Create(ctx, user)
	if err != nil {
		return nil, errutils.NewError(exceptions.ErrInternalError, errutils.InternalError)
	}

	res := &responses.UserResponse{
		ID:          createdUser.ID.Hex(),
		Email:       createdUser.Email,
		FullName:    createdUser.FullName,
		DisplayName: createdUser.DisplayName,
		CreatedAt:   createdUser.CreatedAt,
		UpdatedAt:   createdUser.UpdatedAt,
	}

	return res, nil
}

func (u *userServiceImpl) Login(ctx context.Context, req *requests.LoginRequest) (*responses.UserWithTokenResponse, *errutils.Error) {
	// Find user by email
	user, err := u.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, errutils.NewError(exceptions.ErrInternalError, errutils.InternalError)
	}

	if user == nil {
		return nil, errutils.NewError(exceptions.ErrInvalidCredentials, errutils.Unauthorized)
	}

	// Compare password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		return nil, errutils.NewError(exceptions.ErrInvalidCredentials, errutils.Unauthorized)
	}

	// Generate JWT token
	expireAt := time.Now().Add(time.Hour * 1)

	claims := models.UserCustomClaims{
		ID:          user.ID.Hex(),
		FullName:    user.FullName,
		DisplayName: user.DisplayName,
		Email:       user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID.Hex(),
			ExpiresAt: jwt.NewNumericDate(expireAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the secret
	tokenString, err := token.SignedString([]byte(u.config.JWT.AccessTokenSecret))
	if err != nil {
		return nil, errutils.NewError(exceptions.ErrInternalError, errutils.InternalError)
	}

	res := &responses.UserWithTokenResponse{
		UserResponse: responses.UserResponse{
			ID:          user.ID.Hex(),
			Email:       user.Email,
			FullName:    user.FullName,
			DisplayName: user.DisplayName,
			CreatedAt:   user.CreatedAt,
			UpdatedAt:   user.UpdatedAt,
		},
		Token:         tokenString,
		TokenExpireAt: expireAt,
	}
	return res, nil
}

func (u *userServiceImpl) FindUserByEmail(ctx context.Context, email string) (*responses.UserResponse, *errutils.Error) {
	user, err := u.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, errutils.NewError(exceptions.ErrInternalError, errutils.InternalError)
	}

	if user == nil {
		return nil, errutils.NewError(exceptions.ErrUserNotFound, errutils.NotFound)
	}

	res := &responses.UserResponse{
		ID:          user.ID.Hex(),
		Email:       user.Email,
		FullName:    user.FullName,
		DisplayName: user.DisplayName,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}

	return res, nil
}

func validateSearchUserPaginationRequestSortBy(sortBy string) bool {
	switch sortBy {
	case constant.UserField_EMAIL, constant.UserField_FULL_NAME, constant.UserField_DISPLAY_NAME:
		return true
	}
	return false
}

func (u *userServiceImpl) Search(ctx context.Context, req *requests.SearchUserRequest, searcherUserId string) (*responses.ListUserResponse, *errutils.Error) {
	if req.PaginationRequest != nil {
		if req.PaginationRequest.Page <= 0 {
			req.PaginationRequest.Page = 1
		}
		if req.PaginationRequest.PageSize <= 0 {
			req.PaginationRequest.PageSize = 100
		}
		if req.PaginationRequest.SortBy == "" || !validateSearchUserPaginationRequestSortBy(req.PaginationRequest.SortBy) {
			req.PaginationRequest.SortBy = constant.UserField_EMAIL
		}
		if req.PaginationRequest.Order == "" {
			req.PaginationRequest.Order = constant.ASC
		}
	} else {
		req.PaginationRequest = &requests.PaginationRequest{
			Page:     1,
			PageSize: 100,
			SortBy:   constant.UserField_EMAIL,
			Order:    constant.ASC,
		}
	}

	users, totalUser, err := u.userRepo.Search(ctx, &repositories.SearchUserRequest{
		Keyword:           req.Keyword,
		PaginationRequest: repositories.PaginationRequest{Page: req.PaginationRequest.Page, PageSize: req.PaginationRequest.PageSize, SortBy: req.PaginationRequest.SortBy, Order: req.PaginationRequest.Order},
	})
	if err != nil {
		return nil, errutils.NewError(exceptions.ErrInternalError, errutils.InternalError).WithDebugMessage(err.Error())
	}

	res := &responses.ListUserResponse{
		Users: make([]responses.UserResponse, 0),
		PaginationResponse: responses.PaginationResponse{
			Page:      req.PaginationRequest.Page,
			PageSize:  req.PaginationRequest.PageSize,
			TotalPage: int(math.Ceil(float64(totalUser) / float64(req.PaginationRequest.PageSize))),
			TotalItem: int(totalUser),
		},
	}

	for _, user := range users {
		if user.ID.Hex() == searcherUserId {
			res.PaginationResponse.TotalItem--
			continue
		}

		res.Users = append(res.Users, responses.UserResponse{
			ID:          user.ID.Hex(),
			Email:       user.Email,
			FullName:    user.FullName,
			DisplayName: user.DisplayName,
			CreatedAt:   user.CreatedAt,
			UpdatedAt:   user.UpdatedAt,
		})
	}

	return res, nil
}
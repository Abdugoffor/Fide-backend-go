package auth_service

import (
	"context"
	"errors"
	"strconv"

	"auth-login/helper"
	auth_dto "auth-login/module/auth/dto"
	auth_repository "auth-login/module/auth/repository"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserNotActive      = errors.New("user is not active")
	ErrInvalidToken       = errors.New("invalid or expired token")
)

type AuthService struct {
	repository auth_repository.AuthRepository
	secret     string
}

func NewAuthService(db *pgxpool.Pool, secret string) *AuthService {
	return &AuthService{
		repository: auth_repository.NewAuthRepository(db),
		secret:     secret,
	}
}

func (service *AuthService) SeedSuperadmin(ctx context.Context) error {
	passwordHash, err := helper.HashPassword("123456798")
	{
		if err != nil {
			return err
		}
	}

	return service.repository.SeedSuperadmin(ctx, passwordHash)
}

func (service *AuthService) Login(ctx context.Context, req *auth_dto.Login) (*auth_dto.LoginResponse, error) {
	user, err := service.repository.FindByEmail(ctx, req.Email)
	{
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrInvalidCredentials
		}

		if err != nil {
			return nil, err
		}
	}

	err = helper.CheckPassword(user.PasswordHash, req.Password)
	{
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) || errors.Is(err, bcrypt.ErrPasswordTooLong) {
			return nil, ErrInvalidCredentials
		}

		if err != nil {
			return nil, err
		}
	}

	if user.Status != auth_dto.StatusActive {
		return nil, ErrUserNotActive
	}

	token, expiresAt, err := helper.NewToken(user.ID, user.Role, service.secret)
	{
		if err != nil {
			return nil, err
		}
	}

	err = service.repository.CreateSession(ctx, user.ID, helper.HashToken(token), expiresAt)
	{
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotActive
		}

		if err != nil {
			return nil, err
		}
	}

	return &auth_dto.LoginResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresAt:   expiresAt,
		User:        user.Response,
	}, nil
}

func (service *AuthService) Authenticate(ctx context.Context, token string) (*auth_dto.Response, error) {
	userID, err := service.tokenUserID(token)
	{
		if err != nil {
			return nil, err
		}
	}

	user, err := service.repository.FindBySession(ctx, userID, helper.HashToken(token))
	{
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrInvalidToken
		}

		if err != nil {
			return nil, err
		}
	}

	if user.Status != auth_dto.StatusActive {
		return nil, ErrUserNotActive
	}

	return user, nil
}

func (service *AuthService) Logout(ctx context.Context, token string) error {
	userID, err := service.tokenUserID(token)
	{
		if err != nil {
			return err
		}
	}

	err = service.repository.DeleteSession(ctx, userID, helper.HashToken(token))
	{
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrInvalidToken
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func (service *AuthService) tokenUserID(token string) (int64, error) {
	if len(token) > 4096 {
		return 0, ErrInvalidToken
	}

	claims, err := helper.ParseToken(token, service.secret)
	{
		if err != nil {
			return 0, ErrInvalidToken
		}
	}

	userID, err := strconv.ParseInt(claims.Subject, 10, 64)
	{
		if err != nil || userID < 1 || claims.ID == "" {
			return 0, ErrInvalidToken
		}
	}

	return userID, nil
}

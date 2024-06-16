package users

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/olad5/caution-companion/internal/domain"
	"github.com/olad5/caution-companion/internal/infra"
	"github.com/olad5/caution-companion/internal/services/auth"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	userRepo    infra.UserRepository
	authService auth.AuthService
}

var (
	ErrUserAlreadyExists = errors.New("email already exist")
	ErrPasswordIncorrect = errors.New("invalid credentials")
	ErrInvalidToken      = errors.New("invalid token")
)

func NewUserService(userRepo infra.UserRepository, authService auth.AuthService) (*UserService, error) {
	if userRepo == nil {
		return &UserService{}, errors.New("UserService failed to initialize, userRepo is nil")
	}
	if authService == nil {
		return &UserService{}, errors.New("UserService failed to initialize, authService is nil")
	}
	return &UserService{userRepo, authService}, nil
}

func (u *UserService) CreateUser(ctx context.Context, firstName, lastName, email, password string) (domain.User, error) {
	existingUser, err := u.userRepo.GetUserByEmail(ctx, email)
	if err == nil && existingUser.Email == email {
		return domain.User{}, ErrUserAlreadyExists
	}

	hashedPassword, err := hashAndSalt([]byte(password))
	if err != nil {
		return domain.User{}, err
	}

	newUser := domain.User{
		ID:        uuid.New(),
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		Password:  hashedPassword,
	}

	err = u.userRepo.CreateUser(ctx, newUser)
	if err != nil {
		return domain.User{}, err
	}
	return newUser, nil
}

func (u *UserService) LogUserIn(ctx context.Context, email, password string) (string, string, error) {
	existingUser, err := u.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return "", "", err
	}

	if isPasswordCorrect := comparePasswords(existingUser.Password, []byte(password)); !isPasswordCorrect {
		return "", "", ErrPasswordIncorrect
	}

	accessToken, refreshToken, err := u.authService.GenerateAuthTokens(ctx, existingUser)
	if err != nil {
		return "", "", err
	}
	return accessToken, refreshToken, nil
}

func (u *UserService) GetLoggedInUser(ctx context.Context) (domain.User, error) {
	jwtClaims, ok := auth.GetJWTClaims(ctx)
	if !ok {
		return domain.User{}, fmt.Errorf("error parsing JWTClaims: %v", ErrInvalidToken)
	}
	userId := jwtClaims.ID

	existingUser, err := u.userRepo.GetUserByUserId(ctx, userId)
	if err != nil {
		return domain.User{}, err
	}
	return existingUser, nil
}

func (u *UserService) RefreshUserAccessToken(ctx context.Context, existingRefreshToken string) (string, string, error) {
	userId, err := u.authService.GetUserIdFromRefreshToken(ctx, existingRefreshToken)
	if err != nil {
		return "", "", err
	}

	existingUser, err := u.userRepo.GetUserByUserId(ctx, userId)
	if err != nil {
		return "", "", err
	}

	accessToken, refreshToken, err := u.authService.GenerateAuthTokens(ctx, existingUser)
	if err != nil {
		return "", "", err
	}
	return accessToken, refreshToken, nil
}

func hashAndSalt(plainPassword []byte) (string, error) {
	hash, err := bcrypt.GenerateFromPassword(plainPassword, bcrypt.MinCost)
	if err != nil {
		return "", errors.New("error hashing password")
	}
	return string(hash), nil
}

func comparePasswords(hashedPassword string, plainPassword []byte) bool {
	byteHash := []byte(hashedPassword)
	err := bcrypt.CompareHashAndPassword(byteHash, plainPassword)
	return err == nil
}

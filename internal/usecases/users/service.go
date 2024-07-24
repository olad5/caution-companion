package users

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/olad5/caution-companion/internal/domain"
	"github.com/olad5/caution-companion/internal/infra"
	"github.com/olad5/caution-companion/internal/services/auth"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	userRepo    infra.UserRepository
	authService auth.AuthService
	mailService infra.MailService
}

var (
	ErrUserAlreadyExists     = errors.New("email already exist")
	ErrUserNameAlreadyExists = errors.New("user_name already exist")
	ErrPasswordIncorrect     = errors.New("invalid credentials")
	ErrInvalidToken          = errors.New("invalid token")
)

const DEFAULT_AVATAR = "https://res.cloudinary.com/deda4nfxl/image/upload/v1721583338/caution-companion/caution-companion/avatars/4608bc1b98c84a06838fafb5e38fb552.jpg"

func NewUserService(
	userRepo infra.UserRepository,
	authService auth.AuthService,
	mailService infra.MailService,
) (*UserService, error) {
	if userRepo == nil {
		return &UserService{}, errors.New("UserService failed to initialize, userRepo is nil")
	}
	if authService == nil {
		return &UserService{}, errors.New("UserService failed to initialize, authService is nil")
	}
	if mailService == nil {
		return &UserService{}, errors.New("UserService failed to initialize, mailService is nil")
	}
	return &UserService{userRepo, authService, mailService}, nil
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
		Email:     strings.ToLower(email),
		AvatarUrl: DEFAULT_AVATAR,
		FirstName: strings.ToLower(firstName),
		LastName:  strings.ToLower(lastName),
		UserName:  createDefaultUserName(firstName, lastName),
		Password:  hashedPassword,
	}

	err = u.userRepo.CreateUser(ctx, newUser)
	if err != nil {
		return domain.User{}, err
	}
	return newUser, nil
}

func (u *UserService) EditUser(ctx context.Context, firstName, lastName, userName, email, avatarUrl, location, phone string) (domain.User, error) {
	jwtClaims, ok := auth.GetJWTClaims(ctx)
	if !ok {
		return domain.User{}, fmt.Errorf("error parsing JWTClaims: %v", ErrInvalidToken)
	}
	userId := jwtClaims.ID

	existingUser, err := u.userRepo.GetUserByUserId(ctx, userId)
	if err != nil {
		return domain.User{}, err
	}

	found, err := u.userRepo.GetUserByUserName(ctx, userName)
	if err == nil && found.UserName == userName {
		return domain.User{}, ErrUserNameAlreadyExists
	}

	if existingUser.Email != email {
		err = u.authService.LogUserOut(ctx, userId.String())
		if err != nil {
			return domain.User{}, fmt.Errorf("Error deleting existing JWTClaims: %v", err)
		}
	}

	updatedUser := domain.User{
		ID:        existingUser.ID,
		Email:     email,
		AvatarUrl: avatarUrl,
		FirstName: strings.ToLower(firstName),
		LastName:  strings.ToLower(lastName),
		UserName:  userName,
		Password:  existingUser.Password,
		Location:  location,
		Phone:     phone,
		CreatedAt: existingUser.CreatedAt,
		UpdatedAt: time.Now(),
	}

	err = u.userRepo.UpdateUser(ctx, updatedUser)
	if err != nil {
		return domain.User{}, err
	}
	return updatedUser, nil
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

	// for fixing user_name migration
	if existingUser.UserName == "" {
		existingUser.UserName = createDefaultUserName(existingUser.FirstName, existingUser.LastName)
		existingUser.UpdatedAt = time.Now()
		err = u.userRepo.UpdateUser(ctx, existingUser)
		if err != nil {
			return domain.User{}, err
		}
	}
	return existingUser, nil
}

func (u *UserService) LogUserOut(ctx context.Context) error {
	// TODO:TODO: this jwt check is duplicated multiple times, clean it up
	jwtClaims, ok := auth.GetJWTClaims(ctx)
	if !ok {
		return fmt.Errorf("error parsing JWTClaims: %v", ErrInvalidToken)
	}
	userId := jwtClaims.ID

	return u.authService.LogUserOut(ctx, userId.String())
}

func (u *UserService) ForgotPassword(ctx context.Context, email string) error {
	existingUser, err := u.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return err
	}

	otp := getRandomIntString(6)
	opts := infra.MailOptions{
		To:      existingUser.Email,
		Subject: "forgot password",
		Body:    "otp, expires in 10 min:  " + otp,
	}
	err = u.authService.AddPasswordResetTokenToCache(ctx, existingUser.ID, otp)
	if err != nil {
		// TODO:TODO: i need to log stuff
		return err
	}
	err = u.mailService.Send(ctx, opts)
	if err != nil {
		return err
	}

	return nil
}

func (u *UserService) ChangePassword(ctx context.Context, oldPassword, newPassword string) error {
	jwtClaims, ok := auth.GetJWTClaims(ctx)
	if !ok {
		// TODO:TODO: format the erros well, let Error sto start with uppercase
		return fmt.Errorf("error parsing JWTClaims: %v", ErrInvalidToken)
	}
	userId := jwtClaims.ID

	existingUser, err := u.userRepo.GetUserByUserId(ctx, userId)
	if err != nil {
		return err
	}

	if isOldPasswordCorrect := comparePasswords(existingUser.Password, []byte(oldPassword)); !isOldPasswordCorrect {
		return ErrPasswordIncorrect
	}
	hashedPassword, err := hashAndSalt([]byte(newPassword))
	if err != nil {
		return err
	}

	existingUser.Password = hashedPassword
	existingUser.UpdatedAt = time.Now()
	err = u.userRepo.UpdateUser(ctx, existingUser)
	if err != nil {
		return err
	}

	err = u.authService.LogUserOut(ctx, existingUser.ID.String())
	if err != nil {
		// TODO:TODO:  // log err,
		return fmt.Errorf("Error deleting existing JWTClaims: %v", err)
	}

	return nil
}

func (u *UserService) ResetPassword(ctx context.Context, token, newPassword string) error {
	id, err := u.authService.GetUserIdFromPasswordResetToken(ctx, token)
	if err != nil {
		return err
	}

	userId, err := uuid.Parse(id)
	if err != nil {
		return ErrInvalidToken
	}

	existingUser, err := u.userRepo.GetUserByUserId(ctx, userId)
	if err != nil {
		return err
	}

	hashedPassword, err := hashAndSalt([]byte(newPassword))
	if err != nil {
		return err
	}

	existingUser.Password = hashedPassword
	existingUser.UpdatedAt = time.Now()
	err = u.userRepo.UpdateUser(ctx, existingUser)
	if err != nil {
		return err
	}

	err = u.authService.LogUserOut(ctx, existingUser.ID.String())
	if err == nil {
		// TODO:TODO:  // log err, not important to return
		// err
	}

	return u.authService.DeletePasswordResetToken(ctx, token)
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

func createDefaultUserName(firstName, lastName string) string {
	result := ""
	const maxCharsForName = 4

	if len(firstName) > maxCharsForName {
		result = result + firstName[:maxCharsForName]
	} else {
		result = result + firstName
	}
	if len(lastName) > maxCharsForName {
		result = result + lastName[:maxCharsForName]
	} else {
		result = result + lastName
	}

	const maxUserNameChars = 11
	toComplete := maxUserNameChars - len(result)
	if toComplete > 0 {
		result = result + getRandomIntString(toComplete)
	}
	return result
}

func getRandomIntString(length int) string {
	MAX_INT := 7935425686241
	b := new(big.Int).SetInt64(int64(MAX_INT))
	randomBigInt, _ := rand.Int(rand.Reader, b)
	randomeNewInt := int(randomBigInt.Int64())
	s := fmt.Sprint(randomeNewInt)
	return s[:length]
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

package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/olad5/caution-companion/internal/domain"
	"github.com/olad5/caution-companion/internal/infra"
	"go.uber.org/zap"
)

type PostgresUserRepository struct {
	logger     *zap.Logger
	connection *sqlx.DB
}

func NewPostgresUserRepo(ctx context.Context, logger *zap.Logger, connection *sqlx.DB) (*PostgresUserRepository, error) {
	if connection == nil {
		return &PostgresUserRepository{}, fmt.Errorf("Failed to create PostgresFileRepository: connection is nil")
	}

	return &PostgresUserRepository{logger: logger, connection: connection}, nil
}

func (p *PostgresUserRepository) CreateUser(ctx context.Context, user domain.User) error {
	const query = `
    INSERT INTO users
      (id, first_name, last_name, user_name, email, password, avatar_url, location, phone, created_at, updated_at) 
    VALUES 
    (:id, :first_name, :last_name, :user_name, :email, :password, :avatar_url, :location, :phone, :created_at, :updated_at)
  `

	_, err := p.connection.NamedExec(query, toSqlxUser(user))
	if err != nil {
		p.logger.Error("Error creating user in the db:  ", zap.Error(err))
		return fmt.Errorf("error creating user in the db: %w", err)
	}
	return nil
}

func (p *PostgresUserRepository) UpdateUser(ctx context.Context, user domain.User) error {
	const query = `
  UPDATE  
    users 
	SET
		"first_name" = :first_name,
		"last_name" = :last_name,
		"email" = :email,
		"user_name" = :user_name,
		"avatar_url" = :avatar_url,
		"password" = :password,
		"location" = :location,
		"phone" = :phone,
		"created_at" = :created_at,
		"updated_at" = :updated_at
  WHERE 
      id=:id
  `

	_, err := p.connection.NamedExec(query, toSqlxUser(user))
	if err != nil {
		p.logger.Error("Error updating user in the db:  ", zap.Error(err))
		return fmt.Errorf("error updating user in the db: %w", err)
	}
	return nil
}

func (p *PostgresUserRepository) GetUserByUserName(ctx context.Context, userName string) (domain.User, error) {
	var user SqlxUser

	err := p.connection.Get(&user, "SELECT * FROM users WHERE user_name = $1", userName)
	if err != nil {
		if errors.Is(err, ErrRecordNotFound) {
			return domain.User{}, infra.ErrUserNotFound
		}
		p.logger.Error("Error getting user by userName:  ", zap.Error(err))
		return domain.User{}, fmt.Errorf("error getting user by userName: %w", err)
	}
	return toUser(user), nil
}

func (p *PostgresUserRepository) GetUserByEmail(ctx context.Context, userEmail string) (domain.User, error) {
	var user SqlxUser

	err := p.connection.Get(&user, "SELECT * FROM users WHERE email = $1", userEmail)
	if err != nil {
		if errors.Is(err, ErrRecordNotFound) {
			return domain.User{}, infra.ErrUserNotFound
		}
		p.logger.Error("Error getting user by email:  ", zap.Error(err))
		return domain.User{}, fmt.Errorf("error getting user by email: %w", err)
	}
	return toUser(user), nil
}

func (p *PostgresUserRepository) GetUserByUserId(ctx context.Context, userId uuid.UUID) (domain.User, error) {
	var user SqlxUser

	err := p.connection.Get(&user, "SELECT * FROM users WHERE id = $1", userId)
	if err != nil {
		if errors.Is(err, ErrRecordNotFound) {
			return domain.User{}, infra.ErrUserNotFound
		}
		p.logger.Error("Error getting user by userId:  ", zap.Error(err))
		return domain.User{}, fmt.Errorf("error getting user by userId: %w", err)
	}
	return toUser(user), nil
}

func (p *PostgresUserRepository) Ping(ctx context.Context) error {
	if err := p.connection.Ping(); err != nil {
		p.logger.Error("Error pinging PostgresUserRepository:  ", zap.Error(err))
		return fmt.Errorf("failed to ping postgres database: %w", err)
	}
	return nil
}

type SqlxUser struct {
	ID        uuid.UUID `db:"id"`
	Email     string    `db:"email"`
	AvatarUrl string    `db:"avatar_url"`
	FirstName string    `db:"first_name"`
	LastName  string    `db:"last_name"`
	UserName  string    `db:"user_name"`
	Password  string    `db:"password"`
	Location  string    `db:"location"`
	Phone     string    `db:"phone"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func toUser(u SqlxUser) domain.User {
	return domain.User{
		ID:        u.ID,
		AvatarUrl: u.AvatarUrl,
		Email:     u.Email,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		UserName:  u.UserName,
		Password:  u.Password,
		Location:  u.Location,
		Phone:     u.Phone,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

func toSqlxUser(u domain.User) SqlxUser {
	return SqlxUser{
		ID:        u.ID,
		AvatarUrl: u.AvatarUrl,
		Email:     u.Email,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		UserName:  u.UserName,
		Password:  u.Password,
		Location:  u.Location,
		Phone:     u.Phone,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

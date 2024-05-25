package mongo

import (
	"context"
	"fmt"
	"time"

	"github.com/olad5/go-hackathon-starter-template/config"
	"github.com/olad5/go-hackathon-starter-template/internal/domain"
	"github.com/olad5/go-hackathon-starter-template/internal/infra"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var contextTimeoutDuration = 5 * time.Second

type MongoUserRepository struct {
	users *mongo.Collection
}

func NewMongoUserRepo(ctx context.Context, config *config.Configurations) (*MongoUserRepository, error) {
	opts := options.Client()
	client, err := mongo.Connect(ctx, opts.ApplyURI(config.DatabaseUrl))
	if err != nil {
		return nil, fmt.Errorf("failed to create a mongo client: %w", err)
	}

	userCollection := client.Database(config.DatabaseName).Collection("users")

	return &MongoUserRepository{users: userCollection}, nil
}

func (m *MongoUserRepository) CreateUser(ctx context.Context, user domain.User) error {
	ctx, cancel := context.WithTimeout(ctx, contextTimeoutDuration)
	defer cancel()

	mongoUser, err := toMongoUser(user)
	if err != nil {
		return fmt.Errorf("failed to map domain user to MongoUser: %w", err)
	}

	_, err = m.users.InsertOne(ctx, mongoUser)
	if err != nil {
		return fmt.Errorf("failed to persist todo: %w", err)
	}
	return nil
}

func (m *MongoUserRepository) GetUserByEmail(ctx context.Context, userEmail string) (domain.User, error) {
	user := mongoUser{}
	err := m.users.FindOne(ctx, bson.M{"email": userEmail}).Decode(&user)
	if err != nil {
		return domain.User{}, infra.ErrUserNotFound
	}
	return toDomainUser(user), nil
}

func (m *MongoUserRepository) GetUserByUserId(ctx context.Context, userId primitive.ObjectID) (domain.User, error) {
	user := mongoUser{}
	err := m.users.FindOne(ctx, bson.M{"_id": userId}).Decode(&user)
	if err != nil {
		return domain.User{}, infra.ErrUserNotFound
	}

	return toDomainUser(user), nil
}

func (m *MongoUserRepository) Ping(ctx context.Context) error {
	if err := m.users.Database().Client().Ping(ctx, readpref.Primary()); err != nil {
		return fmt.Errorf("failed to ping mongo database: %w", err)
	}

	return nil
}

type mongoUser struct {
	// I think to convert from ObjectID to uuid, I will add some default digits to
	// start the id sequence before ending it with the actual  ObjectID
	ObjectID  primitive.ObjectID `bson:"_id"`
	Email     string             `bson:"email"`
	FirstName string             `bson:"first_name"`
	LastName  string             `bson:"last_name"`
	Password  string             `bson:"password"`
	CreatedAt time.Time          `bson:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at"`
}

func toMongoUser(user domain.User) (mongoUser, error) {
	return mongoUser{
		ObjectID:  user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Password:  user.Password,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

func toDomainUser(m mongoUser) domain.User {
	return domain.User{
		ID:        m.ObjectID,
		Email:     m.Email,
		FirstName: m.FirstName,
		LastName:  m.LastName,
		Password:  m.Password,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

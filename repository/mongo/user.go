package mongo

import (
	"context"

	"github.com/truongnqse05461/chat-app/models"
	"github.com/truongnqse05461/chat-app/repository"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	collection = "users"
)

type userRepository struct {
	db *mongo.Database
}

// GetByUsername implements repository.User
func (u *userRepository) GetByUsername(ctx context.Context, username string) (models.User, error) {
	col := u.db.Collection(collection)

	var user models.User
	if err := col.FindOne(ctx, bson.M{"username": username}).Decode(&user); err != nil {
		return models.User{}, err
	}
	return user, nil
}

// Delete implements repository.User
func (u *userRepository) Delete(ctx context.Context, id string) error {
	panic("unimplemented")
}

// Get implements repository.User
func (u *userRepository) Get(ctx context.Context) ([]models.User, error) {
	panic("unimplemented")
}

// GetById implements repository.User
func (u *userRepository) GetById(ctx context.Context, id string) (models.User, error) {
	col := u.db.Collection(collection)

	var user models.User
	if err := col.FindOne(ctx, bson.M{"id": id}).Decode(&user); err != nil {
		return models.User{}, err
	}
	return user, nil
}

// Save implements repository.User
func (u *userRepository) Save(ctx context.Context, users ...models.User) error {
	col := u.db.Collection(collection)

	if _, err := col.InsertOne(ctx, users); err != nil {
		return err
	}
	return nil
}

func New(db *mongo.Database) repository.User {
	return &userRepository{db: db}
}

var _ repository.User = (*userRepository)(nil)

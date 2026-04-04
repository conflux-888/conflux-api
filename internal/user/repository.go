package user

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var ErrNotFound = errors.New("user not found")

type Repository struct {
	col *mongo.Collection
}

func NewRepository(db *mongo.Database) *Repository {
	col := db.Collection("users")

	col.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true),
	})

	return &Repository{col: col}
}

func (r *Repository) Create(ctx context.Context, u *User) error {
	now := time.Now()
	u.CreatedAt = now
	u.UpdatedAt = now
	_, err := r.col.InsertOne(ctx, u)
	return err
}

func (r *Repository) FindByEmail(ctx context.Context, email string) (*User, error) {
	var u User
	err := r.col.FindOne(ctx, bson.M{"email": email}).Decode(&u)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrNotFound
	}
	return &u, err
}

func (r *Repository) FindByID(ctx context.Context, id bson.ObjectID) (*User, error) {
	var u User
	err := r.col.FindOne(ctx, bson.M{"_id": id}).Decode(&u)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrNotFound
	}
	return &u, err
}

func (r *Repository) Update(ctx context.Context, id bson.ObjectID, update bson.M) (*User, error) {
	update["updated_at"] = time.Now()

	_, err := r.col.UpdateByID(ctx, id, bson.M{"$set": update})
	if err != nil {
		return nil, err
	}
	return r.FindByID(ctx, id)
}

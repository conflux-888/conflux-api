package sync

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const syncStateID = "gdelt"

type StateRepository struct {
	col *mongo.Collection
}

func NewStateRepository(db *mongo.Database) *StateRepository {
	return &StateRepository{col: db.Collection("sync_state")}
}

func (r *StateRepository) Get(ctx context.Context) (*SyncState, error) {
	var state SyncState
	err := r.col.FindOne(ctx, bson.M{"_id": syncStateID}).Decode(&state)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return &SyncState{ID: syncStateID}, nil
	}
	if err != nil {
		return nil, err
	}
	return &state, nil
}

func (r *StateRepository) Upsert(ctx context.Context, state *SyncState) error {
	state.ID = syncStateID
	_, err := r.col.UpdateOne(
		ctx,
		bson.M{"_id": syncStateID},
		bson.M{"$set": state},
		options.UpdateOne().SetUpsert(true),
	)
	return err
}

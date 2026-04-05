package database

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func Connect(ctx context.Context, uri, dbName string) (*mongo.Database, error) {
	log.Info().Msg("[database.Connect] connecting to mongodb ...")

	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		log.Err(err).Msg("[database.Connect] Failed to connect to mongodb")
		return nil, fmt.Errorf("failed to connect to mongodb: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		log.Err(err).Msg("failed to ping mongodb")
		return nil, fmt.Errorf("failed to ping mongodb: %w", err)
	}

	log.Info().Str("database", dbName).Msg("[database.Connect] connected to mongodb")
	return client.Database(dbName), nil
}

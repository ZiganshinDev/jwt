package mongodb

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func NewClient(uri, username, password string) (*mongo.Client, error) {
	const op = "storage.mongodb.NewClient"

	opts := options.Client().ApplyURI(uri)
	if username != "" && password != "" {
		opts.SetAuth(options.Credential{
			Username: username, Password: password,
		})
	}

	client, err := mongo.Connect(context.TODO(), opts)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Проверьте подключение
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return client, nil
}

func IsDuplicate(err error) bool {
	var e mongo.WriteException
	if errors.As(err, &e) {
		for _, we := range e.WriteErrors {
			if we.Code == 11000 {
				return true
			}
		}
	}

	return false
}

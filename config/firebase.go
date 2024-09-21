package config

import (
	"context"
	"fmt"
	"log"
	"path/filepath"

	firebase "firebase.google.com/go/v4"
	"cloud.google.com/go/firestore"
	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

var FirestoreClient *firestore.Client
var StorageClient *storage.Client

func InitializeFirestore() error {
	ctx := context.Background()

	serviceAccountPath := "serviceAccountKey.json"
	absolutePath, err := filepath.Abs(serviceAccountPath)
	if err != nil {
		return fmt.Errorf("error getting absolute path: %v", err)
	}

	// fmt.Printf("Using service account key file at: %s\n", absolutePath)

	opt := option.WithCredentialsFile(absolutePath)

	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return fmt.Errorf("error initializing Firebase app: %v", err)
	}

	client, err := app.Firestore(ctx)
	if err != nil {
		return fmt.Errorf("error getting Firestore client: %v", err)
	}

	FirestoreClient = client
	return nil
}

func InitializeStorage() error {
	ctx := context.Background()

	var err error
	StorageClient, err = storage.NewClient(ctx, option.WithCredentialsFile("serviceAccountKey.json"))
	if err != nil {
		return fmt.Errorf("error initializing Cloud Storage client: %v", err)
	}

	return nil
}

func init() {
	if err := InitializeFirestore(); err != nil {
		log.Fatalf("Failed to initialize Firestore in init(): %v", err)
	}
	if err := InitializeStorage(); err != nil {
		log.Fatalf("Failed to initialize Cloud Storage in init(): %v", err)
	}
}

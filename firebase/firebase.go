package firebase

import (
	"context"
	"os"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"
)

var FirebaseApp *firebase.App
var FirebaseClient *auth.Client
var ctx context.Context = context.Background()

func InitFirebaseAdminSdk() (*auth.Client, error) {
	configBytes := []byte(os.Getenv("CONFIG"))
	opt := option.WithCredentialsJSON(configBytes)

	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return nil, err
	}

	client, err := app.Auth(ctx)
	if err != nil {
		return nil, err
	}

	FirebaseApp = app
	FirebaseClient = client
	return client, nil
}

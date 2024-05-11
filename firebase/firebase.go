package firebase

import (
	"context"
	"os"

	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/option"
)

var FirebaseApp *firebase.App

func InitFirebaseAdminSdk() error {
	configBytes := []byte(os.Getenv("CONFIG"))
	opt := option.WithCredentialsJSON(configBytes)

	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		return err
	}

	FirebaseApp = app
	return nil
}

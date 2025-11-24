package utils

import (
	"context"
	"log"

	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
)

func LoadFirebaseConfig() (*firebase.App,error ){
	opt := option.WithCredentialsFile("./credentials.json")
	app, err := firebase.NewApp(context.Background(), nil, opt)

	if err != nil {
		log.Fatalf("error initializing app: %v\n", err)
		return nil, err	
	}
	return app, nil
}
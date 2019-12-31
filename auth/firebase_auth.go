package auth

import (
	"context"
	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"fmt"
)

type RegisterFirebaseAuthInfo struct {
	Email         string
	EmailVerified bool
	PhoneNumber   string
	Password      string
	DisplayName   string
	PhotoURL      string
	Disabled      bool
}

type FirebaseAuthService struct {
	app *firebase.App
}

func NewFirebaseAuthService() *FirebaseAuthService {
	// Get firebase.App
	app, err := firebase.NewApp(context.Background(), nil)
	if err != nil {
		return nil
	}

	return &FirebaseAuthService{
		app: app,
	}
}

func (s *FirebaseAuthService) Register(ctx context.Context, info RegisterFirebaseAuthInfo) (*auth.UserRecord, error) {
	// Get an auth client from the firebase.App
	client, err := s.app.Auth(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting Auth client: %w", err)
	}

	return createUser(ctx, client, info)
}

func createUser(ctx context.Context, client *auth.Client, info RegisterFirebaseAuthInfo) (*auth.UserRecord, error) {
	params := (&auth.UserToCreate{}).
		Email(info.Email).
		EmailVerified(info.EmailVerified).
		PhoneNumber(info.PhoneNumber).
		Password(info.Password).
		DisplayName(info.DisplayName).
		PhotoURL(info.PhotoURL).
		Disabled(info.Disabled)
	u, err := client.CreateUser(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("error creating user: %w", err)
	}
	return u, nil
}
